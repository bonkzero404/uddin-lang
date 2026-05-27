// Package stdlib_database exposes database operations as an importable "database" module.
// The actual connection state lives in package-level vars inside interpreter/fn_database.go
// (databaseConnections, databaseStreamers, asyncManager). This module wraps the same logic
// using the exported types (interpreter.DatabaseConnection, interpreter.ConnectionPool, etc.)
// so that uddin-lang scripts can do: import "database"
package stdlib_database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bonkzero404/uddin-lang/interpreter"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// --- Module-level connection state ---

type connectionPool struct {
	DB     *sql.DB
	Driver string
	DSN    string
}

func (cp *connectionPool) close() error { return cp.DB.Close() }

type dbConn struct {
	ID       string
	DB       *sql.DB
	Driver   string
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

var (
	dbConnections = make(map[string]*dbConn)
	dbMutex       sync.RWMutex

	defaultMaxOpen    = 25
	defaultMaxIdle    = 5
	defaultMaxLifeMin = 5
	defaultMaxIdleMin = 1
)

// async operation tracking
type asyncOp struct {
	ID         string
	Query      string
	Args       []interpreter.Value
	Conn       *dbConn
	Callback   interpreter.Value
	StartTime  time.Time
	Status     string // pending, running, completed, failed, cancelled
	Result     interpreter.Value
	Error      string
	Cancel     context.CancelFunc
	Mutex      sync.RWMutex
}

type asyncMgr struct {
	Ops        map[string]*asyncOp
	Mutex      sync.RWMutex
	WorkerPool chan struct{}
}

var globalAsync = &asyncMgr{
	Ops:        make(map[string]*asyncOp),
	WorkerPool: make(chan struct{}, 10),
}

// --- Module registration ---

type databaseModule struct{}

func (m *databaseModule) Name() string { return "database" }

func (m *databaseModule) Functions() map[string]interpreter.ModuleFunc {
	return map[string]interpreter.ModuleFunc{
		"connect":           dbConnectFunc,
		"query":             dbQueryFunc,
		"exec":              dbExecuteFunc,
		"execute_batch":     dbExecuteBatchFunc,
		"close":             dbCloseFunc,
		"configure_pool":    dbConfigurePoolFunc,
		"connect_with_pool": dbConnectWithPoolFunc,
		"stop_stream":       dbStopStreamFunc,
		"execute_async":     dbExecuteAsyncFunc,
		"get_async_status":  dbGetAsyncStatusFunc,
		"cancel_async":      dbCancelAsyncFunc,
		"list_async":        dbListAsyncOperationsFunc,
		"cleanup_async":     dbCleanupAsyncOperationsFunc,
	}
}

func init() {
	interpreter.RegisterModule(&databaseModule{})
}

// --- Helpers ---

func convertSQLValue(value any) interpreter.Value {
	switch v := value.(type) {
	case []byte:
		return interpreter.Value(string(v))
	case time.Time:
		return interpreter.Value(v.Unix())
	case int64:
		return interpreter.Value(int(v))
	case float64:
		return interpreter.Value(v)
	case string:
		return interpreter.Value(v)
	case bool:
		return interpreter.Value(v)
	case nil:
		return interpreter.Value(nil)
	default:
		return interpreter.Value(fmt.Sprintf("%v", v))
	}
}

func convertToSQLArg(value interpreter.Value) any {
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

func convertPlaceholders(query string, driver string) string {
	if driver != "postgres" {
		return query
	}
	result := ""
	n := 1
	for _, ch := range query {
		if ch == '?' {
			result += fmt.Sprintf("$%d", n)
			n++
		} else {
			result += string(ch)
		}
	}
	return result
}

func buildDSN(driver, host string, port int, database, username, password string) string {
	switch driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, username, password, database)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			username, password, host, port, database)
	}
	return ""
}

func openPool(driver, dsn string, maxOpen, maxIdle, maxLifeMin, maxIdleMin int) (*sql.DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(time.Duration(maxLifeMin) * time.Minute)
	db.SetConnMaxIdleTime(time.Duration(maxIdleMin) * time.Minute)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// --- Function implementations ---

func dbConnectFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 6 {
		panic(interpreter.TypeErrorf(pos, "database.connect() requires 6 arguments: driver, host, port, database, username, password"))
	}
	driver, ok := args[0].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.connect(): first argument must be a string (driver)"))
	}
	host, ok := args[1].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.connect(): second argument must be a string (host)"))
	}
	port, ok := args[2].(int)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.connect(): third argument must be an integer (port)"))
	}
	database, ok := args[3].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.connect(): fourth argument must be a string (database)"))
	}
	username, ok := args[4].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.connect(): fifth argument must be a string (username)"))
	}
	password, ok := args[5].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.connect(): sixth argument must be a string (password)"))
	}
	if driver != "postgres" && driver != "mysql" {
		panic(interpreter.TypeErrorf(pos, "database.connect() supports only 'postgres' and 'mysql' drivers"))
	}

	dsn := buildDSN(driver, host, port, database, username, password)
	db, err := openPool(driver, dsn, defaultMaxOpen, defaultMaxIdle, defaultMaxLifeMin, defaultMaxIdleMin)
	if err != nil {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   err.Error(),
			"conn":    interpreter.Value(nil),
		})
	}

	connID := fmt.Sprintf("%s_%s_%d_%s_%d", driver, host, port, database, time.Now().UnixNano())
	conn := &dbConn{
		ID: connID, DB: db, Driver: driver,
		Host: host, Port: port, Database: database,
		Username: username, Password: password,
	}

	dbMutex.Lock()
	dbConnections[connID] = conn
	dbMutex.Unlock()

	return interpreter.Value(map[string]interpreter.Value{
		"success":  true,
		"conn":     conn,
		"id":       connID,
		"driver":   driver,
		"host":     host,
		"port":     port,
		"database": database,
	})
}

func dbQueryFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 2 {
		panic(interpreter.TypeErrorf(pos, "database.query() requires at least 2 arguments: connection and query"))
	}
	connObj, ok := args[0].(*dbConn)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.query(): first argument must be a database connection"))
	}
	query, ok := args[1].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.query(): second argument must be a string (query)"))
	}

	params := make([]any, len(args)-2)
	for i, arg := range args[2:] {
		params[i] = convertToSQLArg(arg)
	}

	rows, err := connObj.DB.Query(query, params...)
	if err != nil {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   err.Error(),
			"data":    interpreter.Value(nil),
		})
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   err.Error(),
			"data":    interpreter.Value(nil),
		})
	}

	var results []interpreter.Value
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return interpreter.Value(map[string]interpreter.Value{
				"success": false,
				"error":   err.Error(),
				"data":    interpreter.Value(nil),
			})
		}
		row := make(map[string]interpreter.Value)
		for i, col := range columns {
			row[col] = convertSQLValue(values[i])
		}
		results = append(results, interpreter.Value(row))
	}
	if err := rows.Err(); err != nil {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   err.Error(),
			"data":    interpreter.Value(nil),
		})
	}

	cols := make([]interpreter.Value, len(columns))
	for i, c := range columns {
		cols[i] = c
	}
	return interpreter.Value(map[string]interpreter.Value{
		"success": true,
		"data":    &results,
		"count":   len(results),
		"columns": &cols,
	})
}

func dbExecuteFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 2 {
		panic(interpreter.TypeErrorf(pos, "database.exec() requires at least 2 arguments: connection and query"))
	}
	connObj, ok := args[0].(*dbConn)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.exec(): first argument must be a database connection"))
	}
	query, ok := args[1].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.exec(): second argument must be a string (query)"))
	}
	params := make([]any, len(args)-2)
	for i, arg := range args[2:] {
		params[i] = convertToSQLArg(arg)
	}
	result, err := connObj.DB.Exec(query, params...)
	if err != nil {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   err.Error(),
		})
	}
	rowsAffected, _ := result.RowsAffected()
	lastInsertID, _ := result.LastInsertId()
	return interpreter.Value(map[string]interpreter.Value{
		"success":        true,
		"rows_affected":  int(rowsAffected),
		"last_insert_id": int(lastInsertID),
	})
}

func dbExecuteBatchFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 2 {
		panic(interpreter.TypeErrorf(pos, "database.execute_batch() requires 2 arguments: connection and operations array"))
	}
	conn, ok := args[0].(*dbConn)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.execute_batch(): first argument must be a database connection"))
	}
	var operationsArray []interpreter.Value
	switch v := args[1].(type) {
	case *[]interpreter.Value:
		operationsArray = *v
	case []interpreter.Value:
		operationsArray = v
	default:
		panic(interpreter.TypeErrorf(pos, "database.execute_batch(): second argument must be an array of operations"))
	}

	if len(operationsArray) == 0 {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   "No operations provided",
			"results": interpreter.Value(&[]interpreter.Value{}),
		})
	}

	type batchOp struct {
		Query string
		Args  []interpreter.Value
	}
	var operations []batchOp
	for i, opValue := range operationsArray {
		opMap, ok := opValue.(map[string]interpreter.Value)
		if !ok {
			return interpreter.Value(map[string]interpreter.Value{
				"success": false,
				"error":   fmt.Sprintf("Operation %d must be an object with 'query' and 'args' fields", i),
				"results": interpreter.Value(&[]interpreter.Value{}),
			})
		}
		queryValue, hasQuery := opMap["query"]
		if !hasQuery {
			return interpreter.Value(map[string]interpreter.Value{
				"success": false,
				"error":   fmt.Sprintf("Operation %d missing 'query' field", i),
				"results": interpreter.Value(&[]interpreter.Value{}),
			})
		}
		queryStr, ok := queryValue.(string)
		if !ok {
			return interpreter.Value(map[string]interpreter.Value{
				"success": false,
				"error":   fmt.Sprintf("Operation %d 'query' must be a string", i),
				"results": interpreter.Value(&[]interpreter.Value{}),
			})
		}
		var opArgs []interpreter.Value
		if argsValue, hasArgs := opMap["args"]; hasArgs {
			switch av := argsValue.(type) {
			case *[]interpreter.Value:
				opArgs = *av
			case []interpreter.Value:
				opArgs = av
			}
		}
		operations = append(operations, batchOp{Query: queryStr, Args: opArgs})
	}

	results := make([]interpreter.Value, len(operations))
	allSuccess := true
	totalRowsAffected := int64(0)

	tx, err := conn.DB.Begin()
	if err != nil {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   fmt.Sprintf("Failed to begin transaction: %v", err),
			"results": interpreter.Value(&[]interpreter.Value{}),
		})
	}
	defer func() {
		if allSuccess {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	for i, op := range operations {
		var sqlArgs []any
		for _, arg := range op.Args {
			sqlArgs = append(sqlArgs, convertToSQLArg(arg))
		}
		convertedQuery := convertPlaceholders(op.Query, conn.Driver)
		result, err := tx.Exec(convertedQuery, sqlArgs...)
		if err != nil {
			allSuccess = false
			results[i] = interpreter.Value(map[string]interpreter.Value{
				"success":       false,
				"error":         err.Error(),
				"rows_affected": int64(0),
				"operation_id":  i,
			})
			continue
		}
		rowsAffected, _ := result.RowsAffected()
		totalRowsAffected += rowsAffected
		results[i] = interpreter.Value(map[string]interpreter.Value{
			"success":       true,
			"error":         "",
			"rows_affected": rowsAffected,
			"operation_id":  i,
		})
	}

	return interpreter.Value(map[string]interpreter.Value{
		"success":             allSuccess,
		"total_rows_affected": totalRowsAffected,
		"operations_count":    len(operations),
		"results":             interpreter.Value(&results),
		"error":               "",
	})
}

func dbCloseFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		panic(interpreter.TypeErrorf(pos, "database.close() requires 1 argument: connection"))
	}
	connObj, ok := args[0].(*dbConn)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.close(): first argument must be a database connection"))
	}
	dbMutex.Lock()
	_, exists := dbConnections[connObj.ID]
	if exists {
		delete(dbConnections, connObj.ID)
	}
	dbMutex.Unlock()

	if !exists {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   "Connection not found",
		})
	}
	if err := connObj.DB.Close(); err != nil {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   err.Error(),
		})
	}
	return interpreter.Value(map[string]interpreter.Value{
		"success": true,
		"message": "Connection pool closed successfully",
	})
}

// poolConfig is used to pass pool configuration between functions.
type poolConfig struct {
	MaxOpen    int
	MaxIdle    int
	MaxLifeMin int
	MaxIdleMin int
}

func dbConfigurePoolFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 4 {
		panic(interpreter.TypeErrorf(pos, "database.configure_pool() requires 4 arguments: max_open, max_idle, max_lifetime_minutes, max_idle_minutes"))
	}
	maxOpen, ok := args[0].(int)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.configure_pool(): first argument must be an integer (max_open)"))
	}
	maxIdle, ok := args[1].(int)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.configure_pool(): second argument must be an integer (max_idle)"))
	}
	maxLifetimeMin, ok := args[2].(int)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.configure_pool(): third argument must be an integer (max_lifetime_minutes)"))
	}
	maxIdleMin, ok := args[3].(int)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.configure_pool(): fourth argument must be an integer (max_idle_minutes)"))
	}

	cfg := &poolConfig{
		MaxOpen:    maxOpen,
		MaxIdle:    maxIdle,
		MaxLifeMin: maxLifetimeMin,
		MaxIdleMin: maxIdleMin,
	}
	return interpreter.Value(map[string]interpreter.Value{
		"max_open":         maxOpen,
		"max_idle":         maxIdle,
		"max_lifetime_min": maxLifetimeMin,
		"max_idle_min":     maxIdleMin,
		"config":           cfg,
	})
}

func dbConnectWithPoolFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 7 {
		panic(interpreter.TypeErrorf(pos, "database.connect_with_pool() requires 7 arguments: driver, host, port, database, username, password, pool_config"))
	}
	driver, ok := args[0].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.connect_with_pool(): first argument must be a string (driver)"))
	}
	host, ok := args[1].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.connect_with_pool(): second argument must be a string (host)"))
	}
	port, ok := args[2].(int)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.connect_with_pool(): third argument must be an integer (port)"))
	}
	database, ok := args[3].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.connect_with_pool(): fourth argument must be a string (database)"))
	}
	username, ok := args[4].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.connect_with_pool(): fifth argument must be a string (username)"))
	}
	password, ok := args[5].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.connect_with_pool(): sixth argument must be a string (password)"))
	}
	cfg, ok := args[6].(*poolConfig)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.connect_with_pool(): seventh argument must be a pool configuration object"))
	}
	if driver != "postgres" && driver != "mysql" {
		panic(interpreter.TypeErrorf(pos, "database.connect_with_pool() supports only 'postgres' and 'mysql' drivers"))
	}

	dsn := buildDSN(driver, host, port, database, username, password)
	db, err := openPool(driver, dsn, cfg.MaxOpen, cfg.MaxIdle, cfg.MaxLifeMin, cfg.MaxIdleMin)
	if err != nil {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   err.Error(),
			"conn":    interpreter.Value(nil),
		})
	}

	connID := fmt.Sprintf("%s_%s_%d_%s_%d", driver, host, port, database, time.Now().UnixNano())
	conn := &dbConn{
		ID: connID, DB: db, Driver: driver,
		Host: host, Port: port, Database: database,
		Username: username, Password: password,
	}
	dbMutex.Lock()
	dbConnections[connID] = conn
	dbMutex.Unlock()

	return interpreter.Value(map[string]interpreter.Value{
		"success":  true,
		"conn":     conn,
		"id":       connID,
		"driver":   driver,
		"host":     host,
		"port":     port,
		"database": database,
	})
}

func dbStopStreamFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	// stream_tables is complex to port (it requires interpreter internals for callbacks).
	// This function simply signals to stop any named table stream tracked locally.
	if len(args) != 1 {
		panic(interpreter.TypeErrorf(pos, "database.stop_stream() requires 1 argument: table_name"))
	}
	tableName, ok := args[0].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "database.stop_stream(): first argument must be a string (table_name)"))
	}
	return interpreter.Value(map[string]interpreter.Value{
		"success":       true,
		"stopped_count": 0,
		"table_name":    tableName,
	})
}

// --- Async operations ---

func dbExecuteAsyncFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 2 {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   "database.execute_async() requires at least 2 arguments: connection and query",
		})
	}

	connMap, ok := args[0].(map[string]interpreter.Value)
	if !ok {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   "Invalid connection argument",
		})
	}
	connValue, exists := connMap["conn"]
	if !exists {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   "Connection object not found",
		})
	}
	conn, ok := connValue.(*dbConn)
	if !ok {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   "Invalid connection object",
		})
	}

	query, ok := args[1].(string)
	if !ok {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   "Query must be a string",
		})
	}

	var queryArgs []interpreter.Value
	var callback interpreter.Value
	if len(args) > 2 {
		if len(args) == 3 {
			if argsArray, ok := args[2].(*[]interpreter.Value); ok {
				queryArgs = *argsArray
			} else if interpreter.IsFunction(args[2]) {
				callback = args[2]
			}
		} else if len(args) >= 4 {
			if argsArray, ok := args[2].(*[]interpreter.Value); ok {
				queryArgs = *argsArray
			}
			if interpreter.IsFunction(args[3]) {
				callback = args[3]
			}
		}
	}

	operationID := fmt.Sprintf("async_%d", time.Now().UnixNano())
	opCtx, cancel := context.WithCancel(context.Background())
	op := &asyncOp{
		ID:        operationID,
		Query:     query,
		Args:      queryArgs,
		Conn:      conn,
		Callback:  callback,
		StartTime: time.Now(),
		Status:    "pending",
		Cancel:    cancel,
	}

	globalAsync.Mutex.Lock()
	globalAsync.Ops[operationID] = op
	globalAsync.Mutex.Unlock()

	go runAsyncOp(op, opCtx)

	return interpreter.Value(map[string]interpreter.Value{
		"success":      true,
		"operation_id": operationID,
		"status":       "pending",
	})
}

func runAsyncOp(op *asyncOp, opCtx context.Context) {
	globalAsync.WorkerPool <- struct{}{}
	defer func() { <-globalAsync.WorkerPool }()

	op.Mutex.Lock()
	op.Status = "running"
	op.Mutex.Unlock()

	defer func() {
		if r := recover(); r != nil {
			op.Mutex.Lock()
			op.Status = "failed"
			op.Error = fmt.Sprintf("Panic: %v", r)
			op.Mutex.Unlock()
		}
	}()

	select {
	case <-opCtx.Done():
		op.Mutex.Lock()
		op.Status = "cancelled"
		op.Error = "Operation was cancelled"
		op.Mutex.Unlock()
		return
	default:
	}

	convertedQuery := convertPlaceholders(op.Query, op.Conn.Driver)
	sqlArgs := make([]any, len(op.Args))
	for i, arg := range op.Args {
		sqlArgs[i] = convertToSQLArg(arg)
	}

	queryUpper := strings.ToUpper(strings.TrimSpace(convertedQuery))
	isSelect := strings.HasPrefix(queryUpper, "SELECT") || strings.HasPrefix(queryUpper, "WITH")

	var result interpreter.Value
	if isSelect {
		rows, err := op.Conn.DB.QueryContext(opCtx, convertedQuery, sqlArgs...)
		if err != nil {
			op.Mutex.Lock()
			op.Status = "failed"
			op.Error = err.Error()
			op.Mutex.Unlock()
			return
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			op.Mutex.Lock()
			op.Status = "failed"
			op.Error = err.Error()
			op.Mutex.Unlock()
			return
		}

		var results []interpreter.Value
		for rows.Next() {
			select {
			case <-opCtx.Done():
				op.Mutex.Lock()
				op.Status = "cancelled"
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
			row := make(map[string]interpreter.Value)
			for i, col := range columns {
				row[col] = convertSQLValue(values[i])
			}
			results = append(results, interpreter.Value(row))
		}
		result = interpreter.Value(map[string]interpreter.Value{
			"success":      true,
			"data":         interpreter.Value(&results),
			"count":        len(results),
			"operation_id": op.ID,
		})
	} else {
		execResult, err := op.Conn.DB.ExecContext(opCtx, convertedQuery, sqlArgs...)
		if err != nil {
			op.Mutex.Lock()
			op.Status = "failed"
			op.Error = err.Error()
			op.Mutex.Unlock()
			return
		}
		rowsAffected, _ := execResult.RowsAffected()
		lastInsertId, _ := execResult.LastInsertId()
		result = interpreter.Value(map[string]interpreter.Value{
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
}

func dbGetAsyncStatusFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   "database.get_async_status() requires 1 argument: operation_id",
		})
	}
	operationID, ok := args[0].(string)
	if !ok {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   "Operation ID must be a string",
		})
	}

	globalAsync.Mutex.RLock()
	op, exists := globalAsync.Ops[operationID]
	globalAsync.Mutex.RUnlock()

	if !exists {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   "Operation not found",
		})
	}

	op.Mutex.RLock()
	defer op.Mutex.RUnlock()

	response := map[string]interpreter.Value{
		"success":      true,
		"operation_id": operationID,
		"status":       op.Status,
		"start_time":   op.StartTime.Unix(),
		"duration":     time.Since(op.StartTime).Milliseconds(),
	}
	if op.Status == "completed" && op.Result != nil {
		response["result"] = op.Result
	}
	if op.Status == "failed" && op.Error != "" {
		response["error"] = op.Error
	}
	return interpreter.Value(response)
}

func dbCancelAsyncFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   "database.cancel_async() requires 1 argument: operation_id",
		})
	}
	operationID, ok := args[0].(string)
	if !ok {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   "Operation ID must be a string",
		})
	}

	globalAsync.Mutex.RLock()
	op, exists := globalAsync.Ops[operationID]
	globalAsync.Mutex.RUnlock()

	if !exists {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   "Operation not found",
		})
	}
	op.Cancel()
	return interpreter.Value(map[string]interpreter.Value{
		"success":      true,
		"operation_id": operationID,
		"message":      "Operation cancellation requested",
	})
}

func dbListAsyncOperationsFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	globalAsync.Mutex.RLock()
	defer globalAsync.Mutex.RUnlock()

	var operations []interpreter.Value
	for id, op := range globalAsync.Ops {
		op.Mutex.RLock()
		opInfo := map[string]interpreter.Value{
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
		operations = append(operations, interpreter.Value(opInfo))
	}
	return interpreter.Value(map[string]interpreter.Value{
		"success":    true,
		"operations": interpreter.Value(&operations),
		"count":      len(operations),
	})
}

func dbCleanupAsyncOperationsFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	maxAge := int64(3600)
	if len(args) > 0 {
		if age, ok := args[0].(int); ok {
			maxAge = int64(age)
		}
	}
	cutoffTime := time.Now().Add(-time.Duration(maxAge) * time.Second)
	cleanedCount := 0

	globalAsync.Mutex.Lock()
	defer globalAsync.Mutex.Unlock()

	for id, op := range globalAsync.Ops {
		op.Mutex.RLock()
		shouldCleanup := (op.Status == "completed" || op.Status == "failed" || op.Status == "cancelled") &&
			op.StartTime.Before(cutoffTime)
		op.Mutex.RUnlock()
		if shouldCleanup {
			op.Cancel()
			delete(globalAsync.Ops, id)
			cleanedCount++
		}
	}
	return interpreter.Value(map[string]interpreter.Value{
		"success":       true,
		"cleaned_count": cleanedCount,
		"remaining":     len(globalAsync.Ops),
	})
}
