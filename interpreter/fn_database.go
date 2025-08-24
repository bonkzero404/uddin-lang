package interpreter

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
)

// ConnectionPoolConfig holds configuration for database connection pool
type ConnectionPoolConfig struct {
	MaxOpenConns    int           // Maximum number of open connections
	MaxIdleConns    int           // Maximum number of idle connections
	ConnMaxLifetime time.Duration // Maximum lifetime of a connection
	ConnMaxIdleTime time.Duration // Maximum idle time of a connection
}

// ConnectionPool manages a pool of database connections
type ConnectionPool struct {
	DB     *sql.DB
	Config ConnectionPoolConfig
	DSN    string
	Driver string
	Mutex  sync.RWMutex
}



// NewConnectionPool creates a new connection pool
func NewConnectionPool(driver, dsn string, config ConnectionPoolConfig) (*ConnectionPool, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return &ConnectionPool{
		DB:     db,
		Config: config,
		DSN:    dsn,
		Driver: driver,
	}, nil
}

// GetConnection returns the underlying database connection
func (cp *ConnectionPool) GetConnection() *sql.DB {
	cp.Mutex.RLock()
	defer cp.Mutex.RUnlock()
	return cp.DB
}

// Close closes the connection pool
func (cp *ConnectionPool) Close() error {
	cp.Mutex.Lock()
	defer cp.Mutex.Unlock()
	return cp.DB.Close()
}

// Database connection pool and listener management
var (
	databaseConnections = make(map[string]*ConnectionPool)
	databaseStreamers   = make(map[string]*DatabaseStreamer)
	databaseMutex       = sync.RWMutex{}
	// Default connection pool configuration
	defaultPoolConfig = ConnectionPoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}
)

// BinlogStreamer represents a MySQL binlog streamer for CDC
type BinlogStreamer struct {
	Syncer   *replication.BinlogSyncer
	Streamer *replication.BinlogStreamer
	Config   replication.BinlogSyncerConfig
	Position mysql.Position
	Tables   map[string]bool // Tables to monitor
}

// TableColumnConfig represents column configuration for a specific table
type TableColumnConfig struct {
	Table   string   `json:"table"`
	Columns []string `json:"columns"`
}

// DatabaseStreamer represents a real-time database streamer using LISTEN/NOTIFY or Binlog
type DatabaseStreamer struct {
	ID           string
	DB           *sql.DB
	Listener     *pq.Listener
	Binlog       *BinlogStreamer
	TableName    string
	Callback     Value // UDDIN-LANG function to call on data change
	Active       bool
	Cancel       context.CancelFunc
	Mutex        sync.RWMutex
	Driver       string
	LastID       int64
	BinlogPos    string              // For MySQL binary log position
	UseBinlog    bool                // Whether to use binlog streaming instead of polling
	ServerID     uint32              // MySQL server ID for binlog replication
	TableColumns map[string][]string // Selected columns per table (empty means all columns)
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

	// Create connection pool
	pool, err := NewConnectionPool(driver, dsn, defaultPoolConfig)
	if err != nil {
		return Value(map[string]Value{
			"success": false,
			"error":   err.Error(),
			"conn":    Value(nil),
		})
	}

	// Generate connection ID
	connID := fmt.Sprintf("%s_%s_%d_%s_%d", driver, host, port, database, time.Now().UnixNano())

	// Store connection pool
	databaseMutex.Lock()
	databaseConnections[connID] = pool
	databaseMutex.Unlock()

	// Create connection object
	connObj := &DatabaseConnection{
		ID:       connID,
		DB:       pool.GetConnection(),
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

// BatchOperation represents a single operation in a batch
type BatchOperation struct {
	Query string
	Args  []Value
}

// BatchResult represents the result of a batch operation
type BatchResult struct {
	Success      bool
	RowsAffected int64
	Error        string
	OperationID  int
}

// dbExecuteBatchFunc implements the db_execute_batch() built-in function
// db_execute_batch(connection, operations_array) -> batch_result_object
// Example: result = db_execute_batch(conn, [{"query": "INSERT INTO users (name) VALUES (?)", "args": ["John"]}, {"query": "UPDATE users SET age = ? WHERE name = ?", "args": [25, "John"]}])
func dbExecuteBatchFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 2 {
		panic(typeError(pos, "db_execute_batch() requires 2 arguments: connection and operations array"))
	}

	connArg := args[0]
	operationsArg := args[1]

	// Extract connection
	conn, ok := connArg.(*DatabaseConnection)
	if !ok {
		panic(typeError(pos, "db_execute_batch() requires first argument to be a database connection"))
	}

	// Extract operations array
	var operationsArray []Value
	if arr, ok := operationsArg.(*[]Value); ok {
		operationsArray = *arr
	} else if arr, ok := operationsArg.([]Value); ok {
		operationsArray = arr
	} else {
		panic(typeError(pos, "db_execute_batch() requires second argument to be an array of operations"))
	}

	if len(operationsArray) == 0 {
		return Value(map[string]Value{
			"success": false,
			"error":   "No operations provided",
			"results": []Value{},
		})
	}

	// Parse operations
	var operations []BatchOperation
	for i, opValue := range operationsArray {
		opMap, ok := opValue.(map[string]Value)
		if !ok {
			return Value(map[string]Value{
				"success": false,
				"error":   fmt.Sprintf("Operation %d must be an object with 'query' and 'args' fields", i),
				"results": []Value{},
			})
		}

		queryValue, hasQuery := opMap["query"]
		if !hasQuery {
			return Value(map[string]Value{
				"success": false,
				"error":   fmt.Sprintf("Operation %d missing 'query' field", i),
				"results": []Value{},
			})
		}

		query, ok := queryValue.(string)
		if !ok {
			return Value(map[string]Value{
				"success": false,
				"error":   fmt.Sprintf("Operation %d 'query' field must be a string", i),
				"results": []Value{},
			})
		}

		var opArgs []Value
		if argsValue, hasArgs := opMap["args"]; hasArgs {
			if argsArray, ok := argsValue.(*[]Value); ok {
				opArgs = *argsArray
			} else if argsArray, ok := argsValue.([]Value); ok {
				opArgs = argsArray
			} else {
				return Value(map[string]Value{
					"success": false,
					"error":   fmt.Sprintf("Operation %d 'args' field must be an array", i),
					"results": []Value{},
				})
			}
		}

		operations = append(operations, BatchOperation{
			Query: query,
			Args:  opArgs,
		})
	}

	// Execute batch operations
	results := make([]Value, len(operations))
	allSuccess := true
	totalRowsAffected := int64(0)

	// Begin transaction for batch operations
	tx, err := conn.DB.Begin()
	if err != nil {
		return Value(map[string]Value{
			"success": false,
			"error":   fmt.Sprintf("Failed to begin transaction: %v", err),
			"results": []Value{},
		})
	}

	defer func() {
		if allSuccess {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	// Execute each operation in the transaction
	for i, op := range operations {
		// Convert UDDIN-LANG values to Go any values
		var sqlArgs []any
		for _, arg := range op.Args {
			sqlArgs = append(sqlArgs, convertToSQLArg(arg))
		}

		// Convert placeholders for PostgreSQL
		convertedQuery := convertPlaceholders(op.Query, conn.Driver)

		// Execute the query
		result, err := tx.Exec(convertedQuery, sqlArgs...)
		if err != nil {
			allSuccess = false
			results[i] = Value(map[string]Value{
				"success":       false,
				"error":         err.Error(),
				"rows_affected": int64(0),
				"operation_id":  i,
			})
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		totalRowsAffected += rowsAffected

		results[i] = Value(map[string]Value{
			"success":       true,
			"error":         "",
			"rows_affected": rowsAffected,
			"operation_id":  i,
		})
	}

	return Value(map[string]Value{
		"success":             allSuccess,
		"total_rows_affected": totalRowsAffected,
		"operations_count":    len(operations),
		"results":             Value(&results),
		"error":               "",
	})
}

// convertToSQLArg converts UDDIN-LANG Value to SQL argument
func convertToSQLArg(value Value) any {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return v
	case float64:
		return v
	case bool:
		return v
	case nil:
		return nil
	default:
		return fmt.Sprintf("%v", v)
	}
}

// convertPlaceholders converts MySQL-style ? placeholders to PostgreSQL-style $1, $2, etc.
func convertPlaceholders(query string, driver string) string {
	if driver != "postgres" {
		return query
	}

	result := ""
	placeholderCount := 1
	for _, char := range query {
		if char == '?' {
			result += fmt.Sprintf("$%d", placeholderCount)
			placeholderCount++
		} else {
			result += string(char)
		}
	}
	return result
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
// stream_tables(connection, table_name, callback, columns) -> nil (with column selection)
// stream_tables(connection, [table_names], callback) -> nil (multi-table)
// stream_tables(connection, [table_names], callback, columns) -> nil (multi-table with column selection)
// Example: stream_tables(conn, "messages", onDataChange)
// Example: stream_tables(conn, "messages", onDataChange, ["id", "name", "email"])
// Uses PostgreSQL LISTEN/NOTIFY for real-time streaming without polling
func streamTablesFunc(interp *interpreter, pos Position, args []Value) Value {
	// Check argument count (3 or 4 arguments supported)
	if len(args) < 3 || len(args) > 4 {
		panic(typeError(pos, "stream_tables() requires 3 or 4 arguments: connection, table_name(s), callback, [table_columns]"))
	}

	// Check if this is multi-table streaming
	if len(args) >= 3 {
		// Check if second argument is an array of table names
		if tableArrayPtr, ok := args[1].(*[]Value); ok {
			// Multi-table streaming
			var tableColumns any
			if len(args) == 4 {
				// Fourth argument can be:
				// 1. Array of strings (backward compatibility)
				// 2. Array of table-column config objects
				if columnArrayPtr, ok := args[3].(*[]Value); ok {
					// Check if it's array of strings or objects
					if len(*columnArrayPtr) > 0 {
						if _, isString := (*columnArrayPtr)[0].(string); isString {
							// Backward compatibility: array of strings
							var columns []string
							for i, colVal := range *columnArrayPtr {
								if colName, ok := colVal.(string); ok {
									columns = append(columns, colName)
								} else {
						panic(typeError(pos, "stream_tables() column array element %d must be a string", i))
					}
							}
							tableColumns = columns
						} else {
							// New format: array of table-column config objects
							tableColumns = *columnArrayPtr
						}
					}
				} else {
					panic(typeError(pos, "stream_tables() fourth argument must be an array"))
				}
			}
			return streamTablesMultiFunc(interp, pos, args[0], *tableArrayPtr, args[2], tableColumns)
		}
	}
	// Single table streaming
	return streamTablesSingleFunc(interp, pos, args)
}

// streamTablesSingleFunc handles single table streaming
func streamTablesSingleFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 3 || len(args) > 4 {
		panic(typeError(pos, "stream_tables() requires 3 or 4 arguments: connection, table_name, callback, [columns]"))
	}

	// Extract column selection if provided
	var columns []string
	if len(args) == 4 {
		if columnArrayPtr, ok := args[3].(*[]Value); ok {
			for i, colVal := range *columnArrayPtr {
				if colName, ok := colVal.(string); ok {
					columns = append(columns, colName)
				} else {
					panic(typeError(pos, "stream_tables() column array element %d must be a string", i))
				}
			}
		} else {
			panic(typeError(pos, "stream_tables() fourth argument must be an array of column names"))
		}
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

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Create streamer based on database type
	var streamer *DatabaseStreamer
	if connObj.Driver == "postgres" {
		// Create PostgreSQL listener
		listener := pq.NewListener(fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
			connObj.Host, connObj.Port, connObj.Username, connObj.Database), 10*time.Second, time.Minute, nil)

		streamer = &DatabaseStreamer{
			ID:           streamerID,
			DB:           connObj.DB,
			Listener:     listener,
			TableName:    tableName,
			Callback:     callback,
			Active:       true,
			Cancel:       cancel,
			Driver:       "postgres",
			TableColumns: make(map[string][]string),
		}
	} else {
		// MySQL - use binlog streaming
		streamer = &DatabaseStreamer{
			ID:           streamerID,
			DB:           connObj.DB,
			TableName:    tableName,
			Callback:     callback,
			Active:       true,
			Cancel:       cancel,
			Driver:       "mysql",
			UseBinlog:    true,
			ServerID:     uint32(time.Now().Unix() % 1000000), // Generate unique server ID
			TableColumns: make(map[string][]string),
		}
	}

	// Set columns for the single table
	if len(columns) > 0 {
		streamer.TableColumns[tableName] = columns
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
			// PostgreSQL streaming using LISTEN/NOTIFY
			// Create notification channel name
			channelName := fmt.Sprintf("%s_changes", tableName)

			// Build column selection for trigger function
			var dataExpression string
			if len(columns) > 0 {
				// Create JSON object with only selected columns
				columnSelections := make([]string, len(columns))
				for i, col := range columns {
					columnSelections[i] = fmt.Sprintf("'%s', CASE WHEN TG_OP = 'DELETE' THEN OLD.%s ELSE NEW.%s END", col, col, col)
				}
				dataExpression = fmt.Sprintf("json_build_object(%s)", strings.Join(columnSelections, ", "))
			} else {
				// Include all columns (default behavior)
				dataExpression = "CASE WHEN TG_OP = 'DELETE' THEN row_to_json(OLD) ELSE row_to_json(NEW) END"
			}

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
							'data', %s
						)::text
					);
					RETURN CASE TG_OP
						WHEN 'DELETE' THEN OLD
						ELSE NEW
					END;
				END;
				$$ LANGUAGE plpgsql;
			`, tableName, channelName, dataExpression)

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
		} else {
			// MySQL streaming using binlog
			tableNames := []string{tableName}
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

// streamTablesMultiFunc implements multi-table streaming
// stream_tables(connection, ["table1", "table2"], callback) -> nil
// stream_tables(connection, ["table1", "table2"], callback, ["col1", "col2"]) -> nil (with column selection)
// stream_tables(connection, ["table1", "table2"], callback, [{"table": "table1", "columns": ["col1", "col2"]}]) -> nil (with table-column mapping)
func streamTablesMultiFunc(interp *interpreter, pos Position, connArg Value, tableArray []Value, callback Value, tableColumns any) Value {
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
			panic(typeError(pos, "stream_tables() table array element %d must be a string", i))
		}
		tableNames = append(tableNames, tableName)
	}

	if len(tableNames) == 0 {
		panic(typeError(pos, "stream_tables() requires at least one table name"))
	}

	// Parse table-column configuration
	tableColumnMap := make(map[string][]string)
	if tableColumns != nil {
		// Try to parse as array of TableColumnConfig objects
		if configArray, ok := tableColumns.([]Value); ok {
			for _, configVal := range configArray {
				if configMap, ok := configVal.(map[string]Value); ok {
					// Parse table name
					tableVal, hasTable := configMap["table"]
					columnsVal, hasColumns := configMap["columns"]
					if !hasTable || !hasColumns {
						continue
					}
					tableName, ok := tableVal.(string)
					if !ok {
						continue
					}
					// Parse columns array - handle both []Value and *[]Value
					var columnsArray []Value
					columnsFound := false

					if arr, isSlice := columnsVal.([]Value); isSlice {
						columnsArray = arr
						columnsFound = true
					} else if columnsPtr, isPtr := columnsVal.(*[]Value); isPtr {
						columnsArray = *columnsPtr
						columnsFound = true
					}

					if columnsFound {
						var columns []string
						for _, colVal := range columnsArray {
							if colStr, isString := colVal.(string); isString {
								columns = append(columns, colStr)
							}
						}
						tableColumnMap[tableName] = columns
					}
				}
			}
		} else if columnsArray, ok := tableColumns.([]string); ok {
			// Backward compatibility: apply same columns to all tables
			for _, tableName := range tableNames {
				tableColumnMap[tableName] = columnsArray
			}
		}
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
			ID:           streamerID,
			DB:           connObj.DB,
			Listener:     listener,
			TableName:    strings.Join(tableNames, ","),
			Callback:     callback,
			Active:       true,
			Cancel:       cancel,
			Driver:       "postgres",
			TableColumns: tableColumnMap,
		}
	} else {
		// MySQL streamer with optimized CDC
		streamer = &DatabaseStreamer{
			ID:           streamerID,
			DB:           connObj.DB,
			Listener:     nil,
			TableName:    strings.Join(tableNames, ","),
			Callback:     callback,
			Active:       true,
			Cancel:       cancel,
			Driver:       "mysql",
			LastID:       0,
			UseBinlog:    true, // Enable optimized CDC
			ServerID:     1001, // Default server ID for binlog replication
			TableColumns: tableColumnMap,
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

	// Close connection pool
	databaseMutex.Lock()
	pool, exists := databaseConnections[connObj.ID]
	if exists {
		delete(databaseConnections, connObj.ID)
	}
	databaseMutex.Unlock()

	if !exists {
		return Value(map[string]Value{
			"success": false,
			"error":   "Connection not found",
		})
	}

	err := pool.Close()
	if err != nil {
		return Value(map[string]Value{
			"success": false,
			"error":   err.Error(),
		})
	}

	return Value(map[string]Value{
		"success": true,
		"message": "Connection pool closed successfully",
	})
}

// dbConfigurePoolFunc implements the db_configure_pool() built-in function
// db_configure_pool(max_open, max_idle, max_lifetime_minutes, max_idle_minutes) -> config_object
// Example: config = db_configure_pool(50, 10, 10, 2)
func dbConfigurePoolFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 4 {
		panic(typeError(pos, "db_configure_pool() requires 4 arguments: max_open, max_idle, max_lifetime_minutes, max_idle_minutes"))
	}

	maxOpen, ok := args[0].(int)
	if !ok {
		panic(typeError(pos, "db_configure_pool() requires first argument to be an integer (max_open)"))
	}

	maxIdle, ok := args[1].(int)
	if !ok {
		panic(typeError(pos, "db_configure_pool() requires second argument to be an integer (max_idle)"))
	}

	maxLifetimeMin, ok := args[2].(int)
	if !ok {
		panic(typeError(pos, "db_configure_pool() requires third argument to be an integer (max_lifetime_minutes)"))
	}

	maxIdleMin, ok := args[3].(int)
	if !ok {
		panic(typeError(pos, "db_configure_pool() requires fourth argument to be an integer (max_idle_minutes)"))
	}

	config := ConnectionPoolConfig{
		MaxOpenConns:    maxOpen,
		MaxIdleConns:    maxIdle,
		ConnMaxLifetime: time.Duration(maxLifetimeMin) * time.Minute,
		ConnMaxIdleTime: time.Duration(maxIdleMin) * time.Minute,
	}

	return Value(map[string]Value{
		"max_open":         maxOpen,
		"max_idle":         maxIdle,
		"max_lifetime_min": maxLifetimeMin,
		"max_idle_min":     maxIdleMin,
		"config":           &config,
	})
}

// dbConnectWithPoolFunc implements the db_connect_with_pool() built-in function
// db_connect_with_pool(driver, host, port, database, username, password, pool_config) -> connection_object
// Example: conn = db_connect_with_pool("postgres", "localhost", 5432, "mydb", "user", "pass", config)
func dbConnectWithPoolFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 7 {
		panic(typeError(pos, "db_connect_with_pool() requires 7 arguments: driver, host, port, database, username, password, pool_config"))
	}

	driver, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "db_connect_with_pool() requires first argument to be a string (driver)"))
	}

	host, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "db_connect_with_pool() requires second argument to be a string (host)"))
	}

	port, ok := args[2].(int)
	if !ok {
		panic(typeError(pos, "db_connect_with_pool() requires third argument to be an integer (port)"))
	}

	database, ok := args[3].(string)
	if !ok {
		panic(typeError(pos, "db_connect_with_pool() requires fourth argument to be a string (database)"))
	}

	username, ok := args[4].(string)
	if !ok {
		panic(typeError(pos, "db_connect_with_pool() requires fifth argument to be a string (username)"))
	}

	password, ok := args[5].(string)
	if !ok {
		panic(typeError(pos, "db_connect_with_pool() requires sixth argument to be a string (password)"))
	}

	poolConfig, ok := args[6].(*ConnectionPoolConfig)
	if !ok {
		panic(typeError(pos, "db_connect_with_pool() requires seventh argument to be a pool configuration"))
	}

	// Validate driver
	if driver != "postgres" && driver != "mysql" {
		panic(typeError(pos, "db_connect_with_pool() supports only 'postgres' and 'mysql' drivers"))
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

	// Create connection pool with custom configuration
	pool, err := NewConnectionPool(driver, dsn, *poolConfig)
	if err != nil {
		return Value(map[string]Value{
			"success": false,
			"error":   err.Error(),
			"conn":    Value(nil),
		})
	}

	// Generate connection ID
	connID := fmt.Sprintf("%s_%s_%d_%s_%d", driver, host, port, database, time.Now().UnixNano())

	// Store connection pool
	databaseMutex.Lock()
	databaseConnections[connID] = pool
	databaseMutex.Unlock()

	// Create connection object
	connObj := &DatabaseConnection{
		ID:       connID,
		DB:       pool.GetConnection(),
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
		"pool_config": map[string]Value{
			"max_open":         poolConfig.MaxOpenConns,
			"max_idle":         poolConfig.MaxIdleConns,
			"max_lifetime_min": int(poolConfig.ConnMaxLifetime.Minutes()),
			"max_idle_min":     int(poolConfig.ConnMaxIdleTime.Minutes()),
		},
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

		// Build column selection for trigger function
		var newDataExpression, oldDataExpression string
		tableColumns := ds.TableColumns[tableName]
		if len(tableColumns) > 0 {
			// Create JSON object with only selected columns for NEW data
			newColumnSelections := make([]string, len(tableColumns))
			oldColumnSelections := make([]string, len(tableColumns))
			for i, col := range tableColumns {
				newColumnSelections[i] = fmt.Sprintf("'%s', NEW.%s", col, col)
				oldColumnSelections[i] = fmt.Sprintf("'%s', OLD.%s", col, col)
			}
			newDataExpression = fmt.Sprintf("json_build_object(%s)", strings.Join(newColumnSelections, ", "))
			oldDataExpression = fmt.Sprintf("json_build_object(%s)", strings.Join(oldColumnSelections, ", "))
		} else {
			// Include all columns (default behavior)
			newDataExpression = "row_to_json(NEW)"
			oldDataExpression = "row_to_json(OLD)"
		}

		// Setup trigger function if it doesn't exist
		setupTriggerSQL := fmt.Sprintf(`
			CREATE OR REPLACE FUNCTION notify_%s_change()
			RETURNS TRIGGER AS $$
			BEGIN
				PERFORM pg_notify('%s',
					CASE TG_OP
						WHEN 'INSERT' THEN
							json_build_object(
								'table', TG_TABLE_NAME,
								'operation', TG_OP,
								'timestamp', extract(epoch from now()),
								'new_data', %s
							)::text
						WHEN 'UPDATE' THEN
							json_build_object(
								'table', TG_TABLE_NAME,
								'operation', TG_OP,
								'timestamp', extract(epoch from now()),
								'old_data', %s,
								'new_data', %s
							)::text
						WHEN 'DELETE' THEN
							json_build_object(
								'table', TG_TABLE_NAME,
								'operation', TG_OP,
								'timestamp', extract(epoch from now()),
								'old_data', %s
							)::text
				END
				);
				RETURN CASE TG_OP
					WHEN 'DELETE' THEN OLD
					ELSE NEW
				END;
			END;
			$$ LANGUAGE plpgsql;
		`, tableName, channelName, newDataExpression, oldDataExpression, newDataExpression, oldDataExpression)

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
		Logger:   slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelError + 1})), // Silent logger
	}

	// Create binlog syncer
	syncer := replication.NewBinlogSyncer(cfg)

	// Get current binlog position
	binlogPos, err := ds.getCurrentBinlogPosition()
	if err != nil {
		binlogPos = mysql.Position{Name: "", Pos: 4} // Start from beginning if no position found
	}

	// Start streaming from the position
	streamer, err := syncer.StartSync(binlogPos)
	if err != nil {
		log.Printf("Error starting binlog sync: %v", err)
		syncer.Close()
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

	// Setup cleanup when context is done
	go func() {
		<-ctx.Done()
		if ds.Binlog != nil && ds.Binlog.Syncer != nil {
			ds.Binlog.Syncer.Close()
		}
	}()

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
	var binlogDoDB, binlogIgnoreDB, executedGtidSet any

	// Try modern MySQL syntax first (MySQL 8.0+)
	query := "SHOW BINARY LOG STATUS"
	row := ds.DB.QueryRow(query)
	err := row.Scan(&file, &position, &binlogDoDB, &binlogIgnoreDB, &executedGtidSet)
	if err != nil {
		// Fallback to older syntax for MySQL 5.7 and below
		query = "SHOW MASTER STATUS"
		row = ds.DB.QueryRow(query)
		err = row.Scan(&file, &position, &binlogDoDB, &binlogIgnoreDB, &executedGtidSet)
		if err != nil {
			return mysql.Position{}, err
		}
	}

	return mysql.Position{Name: file, Pos: position}, nil
}

// processBinlogEvents processes MySQL binlog events in real-time
func (ds *DatabaseStreamer) processBinlogEvents(interp *interpreter, pos Position, ctx context.Context) {
	errorCount := 0
	maxErrors := 10
	backoffDuration := time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Get next binlog event
			ev, err := ds.Binlog.Streamer.GetEvent(ctx)
			if err != nil {
				errorCount++
				log.Printf("Error getting binlog event (%d/%d): %v", errorCount, maxErrors, err)

				// If too many consecutive errors, exit
				if errorCount >= maxErrors {
					log.Printf("Too many consecutive errors, stopping binlog streaming")
					return
				}

				// Exponential backoff
				select {
				case <-ctx.Done():
					return
				case <-time.After(backoffDuration):
					backoffDuration = backoffDuration * 2
					if backoffDuration > 30*time.Second {
						backoffDuration = 30 * time.Second
					}
				}
				continue
			}

			// Reset error count on successful event
			errorCount = 0
			backoffDuration = time.Second

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

	// Create a set of selected columns for quick lookup
	var selectedColumns map[string]bool
	tableColumns := ds.TableColumns[string(table.Table)]
	if len(tableColumns) > 0 {
		selectedColumns = make(map[string]bool)
		for _, col := range tableColumns {
			selectedColumns[col] = true
		}
	}

	// Map row values to column names
	for i, value := range row {
		if i < len(columns) {
			columnName := columns[i]
			// Only include column if it's selected (or if no columns are specified)
			if len(tableColumns) == 0 || selectedColumns[columnName] {
				result[columnName] = value
			}
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

// Async Processing Implementation
type AsyncOperation struct {
	ID         string
	Query      string
	Args       []Value
	Connection *DatabaseConnection
	Callback   Value // Optional callback function
	StartTime  time.Time
	Status     string // "pending", "running", "completed", "failed"
	Result     Value
	Error      string
	Cancel     context.CancelFunc
	Mutex      sync.RWMutex
}

type AsyncManager struct {
	Operations map[string]*AsyncOperation
	Mutex      sync.RWMutex
	WorkerPool chan struct{} // Semaphore for limiting concurrent operations
}

var (
	asyncManager = &AsyncManager{
		Operations: make(map[string]*AsyncOperation),
		WorkerPool: make(chan struct{}, 10), // Max 10 concurrent async operations
	}
)

// dbExecuteAsyncFunc executes a database query asynchronously
func dbExecuteAsyncFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 {
		return Value(map[string]Value{
			"success": false,
			"error":   "db_execute_async requires at least 2 arguments: connection and query",
		})
	}

	// Extract connection
	connArg := args[0]
	connMap, ok := connArg.(map[string]Value)
	if !ok {
		return Value(map[string]Value{
			"success": false,
			"error":   "Invalid connection argument",
		})
	}

	connValue, exists := connMap["conn"]
	if !exists {
		return Value(map[string]Value{
			"success": false,
			"error":   "Connection object not found",
		})
	}

	conn, ok := connValue.(*DatabaseConnection)
	if !ok {
		return Value(map[string]Value{
			"success": false,
			"error":   "Invalid connection object",
		})
	}

	// Extract query
	query, ok := args[1].(string)
	if !ok {
		return Value(map[string]Value{
			"success": false,
			"error":   "Query must be a string",
		})
	}

	// Extract optional arguments
	var queryArgs []Value
	var callback Value
	if len(args) > 2 {
		if len(args) == 3 {
			// Could be args array or callback function
			if argsArray, ok := args[2].(*[]Value); ok {
				queryArgs = *argsArray
			} else if IsFunction(args[2]) {
				callback = args[2]
			} else {
				return Value(map[string]Value{
					"success": false,
					"error":   "Third argument must be an array of arguments or a callback function",
				})
			}
		} else if len(args) == 4 {
			// Both args and callback
			if argsArray, ok := args[2].(*[]Value); ok {
				queryArgs = *argsArray
			} else {
				return Value(map[string]Value{
					"success": false,
					"error":   "Third argument must be an array of arguments",
				})
			}
			if IsFunction(args[3]) {
				callback = args[3]
			} else {
				return Value(map[string]Value{
					"success": false,
					"error":   "Fourth argument must be a callback function",
				})
			}
		}
	}

	// Generate unique operation ID
	operationID := fmt.Sprintf("async_%d_%d", time.Now().UnixNano(), len(asyncManager.Operations))

	// Create async operation
	ctx, cancel := context.WithCancel(context.Background())
	operation := &AsyncOperation{
		ID:         operationID,
		Query:      query,
		Args:       queryArgs,
		Connection: conn,
		Callback:   callback,
		StartTime:  time.Now(),
		Status:     "pending",
		Cancel:     cancel,
	}

	// Add to manager
	asyncManager.Mutex.Lock()
	asyncManager.Operations[operationID] = operation
	asyncManager.Mutex.Unlock()

	// Start async execution
	go executeAsyncOperation(interp, pos, operation, ctx)

	return Value(map[string]Value{
		"success":      true,
		"operation_id": operationID,
		"status":       "pending",
	})
}

// executeAsyncOperation runs the database operation in a goroutine
func executeAsyncOperation(interp *interpreter, pos Position, op *AsyncOperation, ctx context.Context) {
	// Acquire worker slot
	asyncManager.WorkerPool <- struct{}{}
	defer func() { <-asyncManager.WorkerPool }()

	// Update status to running
	op.Mutex.Lock()
	op.Status = "running"
	op.Mutex.Unlock()

	defer func() {
		if r := recover(); r != nil {
			op.Mutex.Lock()
			op.Status = "failed"
			op.Error = fmt.Sprintf("Panic during execution: %v", r)
			op.Mutex.Unlock()
			log.Printf("Async operation %s panicked: %v", op.ID, r)
		}
	}()

	// Check if operation was cancelled
	select {
	case <-ctx.Done():
		op.Mutex.Lock()
		op.Status = "cancelled"
		op.Error = "Operation was cancelled"
		op.Mutex.Unlock()
		return
	default:
	}

	// Convert query placeholders if needed
	convertedQuery := convertPlaceholders(op.Query, op.Connection.Driver)

	// Convert UDDIN-LANG values to Go any values
	sqlArgs := make([]any, len(op.Args))
	for i, arg := range op.Args {
		sqlArgs[i] = convertToSQLArg(arg)
	}

	// Determine if this is a SELECT query or an execute query
	queryUpper := strings.ToUpper(strings.TrimSpace(convertedQuery))
	isSelect := strings.HasPrefix(queryUpper, "SELECT") || strings.HasPrefix(queryUpper, "WITH")

	var result Value
	if isSelect {
		// Execute as query (SELECT)
		rows, err := op.Connection.DB.QueryContext(ctx, convertedQuery, sqlArgs...)
		if err != nil {
			op.Mutex.Lock()
			op.Status = "failed"
			op.Error = err.Error()
			op.Mutex.Unlock()

			// Call callback with error if provided
			if op.Callback != nil && IsFunction(op.Callback) {
				go callAsyncCallback(interp, pos, op.Callback, Value(map[string]Value{
					"success":      false,
					"error":        err.Error(),
					"operation_id": op.ID,
				}))
			}
			return
		}
		defer rows.Close()

		// Process results
		columns, err := rows.Columns()
		if err != nil {
			op.Mutex.Lock()
			op.Status = "failed"
			op.Error = err.Error()
			op.Mutex.Unlock()
			return
		}

		var results []Value
		for rows.Next() {
			// Check for cancellation
			select {
			case <-ctx.Done():
				op.Mutex.Lock()
				op.Status = "cancelled"
				op.Error = "Operation was cancelled"
				op.Mutex.Unlock()
				return
			default:
			}

			values := make([]any, len(columns))
			valuePtrs := make([]any, len(columns))
			for i := range columns {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				op.Mutex.Lock()
				op.Status = "failed"
				op.Error = err.Error()
				op.Mutex.Unlock()
				return
			}

			row := make(map[string]Value)
			for i, col := range columns {
				row[col] = convertSQLValue(values[i])
			}
			results = append(results, Value(row))
		}

		if err := rows.Err(); err != nil {
			op.Mutex.Lock()
			op.Status = "failed"
			op.Error = err.Error()
			op.Mutex.Unlock()
			return
		}

		// Operation completed successfully
		result = Value(map[string]Value{
			"success":      true,
			"data":         Value(&results),
			"count":        len(results),
			"operation_id": op.ID,
		})
	} else {
		// Execute as command (INSERT, UPDATE, DELETE)
		execResult, err := op.Connection.DB.ExecContext(ctx, convertedQuery, sqlArgs...)
		if err != nil {
			op.Mutex.Lock()
			op.Status = "failed"
			op.Error = err.Error()
			op.Mutex.Unlock()

			// Call callback with error if provided
			if op.Callback != nil && IsFunction(op.Callback) {
				go callAsyncCallback(interp, pos, op.Callback, Value(map[string]Value{
					"success":      false,
					"error":        err.Error(),
					"operation_id": op.ID,
				}))
			}
			return
		}

		rowsAffected, _ := execResult.RowsAffected()
		lastInsertId, _ := execResult.LastInsertId()

		// Operation completed successfully
		result = Value(map[string]Value{
			"success":        true,
			"rows_affected":  rowsAffected,
			"last_insert_id": lastInsertId,
			"operation_id":   op.ID,
		})
	}

	op.Mutex.Lock()
	op.Status = "completed"
	op.Result = result
	op.Mutex.Unlock()

	// Call callback if provided
	if op.Callback != nil && IsFunction(op.Callback) {
		go callAsyncCallback(interp, pos, op.Callback, result)
	}
}

// callAsyncCallback calls the callback function for async operations
func callAsyncCallback(interp *interpreter, pos Position, callback Value, result Value) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Async callback error: %v", r)
		}
	}()

	if fn, ok := callback.(functionType); ok {
		fn.call(interp, pos, []Value{result})
	}
}

// dbGetAsyncStatusFunc gets the status of an async operation
func dbGetAsyncStatusFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		return Value(map[string]Value{
			"success": false,
			"error":   "db_get_async_status requires 1 argument: operation_id",
		})
	}

	operationID, ok := args[0].(string)
	if !ok {
		return Value(map[string]Value{
			"success": false,
			"error":   "Operation ID must be a string",
		})
	}

	asyncManager.Mutex.RLock()
	operation, exists := asyncManager.Operations[operationID]
	asyncManager.Mutex.RUnlock()

	if !exists {
		return Value(map[string]Value{
			"success": false,
			"error":   "Operation not found",
		})
	}

	operation.Mutex.RLock()
	defer operation.Mutex.RUnlock()

	response := map[string]Value{
		"success":      true,
		"operation_id": operationID,
		"status":       operation.Status,
		"start_time":   operation.StartTime.Unix(),
		"duration":     time.Since(operation.StartTime).Milliseconds(),
	}

	if operation.Status == "completed" && operation.Result != nil {
		response["result"] = operation.Result
	}

	if operation.Status == "failed" && operation.Error != "" {
		response["error"] = operation.Error
	}

	return Value(response)
}

// dbCancelAsyncFunc cancels an async operation
func dbCancelAsyncFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) != 1 {
		return Value(map[string]Value{
			"success": false,
			"error":   "db_cancel_async requires 1 argument: operation_id",
		})
	}

	operationID, ok := args[0].(string)
	if !ok {
		return Value(map[string]Value{
			"success": false,
			"error":   "Operation ID must be a string",
		})
	}

	asyncManager.Mutex.RLock()
	operation, exists := asyncManager.Operations[operationID]
	asyncManager.Mutex.RUnlock()

	if !exists {
		return Value(map[string]Value{
			"success": false,
			"error":   "Operation not found",
		})
	}

	// Cancel the operation
	operation.Cancel()

	return Value(map[string]Value{
		"success":      true,
		"operation_id": operationID,
		"message":      "Operation cancellation requested",
	})
}

// dbListAsyncOperationsFunc lists all async operations
func dbListAsyncOperationsFunc(interp *interpreter, pos Position, args []Value) Value {
	asyncManager.Mutex.RLock()
	defer asyncManager.Mutex.RUnlock()

	var operations []Value
	for id, op := range asyncManager.Operations {
		op.Mutex.RLock()
		opInfo := map[string]Value{
			"operation_id": id,
			"status":       op.Status,
			"start_time":   op.StartTime.Unix(),
			"duration":     time.Since(op.StartTime).Milliseconds(),
			"query":        op.Query,
		}
		if op.Error != "" {
			opInfo["error"] = op.Error
		}
		op.Mutex.RUnlock()
		operations = append(operations, Value(opInfo))
	}

	return Value(map[string]Value{
		"success":    true,
		"operations": Value(&operations),
		"count":      len(operations),
	})
}

// dbCleanupAsyncOperationsFunc cleans up completed/failed async operations
func dbCleanupAsyncOperationsFunc(interp *interpreter, pos Position, args []Value) Value {
	// Optional: max_age parameter in seconds (default: 3600 = 1 hour)
	maxAge := int64(3600)
	if len(args) > 0 {
		if age, ok := args[0].(int); ok {
			maxAge = int64(age)
		}
	}

	cutoffTime := time.Now().Add(-time.Duration(maxAge) * time.Second)
	cleanedCount := 0

	asyncManager.Mutex.Lock()
	defer asyncManager.Mutex.Unlock()

	for id, op := range asyncManager.Operations {
		op.Mutex.RLock()
		shouldCleanup := (op.Status == "completed" || op.Status == "failed" || op.Status == "cancelled") &&
			op.StartTime.Before(cutoffTime)
		op.Mutex.RUnlock()

		if shouldCleanup {
			// Cancel if still running (shouldn't happen, but safety first)
			op.Cancel()
			delete(asyncManager.Operations, id)
			cleanedCount++
		}
	}

	return Value(map[string]Value{
		"success":       true,
		"cleaned_count": cleanedCount,
		"remaining":     len(asyncManager.Operations),
	})
}
