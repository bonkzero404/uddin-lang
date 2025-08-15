# MySQL Real-time Streaming Implementation

This document describes the implementation of real-time streaming support for MySQL in UddinLang's `stream_tables` function.

## Overview

The `stream_tables` function now supports both PostgreSQL and MySQL databases:

- **PostgreSQL**: Uses native `LISTEN/NOTIFY` mechanism for true real-time streaming
- **MySQL**: Uses trigger-based CDC (Change Data Capture) with polling approach

## MySQL Implementation Details

### Architecture

The MySQL streaming implementation uses a trigger-based CDC approach similar to Debezium:

1. **Change Log Table**: Creates a unified change log table (`uddin_changes_log`) to capture all changes
2. **Database Triggers**: Creates INSERT, UPDATE, DELETE triggers on monitored tables
3. **Polling Mechanism**: Polls the change log table every 500ms for new changes
4. **JSON Payload**: Stores change data in JSON format for consistency with PostgreSQL

### Change Log Table Structure

```sql
CREATE TABLE uddin_changes_log (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    table_name VARCHAR(255) NOT NULL,
    operation ENUM('INSERT', 'UPDATE', 'DELETE') NOT NULL,
    data_before JSON,
    data_after JSON,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_table_id (table_name, id),
    INDEX idx_timestamp (timestamp)
)
```

### Trigger Creation

For each monitored table, three triggers are created:

1. **INSERT Trigger**: Captures new records
2. **UPDATE Trigger**: Captures both old and new values
3. **DELETE Trigger**: Captures deleted records

### Code Implementation

Key components added to `database_functions.go`:

#### DatabaseStreamer Structure
```go
type DatabaseStreamer struct {
    ID        string
    DB        *sql.DB
    Listener  *pq.Listener  // PostgreSQL only
    TableName string
    Callback  Value
    Active    bool
    Cancel    context.CancelFunc
    Mutex     sync.RWMutex
    Driver    string         // "postgres" or "mysql"
    LastID    int64         // For MySQL polling
    BinlogPos string        // For future binary log support
}
```

#### Key Methods

1. **setupMySQLStreaming**: Sets up triggers and starts polling for single table
2. **setupMySQLMultiTableStreaming**: Sets up triggers for multiple tables
3. **pollMySQLChanges**: Polls change log for single table
4. **pollMySQLMultiTableChanges**: Polls unified change log for multiple tables
5. **getColumnList**: Retrieves table column information
6. **buildJSONObjectColumns**: Builds JSON_OBJECT parameters for triggers

## Usage Examples

### Single Table Streaming

```uddin
// Connect to MySQL
conn_result = db_connect("mysql", "localhost", 3306, "testdb", "user", "password")
conn = conn_result.conn

// Define callback function
fun onStreamData(channel, payload):
    print("Change detected: " + payload)
end

// Start streaming
stream_tables(conn, "users", onStreamData)
```

### Multi-Table Streaming

```uddin
// Stream multiple tables
stream_tables(conn, ["users", "orders", "products"], onStreamData)
```

## Performance Considerations

### Advantages
- **Consistency**: Same API for both PostgreSQL and MySQL
- **Reliability**: Trigger-based approach ensures no missed changes
- **JSON Format**: Consistent data format across databases
- **Multi-table Support**: Unified change log for multiple tables

### Limitations
- **Polling Overhead**: 500ms polling interval (configurable)
- **Trigger Performance**: Database triggers add overhead to DML operations
- **Storage Overhead**: Change log table grows over time
- **Not True Real-time**: Small delay due to polling interval

### Optimization Recommendations

1. **Index Optimization**: Ensure proper indexing on change log table
2. **Cleanup Strategy**: Implement periodic cleanup of old change log entries
3. **Polling Interval**: Adjust polling interval based on requirements
4. **Connection Pooling**: Use connection pooling for high-throughput scenarios

## Comparison with PostgreSQL

| Feature | PostgreSQL | MySQL |
|---------|------------|-------|
| Mechanism | LISTEN/NOTIFY | Trigger + Polling |
| Real-time | True (instant) | Near real-time (500ms delay) |
| Performance Impact | Minimal | Moderate (triggers) |
| Setup Complexity | Low | Medium |
| Storage Overhead | None | Change log table |
| Scalability | High | Medium |

## Future Enhancements

1. **Binary Log Support**: Implement true log-based CDC using MySQL binary logs
2. **Configurable Polling**: Make polling interval configurable
3. **Cleanup Automation**: Automatic cleanup of old change log entries
4. **Performance Monitoring**: Add metrics for trigger and polling performance
5. **Batch Processing**: Process multiple changes in batches for better performance

## Testing

Example files are provided:
- `examples/database/mysql_stream_listener.din`: Single table streaming
- `examples/database/mysql_multi_table_stream_listener.din`: Multi-table streaming

To test:
1. Set up MySQL database with proper credentials
2. Run the example files
3. Perform INSERT/UPDATE/DELETE operations on monitored tables
4. Observe real-time notifications in the UddinLang application

## Conclusion

The MySQL streaming implementation provides a robust, trigger-based CDC solution that maintains API consistency with PostgreSQL while offering reliable change capture for MySQL databases. While not as performant as PostgreSQL's native LISTEN/NOTIFY, it provides a practical solution for real-time data streaming in MySQL environments.