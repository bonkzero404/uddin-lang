package interpreter

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
)

// Database connection pool and listener management
var (
	databaseConnections = make(map[string]*sql.DB)
	databaseStreamers   = make(map[string]*DatabaseStreamer)
	databaseMutex       = sync.RWMutex{}
)

// BinlogStreamer represents a MySQL binlog streamer for CDC
type BinlogStreamer struct {
	Syncer   *replication.BinlogSyncer
	Streamer *replication.BinlogStreamer
	Config   replication.BinlogSyncerConfig
	Position mysql.Position
	Tables   map[string]bool // Tables to monitor
}

// DatabaseStreamer represents a real-time database streamer using LISTEN/NOTIFY or Binlog
type DatabaseStreamer struct {
	ID        string
	DB        *sql.DB
	Listener  *pq.Listener
	Binlog    *BinlogStreamer
	TableName string
	Callback  Value // UDDIN-LANG function to call on data change
	Active    bool
	Cancel    context.CancelFunc
	Mutex     sync.RWMutex
	Driver    string
	LastID    int64
	BinlogPos string // For MySQL binary log position
	UseBinlog bool   // Whether to use binlog streaming instead of polling
	ServerID  uint32 // MySQL server ID for binlog replication
}

// DatabaseConnection represents a database connection object
type DatabaseConnection struct {
	ID       string
	DB       *sql.DB
	Driver   string
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

// dbConnectFunc implements the db_connect() built-in function
// db_connect(driver, host, port, database, username, password) -> connection_object
// Example: conn = db_connect("postgres", "localhost", 5432, "mydb", "user", "pass")
func dbConnectFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 6 {
		panic(typeError(pos, "db_connect() requires 6 arguments: driver, host, port, database, username, password"))
	}

	driver, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "db_connect() requires first argument to be a string (driver)"))
	}

	host, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "db_connect() requires second argument to be a string (host)"))
	}

	port, ok := args[2].(int)
	if !ok {
		panic(typeError(pos, "db_connect() requires third argument to be an integer (port)"))
	}

	database, ok := args[3].(string)
	if !ok {
		panic(typeError(pos, "db_connect() requires fourth argument to be a string (database)"))
	}

	username, ok := args[4].(string)
	if !ok {
		panic(typeError(pos, "db_connect() requires fifth argument to be a string (username)"))
	}

	password, ok := args[5].(string)
	if !ok {
		panic(typeError(pos, "db_connect() requires sixth argument to be a string (password)"))
	}

	// Validate driver
	if driver != "postgres" && driver != "mysql" {
		panic(typeError(pos, "db_connect() supports only 'postgres' and 'mysql' drivers"))
	}

	// Build connection string
	var dsn string
	switch driver {
	case "postgres":
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, username, password, database)
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			username, password, host, port, database)
	}

	// Open database connection
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return Value(map[string]Value{
			"success": false,
			"error":   err.Error(),
			"conn":    Value(nil),
		})
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return Value(map[string]Value{
			"success": false,
			"error":   fmt.Sprintf("Failed to ping database: %v", err),
			"conn":    Value(nil),
		})
	}

	// Generate connection ID
	connID := fmt.Sprintf("%s_%s_%d_%s_%d", driver, host, port, database, time.Now().UnixNano())

	// Store connection
	databaseMutex.Lock()
	databaseConnections[connID] = db
	databaseMutex.Unlock()

	// Create connection object
	connObj := &DatabaseConnection{
		ID:       connID,
		DB:       db,
		Driver:   driver,
		Host:     host,
		Port:     port,
		Database: database,
		Username: username,
		Password: password,
	}

	return Value(map[string]Value{
		"success":  true,
		"conn":     connObj,
		"id":       connID,
		"driver":   driver,
		"host":     host,
		"port":     port,
		"database": database,
	})
}

// dbQueryFunc implements the db_query() built-in function
// db_query(connection, query, params...) -> result_object
// Example: result = db_query(conn, "SELECT * FROM users WHERE id = $1", 123)
func dbQueryFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 {
		panic(typeError(pos, "db_query() requires at least 2 arguments: connection and query"))
	}

	// Get connection
	connObj, ok := args[0].(*DatabaseConnection)
	if !ok {
		panic(typeError(pos, "db_query() requires first argument to be a database connection"))
	}

	query, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "db_query() requires second argument to be a string (query)"))
	}

	// Prepare parameters
	params := make([]any, len(args)-2)
	for i, arg := range args[2:] {
		params[i] = arg
	}

	// Execute query
	rows, err := connObj.DB.Query(query, params...)
	if err != nil {
		return Value(map[string]Value{
			"success": false,
			"error":   err.Error(),
			"data":    Value(nil),
		})
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return Value(map[string]Value{
			"success": false,
			"error":   err.Error(),
			"data":    Value(nil),
		})
	}

	// Scan results
	var results []Value
	for rows.Next() {
		// Create slice to hold column values
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan row
		if err := rows.Scan(valuePtrs...); err != nil {
			return Value(map[string]Value{
				"success": false,
				"error":   err.Error(),
				"data":    Value(nil),
			})
		}

		// Convert to map
		row := make(map[string]Value)
		for i, col := range columns {
			if values[i] != nil {
				row[col] = convertSQLValue(values[i])
			} else {
				row[col] = Value(nil)
			}
		}
		results = append(results, Value(row))
	}

	if err := rows.Err(); err != nil {
		return Value(map[string]Value{
			"success": false,
			"error":   err.Error(),
			"data":    Value(nil),
		})
	}

	return Value(map[string]Value{
		"success": true,
		"data":    &results,
		"count":   len(results),
		"columns": columns,
	})
}

// dbExecuteFunc implements the db_execute() built-in function
// db_execute(connection, query, params...) -> result_object
// Example: result = db_execute(conn, "INSERT INTO users (name, email) VALUES ($1, $2)", "John", "john@example.com")
func dbExecuteFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 {
		panic(typeError(pos, "db_execute() requires at least 2 arguments: connection and query"))
	}

	// Get connection
	connObj, ok := args[0].(*DatabaseConnection)
	if !ok {
		panic(typeError(pos, "db_execute() requires first argument to be a database connection"))
	}

	query, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "db_execute() requires second argument to be a string (query)"))
	}

	// Prepare parameters
	params := make([]any, len(args)-2)
	for i, arg := range args[2:] {
		params[i] = arg
	}

	// Execute query
	result, err := connObj.DB.Exec(query, params...)
	if err != nil {
		return Value(map[string]Value{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Get affected rows
	rowsAffected, _ := result.RowsAffected()
	lastInsertID, _ := result.LastInsertId()

	return Value(map[string]Value{
		"success":        true,
		"rows_affected":  int(rowsAffected),
		"last_insert_id": int(lastInsertID),
	})
}

// streamTablesFunc implements the stream_tables() built-in function for real-time streaming
// stream_tables(connection, table_name, callback) -> nil
// Example: stream_tables(conn, "messages", onDataChange)
// Uses PostgreSQL LISTEN/NOTIFY for real-time streaming without polling
func streamTablesFunc(interp *interpreter, pos Position, args []Value) Value {
	// Check if this is multi-table streaming
	if len(args) == 3 {
		// Check if second argument is an array of table names
		if tableArrayPtr, ok := args[1].(*[]Value); ok {
			return streamTablesMultiFunc(interp, pos, args[0], *tableArrayPtr, args[2])
		}
	}
	return streamTablesSingleFunc(interp, pos, args)
}

// streamTablesSingleFunc handles single table streaming
func streamTablesSingleFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 3 {
		panic(typeError(pos, "stream_tables() requires 3 arguments: connection, table_name, callback"))
	}

	// Get connection
	connObj, ok := args[0].(*DatabaseConnection)
	if !ok {
		panic(typeError(pos, "stream_tables() requires first argument to be a database connection"))
	}

	// Support both PostgreSQL and MySQL
	if connObj.Driver != "postgres" && connObj.Driver != "mysql" {
		panic(typeError(pos, "stream_tables() only supports PostgreSQL and MySQL databases"))
	}

	tableName, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "stream_tables() requires second argument to be a string (table_name)"))
	}

	callback := args[2] // Function to call on change

	// Generate streamer ID
	streamerID := fmt.Sprintf("streamer_%s_%d", connObj.ID, time.Now().UnixNano())

	// Create PostgreSQL listener
	// For simplicity, we'll create a new connection for listening
	// In production, you might want to reuse the existing connection
	listener := pq.NewListener(fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
		connObj.Host, connObj.Port, connObj.Username, connObj.Database), 10*time.Second, time.Minute, nil)

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Create streamer
	streamer := &DatabaseStreamer{
		ID:        streamerID,
		DB:        connObj.DB,
		Listener:  listener,
		TableName: tableName,
		Callback:  callback,
		Active:    true,
		Cancel:    cancel,
	}

	// Store streamer
	databaseMutex.Lock()
	databaseStreamers[streamerID] = streamer
	databaseMutex.Unlock()

	// Setup database trigger and notification
	go func() {
		defer func() {
			// Mark streamer as inactive when goroutine exits
			streamer.Mutex.Lock()
			streamer.Active = false
			streamer.Mutex.Unlock()

			// Close listener
			if streamer.Listener != nil {
				streamer.Listener.Close()
			}
		}()

		// Create notification channel name
		channelName := fmt.Sprintf("%s_changes", tableName)

		// Setup trigger function if it doesn't exist
		setupTriggerSQL := fmt.Sprintf(`
			CREATE OR REPLACE FUNCTION notify_%s_change()
			RETURNS TRIGGER AS $$
			BEGIN
				PERFORM pg_notify('%s',
					json_build_object(
						'table', TG_TABLE_NAME,
						'operation', TG_OP,
						'timestamp', extract(epoch from now()),
						'data', CASE
							WHEN TG_OP = 'DELETE' THEN row_to_json(OLD)
							ELSE row_to_json(NEW)
						END
					)::text
				);
				RETURN CASE TG_OP
					WHEN 'DELETE' THEN OLD
					ELSE NEW
				END;
			END;
			$$ LANGUAGE plpgsql;
		`, tableName, channelName)

		// Create trigger
		triggerSQL := fmt.Sprintf(`
			DROP TRIGGER IF EXISTS %s_notify ON %s;
			CREATE TRIGGER %s_notify
				AFTER INSERT OR UPDATE OR DELETE ON %s
				FOR EACH ROW EXECUTE FUNCTION notify_%s_change();
		`, tableName, tableName, tableName, tableName, tableName)

		// Execute setup queries
		if _, err := streamer.DB.Exec(setupTriggerSQL); err != nil {
			log.Printf("Error setting up trigger function: %v", err)
			return
		}

		if _, err := streamer.DB.Exec(triggerSQL); err != nil {
			log.Printf("Error setting up trigger: %v", err)
			return
		}

		// Listen to the channel
		if err := streamer.Listener.Listen(channelName); err != nil {
			log.Printf("Error starting listener: %v", err)
			return
		}

		log.Printf("Started streaming for table '%s' on channel '%s'", tableName, channelName)

		// Listen for notifications
		for {
			select {
			case <-ctx.Done():
				return
			case notification := <-streamer.Listener.Notify:
				if notification != nil {
					// Call the callback function with the notification data
					streamer.handleNotification(interp, pos, notification)
				}
			case <-time.After(90 * time.Second):
				// Ping to keep connection alive
				go func() {
					streamer.Listener.Ping()
				}()
			}
		}
	}()

	// Block the main thread to keep program alive (same as query_listen)
	// Wait for context cancellation (Ctrl+C or manual stop)
	<-ctx.Done()

	// Pure callback pattern - no return value
	return Value(nil)
}

// streamTablesMultiFunc implements multi-table streaming
// stream_tables(connection, ["table1", "table2"], callback) -> nil
func streamTablesMultiFunc(interp *interpreter, pos Position, connArg Value, tableArray []Value, callback Value) Value {
	// Get connection
	connObj, ok := connArg.(*DatabaseConnection)
	if !ok {
		panic(typeError(pos, "stream_tables() requires first argument to be a database connection"))
	}

	// Support both PostgreSQL and MySQL
	if connObj.Driver != "postgres" && connObj.Driver != "mysql" {
		panic(typeError(pos, "stream_tables() only supports PostgreSQL and MySQL databases"))
	}

	// Convert table array to string slice
	var tableNames []string
	for i, tableVal := range tableArray {
		tableName, ok := tableVal.(string)
		if !ok {
			panic(typeError(pos, fmt.Sprintf("stream_tables() table array element %d must be a string", i)))
		}
		tableNames = append(tableNames, tableName)
	}

	if len(tableNames) == 0 {
		panic(typeError(pos, "stream_tables() requires at least one table name"))
	}

	// Generate streamer ID for multi-table
	streamerID := fmt.Sprintf("multi_streamer_%s_%d", connObj.ID, time.Now().UnixNano())

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Create streamer based on driver
	var streamer *DatabaseStreamer
	if connObj.Driver == "postgres" {
		// Create PostgreSQL listener
		listener := pq.NewListener(fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
			connObj.Host, connObj.Port, connObj.Username, connObj.Database), 10*time.Second, time.Minute, nil)

		streamer = &DatabaseStreamer{
			ID:        streamerID,
			DB:        connObj.DB,
			Listener:  listener,
			TableName: strings.Join(tableNames, ","),
			Callback:  callback,
			Active:    true,
			Cancel:    cancel,
			Driver:    "postgres",
		}
	} else {
		// MySQL streamer with optimized CDC
		streamer = &DatabaseStreamer{
			ID:        streamerID,
			DB:        connObj.DB,
			Listener:  nil,
			TableName: strings.Join(tableNames, ","),
			Callback:  callback,
			Active:    true,
			Cancel:    cancel,
			Driver:    "mysql",
			LastID:    0,
			UseBinlog: true, // Enable optimized CDC
			ServerID:  1001, // Default server ID for binlog replication
		}
	}

	// Store streamer
	databaseMutex.Lock()
	databaseStreamers[streamerID] = streamer
	databaseMutex.Unlock()

	// Setup database streaming based on driver
	go func() {
		defer func() {
			// Mark streamer as inactive when goroutine exits
			streamer.Mutex.Lock()
			streamer.Active = false
			streamer.Mutex.Unlock()

			// Close listener for PostgreSQL
			if streamer.Listener != nil {
				streamer.Listener.Close()
			}
		}()

		if connObj.Driver == "postgres" {
			streamer.setupPostgreSQLMultiTableStreaming(interp, pos, ctx, tableNames)
		} else {
			// Use optimized CDC for MySQL
			if streamer.UseBinlog {
				streamer.setupMySQLBinlogStreaming(interp, pos, ctx, tableNames, connObj.Host, connObj.Port, connObj.Username, connObj.Password)
			} else {
				streamer.setupMySQLMultiTableStreaming(interp, pos, ctx, tableNames)
			}
		}
	}()

	// Block the main thread to keep program alive (same as query_listen)
	// Wait for context cancellation (Ctrl+C or manual stop)
	<-ctx.Done()

	// Pure callback pattern - no return value
	return Value(nil)
}

// handleNotification processes PostgreSQL notifications and calls the callback
func (ds *DatabaseStreamer) handleNotification(interp *interpreter, pos Position, notification *pq.Notification) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Error in stream callback: %v", r)
		}
	}()

	// Parse the JSON payload
	payload := notification.Extra
	log.Printf("Received notification on channel '%s': %s", notification.Channel, payload)

	// Call the callback function with the notification data
	if fn, ok := ds.Callback.(functionType); ok {
		// Parse JSON payload to extract structured data
		var payloadData map[string]Value
		if jsonResult := jsonParseFunc(interp, pos, []Value{payload}); jsonResult != nil {
			if parsedData, ok := jsonResult.(map[string]Value); ok {
				payloadData = parsedData
			} else {
				// Fallback: create simple data structure
				payloadData = map[string]Value{
					"channel": notification.Channel,
					"payload": payload,
				}
			}
		} else {
			// Fallback: create simple data structure
			payloadData = map[string]Value{
				"channel": notification.Channel,
				"payload": payload,
			}
		}

		// Add channel information to the data
		payloadData["channel"] = notification.Channel

		// Create arguments for callback: channel and payload as separate parameters
		callbackArgs := []Value{notification.Channel, payload}

		// Call the callback function
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Database stream callback error: %v", r)
				}
			}()
			fn.call(interp, pos, callbackArgs)
		}()
	} else {
		log.Printf("Invalid callback function for streamer %s", ds.ID)
	}
}

// dbStopStreamFunc implements the db_stop_stream() built-in function
// db_stop_stream(table_name) -> result_object
// Example: result = db_stop_stream("messages")
func dbStopStreamFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		panic(typeError(pos, "db_stop_stream() requires 1 argument: table_name"))
	}

	tableName, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "db_stop_stream() requires first argument to be a string (table_name)"))
	}

	databaseMutex.Lock()
	defer databaseMutex.Unlock()

	// Find and stop streamers for this table
	stoppedCount := 0
	for id, streamer := range databaseStreamers {
		if streamer.TableName == tableName && streamer.Active {
			streamer.Cancel()
			delete(databaseStreamers, id)
			stoppedCount++
		}
	}

	result := map[string]Value{
		"success":       true,
		"stopped_count": stoppedCount,
		"table_name":    tableName,
	}

	return result
}

// dbCloseFunc implements the db_close() built-in function
// db_close(connection) -> result_object
// Example: result = db_close(conn)
func dbCloseFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		panic(typeError(pos, "db_close() requires 1 argument: connection"))
	}

	connObj, ok := args[0].(*DatabaseConnection)
	if !ok {
		panic(typeError(pos, "db_close() requires first argument to be a database connection"))
	}

	// Close database connection
	databaseMutex.Lock()
	delete(databaseConnections, connObj.ID)
	databaseMutex.Unlock()

	err := connObj.DB.Close()
	if err != nil {
		return Value(map[string]Value{
			"success": false,
			"error":   err.Error(),
		})
	}

	return Value(map[string]Value{
		"success": true,
		"message": "Connection closed successfully",
	})
}

// Helper function to convert SQL values to UDDIN-LANG values
func convertSQLValue(value any) Value {
	switch v := value.(type) {
	case []byte:
		return Value(string(v))
	case time.Time:
		return Value(v.Unix())
	case int64:
		return Value(int(v))
	case float64:
		return Value(v)
	case string:
		return Value(v)
	case bool:
		return Value(v)
	case nil:
		return Value(nil)
	default:
		return Value(fmt.Sprintf("%v", v))
	}
}

// setupPostgreSQLMultiTableStreaming sets up PostgreSQL LISTEN/NOTIFY for multiple tables
func (ds *DatabaseStreamer) setupPostgreSQLMultiTableStreaming(interp *interpreter, pos Position, ctx context.Context, tableNames []string) {
	// Setup triggers for each table
	for _, tableName := range tableNames {
		channelName := fmt.Sprintf("%s_changes", tableName)

		// Setup trigger function if it doesn't exist
		setupTriggerSQL := fmt.Sprintf(`
			CREATE OR REPLACE FUNCTION notify_%s_change()
			RETURNS TRIGGER AS $$
			BEGIN
				PERFORM pg_notify('%s',
					json_build_object(
						'table', TG_TABLE_NAME,
						'operation', TG_OP,
						'timestamp', extract(epoch from now()),
						'data', CASE
							WHEN TG_OP = 'DELETE' THEN row_to_json(OLD)
							ELSE row_to_json(NEW)
						END
					)::text
				);
				RETURN CASE TG_OP
					WHEN 'DELETE' THEN OLD
					ELSE NEW
				END;
			END;
			$$ LANGUAGE plpgsql;
		`, tableName, channelName)

		// Create trigger
		triggerSQL := fmt.Sprintf(`
			DROP TRIGGER IF EXISTS %s_notify ON %s;
			CREATE TRIGGER %s_notify
				AFTER INSERT OR UPDATE OR DELETE ON %s
				FOR EACH ROW EXECUTE FUNCTION notify_%s_change();
		`, tableName, tableName, tableName, tableName, tableName)

		// Execute setup queries
		if _, err := ds.DB.Exec(setupTriggerSQL); err != nil {
			log.Printf("Error setting up trigger function for table %s: %v", tableName, err)
			return
		}

		if _, err := ds.DB.Exec(triggerSQL); err != nil {
			log.Printf("Error setting up trigger for table %s: %v", tableName, err)
			return
		}

		// Listen to the channel
		if err := ds.Listener.Listen(channelName); err != nil {
			log.Printf("Error starting listener for table %s: %v", tableName, err)
			return
		}
	}

	log.Printf("Started multi-table streaming for tables %v", tableNames)

	// Listen for notifications from all channels
	for {
		select {
		case <-ctx.Done():
			return
		case notification := <-ds.Listener.Notify:
			if notification != nil {
				// Call the callback function with the notification data
				ds.handleNotification(interp, pos, notification)
			}
		case <-time.After(90 * time.Second):
			// Ping to keep connection alive
			go func() {
				ds.Listener.Ping()
			}()
		}
	}
}

// setupMySQLBinlogStreaming sets up true MySQL binlog streaming like Debezium
func (ds *DatabaseStreamer) setupMySQLBinlogStreaming(interp *interpreter, pos Position, ctx context.Context, tableNames []string, host string, port int, username, password string) {
	// Initialize binlog configuration
	cfg := replication.BinlogSyncerConfig{
		ServerID: ds.ServerID,
		Flavor:   "mysql",
		Host:     host,
		Port:     uint16(port),
		User:     username,
		Password: password,
	}

	// Create binlog syncer
	syncer := replication.NewBinlogSyncer(cfg)
	defer syncer.Close()

	// Get current binlog position
	binlogPos, err := ds.getCurrentBinlogPosition()
	if err != nil {
		log.Printf("Error getting binlog position: %v, starting from latest", err)
		binlogPos = mysql.Position{Name: "", Pos: 4} // Start from beginning if no position found
	}

	// Start streaming from the position
	streamer, err := syncer.StartSync(binlogPos)
	if err != nil {
		log.Printf("Error starting binlog sync: %v", err)
		return
	}

	// Initialize binlog streamer
	ds.Binlog = &BinlogStreamer{
		Syncer:   syncer,
		Streamer: streamer,
		Config:   cfg,
		Position: binlogPos,
		Tables:   make(map[string]bool),
	}

	// Mark tables to monitor
	for _, tableName := range tableNames {
		ds.Binlog.Tables[tableName] = true
	}

	log.Printf("Started MySQL binlog streaming for tables %v from position %v", tableNames, binlogPos)

	// Start processing binlog events
	ds.processBinlogEvents(interp, pos, ctx)
}

// setupMySQLMultiTableStreaming is deprecated - use setupMySQLBinlogStreaming instead
// This function is kept for backward compatibility but should not be used
func (ds *DatabaseStreamer) setupMySQLMultiTableStreaming(interp *interpreter, pos Position, ctx context.Context, tableNames []string) {
	log.Printf("Warning: setupMySQLMultiTableStreaming is deprecated. Use binlog streaming instead.")
	// Fallback to binlog streaming
	ds.setupMySQLBinlogStreaming(interp, pos, ctx, tableNames, "", 0, "", "")
}

// getCurrentBinlogPosition gets the current MySQL binlog position
func (ds *DatabaseStreamer) getCurrentBinlogPosition() (mysql.Position, error) {
	var file string
	var position uint32

	query := "SHOW MASTER STATUS"
	row := ds.DB.QueryRow(query)
	err := row.Scan(&file, &position, nil, nil, nil)
	if err != nil {
		return mysql.Position{}, err
	}

	return mysql.Position{Name: file, Pos: position}, nil
}

// processBinlogEvents processes MySQL binlog events in real-time
func (ds *DatabaseStreamer) processBinlogEvents(interp *interpreter, pos Position, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Get next binlog event
			ev, err := ds.Binlog.Streamer.GetEvent(ctx)
			if err != nil {
				log.Printf("Error getting binlog event: %v", err)
				continue
			}

			// Update binlog position
			ds.Binlog.Position.Pos = ev.Header.LogPos

			// Process different event types
			switch e := ev.Event.(type) {
			case *replication.RowsEvent:
				// Handle INSERT, UPDATE, DELETE events
				ds.handleRowsEvent(interp, pos, ev.Header, e)
			case *replication.RotateEvent:
				// Handle binlog rotation
				ds.Binlog.Position.Name = string(e.NextLogName)
				ds.Binlog.Position.Pos = uint32(e.Position)
				log.Printf("Binlog rotated to %s:%d", ds.Binlog.Position.Name, ds.Binlog.Position.Pos)
			}
		}
	}
}

// handleRowsEvent processes MySQL row change events
func (ds *DatabaseStreamer) handleRowsEvent(interp *interpreter, pos Position, header *replication.EventHeader, e *replication.RowsEvent) {
	tableName := string(e.Table.Table)

	// Check if we're monitoring this table
	if !ds.Binlog.Tables[tableName] {
		return
	}

	// Determine operation type
	var operation string
	switch header.EventType {
	case replication.WRITE_ROWS_EVENTv1, replication.WRITE_ROWS_EVENTv2:
		operation = "INSERT"
	case replication.UPDATE_ROWS_EVENTv1, replication.UPDATE_ROWS_EVENTv2:
		operation = "UPDATE"
	case replication.DELETE_ROWS_EVENTv1, replication.DELETE_ROWS_EVENTv2:
		operation = "DELETE"
	default:
		return
	}

	// Process each row in the event
	for i, row := range e.Rows {
		var oldData, newData map[string]any

		switch operation {
		case "UPDATE":
			// For UPDATE, rows come in pairs: [old_row, new_row]
			if i%2 == 0 {
				// Skip the old row, it will be processed with the new row
				continue
			} else {
				newData = ds.rowToMap(e.Table, row)
				oldData = ds.rowToMap(e.Table, e.Rows[i-1])

				// Call callback for UPDATE
				ds.callBinlogCallback(interp, pos, tableName, operation, oldData, newData)
			}
		case "INSERT":
			newData = ds.rowToMap(e.Table, row)
			ds.callBinlogCallback(interp, pos, tableName, operation, nil, newData)
		case "DELETE":
			oldData = ds.rowToMap(e.Table, row)
			ds.callBinlogCallback(interp, pos, tableName, operation, oldData, nil)
		}
	}
}

// getColumnList retrieves the column names for a given table (used by binlog streaming)
func (ds *DatabaseStreamer) getColumnList(tableName string) ([]string, error) {
	query := "SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = ? AND TABLE_SCHEMA = DATABASE() ORDER BY ORDINAL_POSITION"
	rows, err := ds.DB.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, err
		}
		columns = append(columns, columnName)
	}

	return columns, nil
}

// rowToMap converts a binlog row to a map
func (ds *DatabaseStreamer) rowToMap(table *replication.TableMapEvent, row []any) map[string]any {
	result := make(map[string]any)

	// Get column information
	columns, err := ds.getColumnList(string(table.Table))
	if err != nil {
		log.Printf("Error getting columns for table %s: %v", string(table.Table), err)
		return result
	}

	// Map row values to column names
	for i, value := range row {
		if i < len(columns) {
			result[columns[i]] = value
		}
	}

	return result
}

// callBinlogCallback calls the UDDIN callback function for binlog events
func (ds *DatabaseStreamer) callBinlogCallback(interp *interpreter, pos Position, tableName, operation string, oldData, newData map[string]any) {
	// Create payload similar to the polling version
	payload := map[string]any{
		"table":     tableName,
		"operation": operation,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if newData != nil {
		payload["new_data"] = newData
	}
	if oldData != nil {
		payload["old_data"] = oldData
	}

	// Convert to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling binlog payload: %v", err)
		return
	}

	// Call the callback function using the same pattern as handleNotification
	ds.Mutex.RLock()
	callback := ds.Callback
	ds.Mutex.RUnlock()

	if fn, ok := callback.(functionType); ok {
		// Create arguments for callback: channel and payload as separate parameters
		channelName := fmt.Sprintf("%s_changes", tableName)
		callbackArgs := []Value{channelName, string(payloadJSON)}

		// Call the callback function
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Binlog stream callback error: %v", r)
				}
			}()
			fn.call(interp, pos, callbackArgs)
		}()
	} else {
		log.Printf("Invalid callback function for binlog streamer %s", ds.ID)
	}
}
