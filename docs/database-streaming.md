# Database Streaming in UddinLang

UddinLang provides real-time database streaming capabilities using PostgreSQL's LISTEN/NOTIFY mechanism. This allows your applications to receive instant notifications when data changes, without the overhead of polling.

## Overview

Database streaming in UddinLang offers:
- **Real-time notifications** - Instant updates when data changes
- **Zero polling overhead** - No repeated queries
- **Automatic trigger setup** - Handles PostgreSQL triggers and functions automatically
- **JSON payloads** - Structured data about what changed
- **Low resource usage** - Efficient connection-based notifications

## Functions

### `query_stream(connection, table_name, callback_function)`

Starts real-time streaming for a specific database table.

**Parameters:**
- `connection`: Database connection object (PostgreSQL only)
- `table_name`: Name of the table to monitor
- `callback_function`: Function to call when changes occur

**Callback Function Signature:**
```uddin
function onStreamData(channel, payload) {
    // channel: The notification channel name
    // payload: JSON string with change details
}
```

**Payload Structure:**
```json
{
    "operation": "INSERT|UPDATE|DELETE",
    "table": "table_name",
    "timestamp": "2024-01-01T12:00:00Z",
    "data": {
        // Changed row data (for INSERT/UPDATE)
    },
    "old_data": {
        // Previous row data (for UPDATE/DELETE)
    }
}
```

### `db_stop_stream(table_name)`

Stops streaming for a specific table.

**Parameters:**
- `table_name`: Name of the table to stop monitoring

## Example Usage

### Basic Streaming Example

```uddin
// Connect to PostgreSQL
conn = db_connect("postgres", "localhost", 5432, "mydb", "user", "pass").conn

// Define callback function
function onDataChange(channel, payload) {
    print("Data changed in channel: " + channel)
    print("Payload: " + payload)
    
    // Parse JSON payload
    data = json_parse(payload)
    print("Operation: " + data.operation)
    print("Table: " + data.table)
}

// Start streaming
query_stream(conn, "users", onDataChange)

// Keep program running
while (true) {
    sleep(1000)
}
```

### Advanced Example with Multiple Tables

```uddin
// Connect to database
conn = db_connect("postgres", "localhost", 5432, "mydb", "user", "pass").conn

// Callback for user changes
function onUserChange(channel, payload) {
    data = json_parse(payload)
    if (data.operation == "INSERT") {
        print("New user created: " + data.data.name)
    } else if (data.operation == "UPDATE") {
        print("User updated: " + data.data.name)
    } else if (data.operation == "DELETE") {
        print("User deleted: " + data.old_data.name)
    }
}

// Callback for order changes
function onOrderChange(channel, payload) {
    data = json_parse(payload)
    print("Order " + data.operation + ": $" + str(data.data.amount))
}

// Start streaming for multiple tables
query_stream(conn, "users", onUserChange)
query_stream(conn, "orders", onOrderChange)

// Application logic here...

// Stop streaming when done
db_stop_stream("users")
db_stop_stream("orders")
```

## How It Works

1. **Automatic Setup**: When you call `query_stream()`, UddinLang automatically:
   - Creates a PostgreSQL function to send notifications
   - Creates triggers on INSERT, UPDATE, and DELETE operations
   - Sets up a LISTEN connection for the table

2. **Real-time Notifications**: When data changes:
   - PostgreSQL triggers fire immediately
   - Notification is sent through LISTEN/NOTIFY
   - Your callback function is called with the change data

3. **Cleanup**: When you call `db_stop_stream()`:
   - Stops listening for notifications
   - Optionally removes triggers and functions

## Key Features

- **Real-time**: Instant notifications using PostgreSQL LISTEN/NOTIFY
- **Low Latency**: < 100ms response time
- **Efficient**: Minimal database load and resource usage
- **Automatic Setup**: Triggers and functions created automatically
- **PostgreSQL Only**: Designed specifically for PostgreSQL databases

## Best Practices

1. **Use for Real-time Applications**: Perfect for chat apps, dashboards, notifications
2. **Handle Connection Errors**: Implement reconnection logic for production use
3. **Parse JSON Safely**: Always validate payload structure
4. **Clean Up Resources**: Call `db_stop_stream()` when done
5. **Monitor Performance**: Watch for high-frequency changes that might overwhelm callbacks

## Limitations

- **PostgreSQL Only**: Currently only supports PostgreSQL databases
- **Network Dependency**: Requires persistent database connection
- **Trigger Overhead**: Small overhead for each data modification
- **JSON Parsing**: Payload parsing adds minimal processing time

## Error Handling

```uddin
try:
    query_stream(conn, "users", onUserChange)
catch (error):
    print("Failed to start streaming: " + str(error))
    // Handle error (reconnect, fallback to polling, etc.)
end
```

## Usage Examples

Basic streaming setup:

```uddin
// Start real-time streaming
query_stream(conn, "users", onDataChange)
```

The callback signature changes slightly:
- Polling: `callback(data)` - receives query results
- Streaming: `callback(channel, payload)` - receives notification data

## Troubleshooting

**Connection Issues:**
- Ensure PostgreSQL server is running
- Check network connectivity
- Verify database permissions

**No Notifications:**
- Check if triggers were created successfully
- Verify table name spelling
- Ensure data is actually changing

**Performance Issues:**
- Monitor callback execution time
- Consider batching high-frequency changes
- Use connection pooling for multiple streams