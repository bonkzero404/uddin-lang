---
sidebar_position: 5
---

# Database Integration

UDDIN-LANG provides powerful database integration capabilities with support for PostgreSQL and MySQL. The database system includes not only standard query operations but also real-time data change listening for building reactive applications.

## Supported Databases

- **PostgreSQL** - Full support with advanced features
- **MySQL** - Complete compatibility with MySQL/MariaDB
- **Real-time Listening** - Monitor data changes in real-time
- **High Performance** - Optimized connection pooling and query execution

## Database Functions

### Connection Management

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `db_connect(driver, host, port, database, username, password)` | Connect to database | `conn = db_connect("postgres", "localhost", 5432, "mydb", "user", "pass")` | object |
| `db_connect_with_pool(driver, host, port, database, username, password, pool_size)` | Connect with connection pool | `conn = db_connect_with_pool("postgres", "localhost", 5432, "mydb", "user", "pass", 10)` | object |
| `db_configure_pool(connection, max_open, max_idle, max_lifetime)` | Configure connection pool | `db_configure_pool(conn, 20, 10, 3600)` | object |
| `db_close(connection)` | Close database connection | `db_close(conn)` | void |

### Query Operations

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `db_query(connection, query, params...)` | Execute SELECT query | `result = db_query(conn, "SELECT * FROM users WHERE id = $1", 123)` | object |
| `db_execute(connection, query, params...)` | Execute INSERT/UPDATE/DELETE | `result = db_execute(conn, "INSERT INTO users (name) VALUES ($1)", "John")` | object |
| `db_execute_batch(connection, queries)` | Execute multiple queries in batch | `result = db_execute_batch(conn, [query1, query2, query3])` | object |

### Asynchronous Operations

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `db_execute_async(connection, query, params...)` | Execute query asynchronously | `async_result = db_execute_async(conn, "SELECT * FROM large_table")` | object |
| `db_get_async_status(operation_id)` | Get status of async operation | `status = db_get_async_status("op_123")` | object |
| `db_cancel_async(operation_id)` | Cancel async operation | `cancel_result = db_cancel_async("op_123")` | object |
| `db_list_async_operations()` | List all async operations | `ops = db_list_async_operations()` | object |
| `db_cleanup_async_operations()` | Clean up completed async operations | `cleanup_result = db_cleanup_async_operations()` | object |

### Real-time Streaming

| Function | Description | Example |
|----------|-------------|----------|
| `stream_tables(connection, table_name, callback)` | Start real-time streaming for single table | `stream_tables(conn, "users", on_change)` |
| `stream_tables(connection, [table_names], callback)` | Start real-time streaming for multiple tables | `stream_tables(conn, ["users", "orders"], on_multi_change)` |
| `db_stop_stream(streamer_id)` | Stop real-time streaming | `db_stop_stream("streamer_123")` |

#### Callback Function Signatures

**Single Table Streaming:**
```uddin
fun onSingleTableChange(channel, payload):
    // channel: notification channel name (string)
    // payload: JSON string with change details
end
```

**Multi-Table Streaming:**
```uddin
fun onMultiTableChange(channel, payload):
    // channel: notification channel name (string)
    // payload: JSON string with change details including table name
end
```

## Basic Database Operations

### Connecting to PostgreSQL

```uddin
// Connect to PostgreSQL
conn_result = db_connect(
    "postgres",     // driver
    "localhost",    // host
    5432,           // port
    "myapp",        // database
    "postgres",     // username
    "password"      // password
)

if (conn_result.success) then:
    print("Connected successfully!")
    conn = conn_result.conn

    // Use the connection...

    // Always close when done
    db_close(conn)
else:
    print("Connection failed: " + conn_result.error)
end
```

### Connecting to MySQL

```uddin
// Connect to MySQL
conn_result = db_connect(
    "mysql",        // driver
    "localhost",    // host
    3306,           // port
    "myapp",        // database
    "root",         // username
    "password"      // password
)

if (conn_result.success) then:
    print("MySQL connected!")
    conn = conn_result.conn

    // Use the connection...

    db_close(conn)
else:
    print("MySQL connection failed: " + conn_result.error)
end
```

**Note:** For MySQL CDC (Change Data Capture) and binlog streaming, additional configuration is required. See the [Advanced MySQL Configuration](#advanced-mysql-configuration) section below.

## Query Operations

### SELECT Queries

```uddin
// Simple SELECT
result = db_query(conn, "SELECT id, name, email FROM users")

if (result.success) then:
    print("Found " + str(result.count) + " users:")
    for (row in result.data):
        print("ID: " + str(row.id) + ", Name: " + str(row.name))
    end
else:
    print("Query failed: " + result.error)
end

// Parameterized SELECT
user_id = 123
result = db_query(conn, "SELECT * FROM users WHERE id = $1", user_id)

if (result.success and result.count > 0) then:
    user = result.data[0]
    print("User found: " + str(user.name))
else:
    print("User not found")
end
```

### INSERT Operations

```uddin
// Insert new record
result = db_execute(conn,
    "INSERT INTO users (name, email, age) VALUES ($1, $2, $3)",
    "John Doe", "john@example.com", 30
)

if (result.success) then:
    print("User inserted successfully!")
    print("Rows affected: " + str(result.rows_affected))
    print("Last insert ID: " + str(result.last_insert_id))
else:
    print("Insert failed: " + result.error)
end
```

### UPDATE Operations

```uddin
// Update existing record
result = db_execute(conn,
    "UPDATE users SET email = $1 WHERE id = $2",
    "newemail@example.com", 123
)

if (result.success) then:
    print("Updated " + str(result.rows_affected) + " rows")
else:
    print("Update failed: " + result.error)
end
```

### DELETE Operations

```uddin
// Delete record
result = db_execute(conn,
    "DELETE FROM users WHERE id = $1",
    123
)

if (result.success) then:
    print("Deleted " + str(result.rows_affected) + " rows")
else:
    print("Delete failed: " + result.error)
end
```

## Advanced Database Operations

### Connection Pooling

Connection pooling improves performance by reusing database connections:

```uddin
// Connect with connection pool
conn_result = db_connect_with_pool(
    "postgres",     // driver
    "localhost",    // host
    5432,           // port
    "myapp",        // database
    "postgres",     // username
    "password",     // password
    10              // pool size
)

if (conn_result.success) then:
    conn = conn_result.conn
    
    // Configure pool settings
    pool_config = db_configure_pool(
        conn,
        20,    // max_open_connections
        10,    // max_idle_connections
        3600   // max_lifetime_seconds
    )
    
    if (pool_config.success) then:
        print("Pool configured successfully")
    end
    
    // Use connection for queries...
    
    db_close(conn)
end
```

### Batch Processing

Execute multiple queries efficiently in a single batch:

```uddin
// Prepare batch operations
batch_operations = [
    {
        "query": "INSERT INTO users (name, email) VALUES ($1, $2)",
        "params": ["John Doe", "john@example.com"]
    },
    {
        "query": "INSERT INTO users (name, email) VALUES ($1, $2)",
        "params": ["Jane Smith", "jane@example.com"]
    },
    {
        "query": "UPDATE users SET active = true WHERE id > $1",
        "params": [0]
    }
]

// Execute batch
batch_result = db_execute_batch(conn, batch_operations)

if (batch_result.success) then:
    print("Batch executed successfully")
    print("Operations completed: " + str(batch_result.operations_completed))
    print("Total rows affected: " + str(batch_result.total_rows_affected))
else:
    print("Batch failed: " + batch_result.error)
    print("Failed at operation: " + str(batch_result.failed_operation_index))
end
```

### Asynchronous Operations

Execute long-running queries without blocking:

```uddin
// Start async operation
async_result = db_execute_async(conn, 
    "SELECT * FROM large_table WHERE created_at > $1",
    "2024-01-01"
)

if (async_result.success) then:
    operation_id = async_result.operation_id
    print("Async operation started: " + operation_id)
    
    // Monitor progress
    while (true):
        status = db_get_async_status(operation_id)
        
        if (status.status == "running") then:
            print("Operation still running...")
            sleep(1000)  // Wait 1 second
        elif (status.status == "completed") then:
            print("Operation completed!")
            print("Rows returned: " + str(status.result.count))
            
            // Process results
            for (row in status.result.data):
                print("Row: " + str(row))
            end
            break
        elif (status.status == "failed") then:
            print("Operation failed: " + status.error)
            break
        elif (status.status == "cancelled") then:
            print("Operation was cancelled")
            break
        end
    end
else:
    print("Failed to start async operation: " + async_result.error)
end
```

### Managing Async Operations

```uddin
// List all async operations
all_ops = db_list_async_operations()

if (all_ops.success) then:
    print("Total operations: " + str(all_ops.count))
    
    for (op in all_ops.operations):
        print("Operation " + op.operation_id + ": " + op.status)
        
        // Cancel long-running operations if needed
        if (op.status == "running" and op.duration > 30000) then:
            cancel_result = db_cancel_async(op.operation_id)
            if (cancel_result.success) then:
                print("Cancelled operation: " + op.operation_id)
            end
        end
    end
end

// Clean up completed operations
cleanup_result = db_cleanup_async_operations()
if (cleanup_result.success) then:
    print("Cleaned up " + str(cleanup_result.cleaned_count) + " operations")
end
```

## Change Data Capture (CDC) and Real-time Streaming

UDDIN-LANG provides powerful Change Data Capture (CDC) capabilities for building real-time applications that respond to data changes. CDC enables applications to receive instant notifications when data changes, without polling overhead.

### Change Data Capture Concepts

Change Data Capture is a technique for identifying and capturing data changes in databases in real-time. UDDIN-LANG implements CDC with different approaches for each database:

- **Event-driven Architecture**: Applications react to data changes automatically
- **Low Latency**: Change notifications received in milliseconds
- **Scalable**: Supports monitoring multiple tables and high-throughput scenarios
- **Reliable**: Robust retry mechanisms and error handling

### UDDIN-LANG CDC Architecture

#### 1. Database Event Detection
- **PostgreSQL**: Uses LISTEN/NOTIFY mechanism with trigger functions
- **MySQL**: Binary log (binlog) streaming for real-time change detection
- **Event Filtering**: Ability to filter events based on specific conditions

#### 2. Event Processing Pipeline
- **Event Capture**: Detects data changes (INSERT, UPDATE, DELETE)
- **Event Enrichment**: Adds metadata such as timestamp, user context
- **Event Routing**: Sends events to appropriate callback functions
- **Error Handling**: Retry mechanism and dead letter queue for failed events

#### 3. Application Integration
- **Callback Functions**: Handler functions to process events
- **Multi-table Support**: Monitoring multiple tables in a single stream
- **Payload Structure**: Standardized JSON payload with complete information

### Stream Tables Function

The `stream_tables()` function is the main method for setting up real-time data streaming:

```uddin
// Single table streaming
stream_tables(connection, "table_name", callback_function)

// Multi-table streaming
stream_tables(connection, ["table1", "table2", "table3"], callback_function)
```

### Payload Structure

The callback function receives a JSON payload with detailed information about the data change:

```json
{
    "operation": "INSERT|UPDATE|DELETE",
    "table": "table_name",
    "timestamp": "2024-01-01T12:00:00Z",
    "data": {
        // New/current row data (for INSERT/UPDATE)
        "id": 123,
        "name": "John Doe",
        "email": "john@example.com"
    },
    "old_data": {
        // Previous row data (for UPDATE/DELETE)
        "id": 123,
        "name": "John Smith",
        "email": "john.smith@example.com"
    }
}
```

### Database-Specific Implementation

#### PostgreSQL CDC

PostgreSQL uses native LISTEN/NOTIFY mechanism:
- **Real-time**: Instant notifications with latency < 100ms
- **Automatic Setup**: Automatically creates triggers and notification functions
- **Efficient**: No polling overhead
- **Persistent**: Notifications persist even if connection drops
- **Scalable**: Supports thousands of concurrent listeners

#### MySQL CDC

MySQL uses binary log (binlog) streaming for real-time CDC:
- **Binlog Streaming**: Reads directly from MySQL binary logs (like Debezium)
- **Real-time**: 1-10ms latency for change detection
- **Minimal Impact**: <1% CPU usage, minimal database load
- **High Performance**: Supports 10,000+ events/sec throughput
- **Compatible**: Works with MySQL 5.7+ and MariaDB
- **Reliable**: Handles connection interruptions and binlog rotation gracefully
- **No Database Changes**: No triggers or additional tables required

#### Oracle CDC (Planned)

Planned implementation for Oracle Database:
- **Oracle Streams**: Uses Oracle Streams API
- **LogMiner**: Integration with Oracle LogMiner
- **XStream**: Support for Oracle XStream Out

#### MongoDB CDC (Planned)

Planned implementation for MongoDB:
- **Change Streams**: Uses MongoDB Change Streams
- **Oplog Tailing**: Monitors oplog for changes
- **Aggregation Pipeline**: Custom change stream pipelines

### CDC Use Cases

#### 1. Real-time Analytics and Dashboard
- **Live Metrics**: Update dashboard metrics in real-time
- **User Activity Tracking**: Monitor user behavior and engagement
- **Performance Monitoring**: Track application performance metrics

#### 2. Event-driven Microservices
- **Service Communication**: Trigger actions in other microservices
- **Data Synchronization**: Sync data between services
- **Event Sourcing**: Implement event sourcing patterns

#### 3. Business Process Automation
- **Order Processing**: Automatically process orders and inventory updates
- **Notification Systems**: Send real-time notifications to users
- **Workflow Triggers**: Trigger business workflows based on data changes

#### 4. Fraud Detection and Security
- **Anomaly Detection**: Detect suspicious patterns in real-time
- **Security Monitoring**: Monitor security events and threats
- **Compliance Tracking**: Track compliance-related data changes

## Future CDC Development Plans

### Advanced CDC Features (Roadmap)

#### 1. Multi-Database CDC Orchestration
- **Cross-Database Streaming**: Sync data changes across multiple databases
- **Conflict Resolution**: Automatic conflict resolution for distributed systems
- **Global Transaction Coordination**: Coordinate transactions across databases

#### 2. Advanced Filtering and Transformation
- **Conditional CDC**: Filter events based on complex conditions
- **Data Transformation**: Transform data before sending to callbacks
- **Schema Evolution**: Handle schema changes automatically

#### 3. Performance Enhancements
- **Batch Processing**: Batch multiple events for improved performance
- **Compression**: Compress event payloads for reduced network overhead
- **Partitioning**: Partition events based on keys for parallel processing

#### 4. Enterprise Features
- **Dead Letter Queue**: Handle failed events with retry mechanisms
- **Event Replay**: Replay historical events for debugging or recovery
- **Monitoring Dashboard**: Built-in monitoring and alerting for CDC streams

### Cloud Services Integration

#### AWS Integration (Planned)
- **Amazon RDS**: Native integration with RDS PostgreSQL and MySQL
- **Amazon DynamoDB**: CDC support for DynamoDB Streams
- **Amazon Kinesis**: Stream events to Kinesis for further processing

#### Google Cloud Integration (Planned)
- **Cloud SQL**: Integration with Cloud SQL PostgreSQL and MySQL
- **Cloud Spanner**: CDC support for globally distributed databases
- **Pub/Sub**: Stream events to Google Cloud Pub/Sub

#### Azure Integration (Planned)
- **Azure Database**: Support for Azure Database for PostgreSQL and MySQL
- **Cosmos DB**: CDC integration with Cosmos DB Change Feed
- **Service Bus**: Stream events to Azure Service Bus
### Best Practices for CDC Implementation

#### 1. Performance Optimization
- **Connection Pooling**: Use connection pooling for multiple CDC streams
- **Batch Processing**: Process multiple events in batches for efficiency
- **Index Optimization**: Ensure proper indexing on monitored tables
- **Memory Management**: Monitor memory usage for long-running CDC processes

#### 2. Error Handling and Resilience
- **Retry Logic**: Implement exponential backoff for failed connections
- **Circuit Breaker**: Prevent cascade failures with circuit breaker pattern
- **Health Checks**: Regular health checks for CDC connections
- **Graceful Degradation**: Fallback mechanisms when CDC unavailable

#### 3. Security Considerations
- **Least Privilege**: Grant minimal permissions required for CDC
- **Data Encryption**: Encrypt sensitive data in CDC streams
- **Audit Logging**: Log all CDC activities for compliance
- **Access Control**: Implement proper access control for CDC endpoints

## Troubleshooting CDC Issues

### Common Issues and Solutions

#### 1. No Events Received
- **Check Database Permissions**: Ensure user has permissions for LISTEN/NOTIFY
- **Verify Table Access**: Confirm access to monitored tables
- **Network Connectivity**: Check network connection to database
- **Firewall Settings**: Ensure ports are open for database connections

#### 2. Connection Drops
- **Connection Timeout**: Adjust connection timeout settings
- **Network Stability**: Monitor network stability and latency
- **Database Load**: Check database performance and resource usage
- **Reconnection Logic**: Implement automatic reconnection with exponential backoff

#### 3. High Memory Usage
- **Batch Size Optimization**: Reduce batch sizes for large datasets
- **Memory Limits**: Set memory limits for CDC processes
- **Garbage Collection**: Implement proper memory cleanup
- **Data Filtering**: Filter unnecessary data before processing

#### 4. Performance Issues
- **Index Optimization**: Ensure proper indexing on monitored tables
- **Query Optimization**: Optimize CDC queries for better performance
- **Connection Pooling**: Use connection pooling for multiple streams
- **Parallel Processing**: Process events in parallel when possible

## Advanced MySQL Configuration

For MySQL CDC (Change Data Capture) and binlog streaming, additional database configuration is required to enable real-time change detection.

### Prerequisites

MySQL must be configured for binary logging:

```sql
-- Enable binary logging
SET GLOBAL log_bin = ON;

-- Set binlog format to ROW (required)
SET GLOBAL binlog_format = 'ROW';

-- Set full row image (recommended)
SET GLOBAL binlog_row_image = 'FULL';

-- Create replication user (optional, for security)
CREATE USER 'repl_user'@'%' IDENTIFIED BY 'password';
GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'repl_user'@'%';
FLUSH PRIVILEGES;
```

### my.cnf Configuration

Add the following to your MySQL configuration file:

```ini
[mysqld]
# Enable binary logging
log-bin=mysql-bin
server-id=1

# Set binlog format
binlog-format=ROW
binlog-row-image=FULL

# Binlog retention (optional)
expire_logs_days=7
max_binlog_size=100M
```

### Runtime Configuration

You can also configure binlog settings at runtime:

```uddin
// Enable binlog configuration (if not set globally)
db_execute(db_conn, "SET GLOBAL log_bin = ON")
db_execute(db_conn, "SET GLOBAL binlog_format = 'ROW'")
db_execute(db_conn, "SET GLOBAL binlog_row_image = 'FULL'")

// Verify configuration
result = db_query(db_conn, "SHOW VARIABLES LIKE 'binlog_format'")
if (result.success and result.count > 0) then:
    print("Binlog format: " + result.data[0].Value)
end
```

### Security Considerations

1. **Dedicated User**: Create a dedicated user for CDC operations
2. **Minimal Privileges**: Grant only necessary replication privileges
3. **Network Security**: Use SSL connections for production environments
4. **Access Control**: Restrict access to binlog files

### Performance Tuning

1. **Binlog Size**: Adjust `max_binlog_size` based on your write volume
2. **Retention**: Set appropriate `expire_logs_days` to manage disk space
3. **Buffer Size**: Tune `binlog_cache_size` for high-throughput scenarios
4. **Sync Settings**: Configure `sync_binlog` based on durability requirements

For more detailed information, see the [MySQL Binlog Streaming Reference](../../reference/mysql-binlog-streaming.md).

## Summary

UDDIN-LANG provides powerful database integration capabilities with a focus on Change Data Capture (CDC) and real-time streaming:

### Core Capabilities
- **Multi-database Support**: Connection to PostgreSQL, MySQL, and other databases
- **Change Data Capture (CDC)**: Monitor data changes in real-time
- **Simple API**: Easy-to-use functions for common database operations
- **Real-time Streaming**: Stream data changes with low latency

### CDC Features
- **Event-driven Architecture**: Respond to data changes automatically
- **Multiple Event Types**: Handle INSERT, UPDATE, DELETE operations
- **Flexible Filtering**: Monitor specific tables or columns
- **Scalable Processing**: Handle high-volume data changes efficiently

### Use Cases for CDC
- **Real-time Analytics**: Update dashboards and reports instantly
- **Event-driven Microservices**: Trigger business processes from data changes
- **Data Synchronization**: Sync data between systems in real-time
- **Audit and Compliance**: Track all data changes for regulatory requirements
- **Cache Invalidation**: Update caches when underlying data changes

### Development Roadmap
UDDIN-LANG continues to develop CDC capabilities with planned support for:
- Oracle and MongoDB CDC
- Advanced filtering and transformation
- Cloud service integrations (AWS, GCP, Azure)
- Enterprise features such as dead letter queues and event replay

With a focus on CDC and real-time streaming, UDDIN-LANG enables developers to build applications that are responsive to data changes, supporting modern event-driven and scalable architectures.
