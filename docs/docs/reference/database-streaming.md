---
sidebar_position: 6
---

# Database Streaming Reference

This reference guide covers all aspects of UddinLang's real-time database streaming capabilities.

## Functions

### `stream_tables(connection, table_name, callback)`

Starts real-time streaming for a single database table.

**Parameters:**
- `connection` (object): Database connection object
- `table_name` (string): Name of the table to monitor
- `callback` (function): Function to call when changes occur

**Returns:** Stream identifier for stopping the stream

**Example:**
```uddin
stream_tables(conn, "users", onUserChange)
```

### `stream_tables(connection, [table_names], callback)`

Starts real-time streaming for multiple database tables.

**Parameters:**
- `connection` (object): Database connection object
- `table_names` (array): Array of table names to monitor
- `callback` (function): Function to call when changes occur

**Returns:** Stream identifier for stopping the stream

**Example:**
```uddin
stream_tables(conn, ["users", "orders", "products"], onMultiTableChange)
```

### `db_stop_stream(stream_id)`

Stops an active database stream.

**Parameters:**
- `stream_id` (string): Stream identifier returned by `stream_tables()`

**Returns:** Boolean indicating success

**Example:**
```uddin
db_stop_stream("stream_123")
```

## Callback Function Signatures

### Single Table Callback

```uddin
fun onTableChange(channel, payload):
    // channel: notification channel name (string)
    // payload: JSON string with change details
end
```

### Multi-Table Callback

```uddin
fun onMultiTableChange(channel, payload):
    // channel: notification channel name (string)
    // payload: JSON string with change details including table name
end
```

## Payload Structure

The callback function receives a JSON payload with the following structure:

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

### Payload Fields

| Field | Type | Description | Available For |
|-------|------|-------------|---------------|
| `operation` | string | Type of database operation | All operations |
| `table` | string | Name of the affected table | All operations |
| `timestamp` | string | ISO 8601 timestamp of the change | All operations |
| `data` | object | New/current row data | INSERT, UPDATE |
| `old_data` | object | Previous row data | UPDATE, DELETE |

## Database Support

### PostgreSQL

**Technology:** LISTEN/NOTIFY mechanism

**Features:**
- Real-time notifications (< 100ms latency)
- Automatic trigger and function creation
- No polling overhead
- Persistent notifications

**Requirements:**
- PostgreSQL 9.0+
- Database user must have trigger creation privileges

**Example:**
```uddin
conn = db_connect("postgres", "localhost", 5432, "mydb", "postgres", "password").conn
stream_tables(conn, "users", onUserChange)
```

### MySQL

**Technology:** Trigger-based change data capture

**Features:**
- Trigger-based change detection
- Efficient polling mechanism
- Compatible with MySQL 5.7+ and MariaDB
- Handles connection interruptions gracefully

**Requirements:**
- MySQL 5.7+ or MariaDB 10.2+
- Database user must have trigger creation privileges

**Example:**
```uddin
conn = db_connect("mysql", "localhost", 3306, "mydb", "root", "password").conn
stream_tables(conn, "users", onUserChange)
```

## Error Handling

### Common Errors

| Error | Description | Solution |
|-------|-------------|----------|
| `Permission denied` | User lacks trigger privileges | Grant trigger privileges to database user |
| `Table not found` | Specified table doesn't exist | Verify table name and existence |
| `Connection lost` | Database connection dropped | Implement reconnection logic |
| `Invalid callback` | Callback function not found | Ensure callback function is defined |

### Error Handling Example

```uddin
fun robust_stream_setup(conn, tables, callback):
    try:
        stream_id = stream_tables(conn, tables, callback)
        print("✅ Streaming started successfully")
        return stream_id
    catch (error):
        print("❌ Failed to start streaming: " + str(error))
        
        // Handle specific errors
        if (contains(str(error), "permission")) then:
            print("💡 Grant trigger privileges to database user")
        elif (contains(str(error), "table")) then:
            print("💡 Verify table names exist in database")
        end
        
        return null
    end
end
```

## Performance Considerations

### Latency

- **PostgreSQL**: < 100ms typical latency
- **MySQL**: 100-500ms typical latency (depends on polling interval)

### Memory Usage

- Minimal memory overhead for connection management
- Payload size depends on row data size
- Consider batching for high-frequency changes

### Connection Limits

- Each `stream_tables()` call uses one database connection
- Monitor connection pool usage in high-concurrency scenarios
- Use connection pooling for multiple streams

### Optimization Tips

1. **Batch Processing**: Group rapid changes to reduce callback overhead
2. **Selective Monitoring**: Only monitor tables that require real-time updates
3. **Connection Pooling**: Reuse connections when possible
4. **Error Recovery**: Implement automatic reconnection logic

## Security Considerations

### Database Privileges

**Required Privileges:**
- `SELECT` on monitored tables
- `TRIGGER` privilege for creating triggers
- `CREATE FUNCTION` privilege (PostgreSQL only)

**Recommended Setup:**
```sql
-- PostgreSQL
GRANT SELECT, TRIGGER ON TABLE users TO streaming_user;
GRANT CREATE ON SCHEMA public TO streaming_user;

-- MySQL
GRANT SELECT, TRIGGER ON mydb.users TO 'streaming_user'@'%';
```

### Connection Security

- Use SSL/TLS connections in production
- Limit database user privileges to minimum required
- Use connection pooling with authentication
- Monitor for unusual connection patterns

## Troubleshooting

### No Notifications Received

1. **Check Triggers**: Verify triggers were created successfully
2. **Test Data Changes**: Manually insert/update/delete data
3. **Connection Status**: Ensure database connection is active
4. **Privileges**: Verify user has required database privileges

### High Memory Usage

1. **Callback Efficiency**: Optimize callback function performance
2. **Batch Processing**: Group multiple changes together
3. **Memory Cleanup**: Clear references in callback functions
4. **Connection Limits**: Monitor active connection count

### Connection Drops

1. **Network Stability**: Check network connection reliability
2. **Database Timeouts**: Adjust connection timeout settings
3. **Reconnection Logic**: Implement automatic reconnection
4. **Health Monitoring**: Regular connection health checks

## Best Practices

### 1. Error Handling

```uddin
fun safe_callback(channel, payload):
    try:
        data = json_parse(payload)
        process_change(data)
    catch (error):
        log_error("Callback error: " + str(error))
        // Don't let errors break the stream
    end
end
```

### 2. Connection Management

```uddin
fun setup_resilient_connection(config):
    max_retries = 3
    for (i in range(max_retries)):
        conn_result = db_connect(
            config.driver, config.host, config.port,
            config.database, config.username, config.password
        )
        
        if (conn_result.success) then:
            return conn_result.conn
        end
        
        sleep(5000)  // Wait before retry
    end
    
    return null
end
```

### 3. Performance Optimization

```uddin
fun optimized_callback(channel, payload):
    // Use static variables for batching
    static batch = []
    static last_process = time_now()
    
    append(batch, json_parse(payload))
    
    // Process batch every second or when it reaches 100 items
    if (len(batch) >= 100 or (time_now() - last_process) >= 1000) then:
        process_batch(batch)
        batch = []
        last_process = time_now()
    end
end
```

### 4. Monitoring and Logging

```uddin
fun monitored_callback(channel, payload):
    start_time = time_now()
    
    try:
        data = json_parse(payload)
        process_change(data)
        
        // Log successful processing
        processing_time = time_now() - start_time
        if (processing_time > 1000) then:  // Log slow operations
            print("⚠️ Slow callback processing: " + str(processing_time) + "ms")
        end
        
    catch (error):
        print("❌ Callback error: " + str(error))
    end
end
```

## Examples

### Basic Single Table Streaming

```uddin
conn = db_connect("postgres", "localhost", 5432, "mydb", "user", "pass").conn

fun onUserChange(channel, payload):
    data = json_parse(payload)
    print("User " + data.operation + ": " + data.data.name)
end

stream_tables(conn, "users", onUserChange)
```

### Multi-Table E-commerce Monitoring

```uddin
conn = db_connect("postgres", "localhost", 5432, "ecommerce", "user", "pass").conn

fun onEcommerceChange(channel, payload):
    data = json_parse(payload)
    
    if (data.table == "orders" and data.operation == "INSERT") then:
        print("🛒 New order: $" + str(data.data.total))
    elif (data.table == "inventory" and data.data.quantity < 10) then:
        print("⚠️ Low stock: " + data.data.product_name)
    end
end

stream_tables(conn, ["orders", "inventory", "users"], onEcommerceChange)
```

### Real-time Analytics Dashboard

```uddin
conn = db_connect("postgres", "localhost", 5432, "analytics", "user", "pass").conn

// Global counters
static daily_orders = 0
static daily_revenue = 0.0

fun onAnalyticsUpdate(channel, payload):
    data = json_parse(payload)
    
    if (data.table == "orders" and data.operation == "INSERT") then:
        daily_orders = daily_orders + 1
        daily_revenue = daily_revenue + data.data.total
        
        print("📊 Daily Stats: " + str(daily_orders) + " orders, $" + str(daily_revenue))
        
        // Update dashboard
        update_dashboard({
            "orders": daily_orders,
            "revenue": daily_revenue
        })
    end
end

stream_tables(conn, "orders", onAnalyticsUpdate)
```

This reference guide provides comprehensive information for implementing real-time database streaming in UddinLang applications.