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

| Function | Description | Example |
|----------|-------------|----------|
| `db_connect(driver, host, port, database, username, password)` | Connect to database | `conn = db_connect("postgres", "localhost", 5432, "mydb", "user", "pass")` |
| `db_close(connection)` | Close database connection | `db_close(conn)` |

### Query Operations

| Function | Description | Example |
|----------|-------------|----------|
| `db_query(connection, query, params...)` | Execute SELECT query | `result = db_query(conn, "SELECT * FROM users WHERE id = $1", 123)` |
| `db_execute(connection, query, params...)` | Execute INSERT/UPDATE/DELETE | `result = db_execute(conn, "INSERT INTO users (name) VALUES ($1)", "John")` |

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

## Change Data Capture (CDC) dan Real-time Streaming

UDDIN-LANG menyediakan kemampuan Change Data Capture (CDC) yang powerful untuk membangun aplikasi real-time yang responsif terhadap perubahan data. CDC memungkinkan aplikasi untuk menerima notifikasi instan ketika data berubah, tanpa overhead polling.

### Konsep Change Data Capture

Change Data Capture adalah teknik untuk mengidentifikasi dan menangkap perubahan data dalam database secara real-time. UDDIN-LANG mengimplementasikan CDC dengan pendekatan yang berbeda untuk setiap database:

- **Event-driven Architecture**: Aplikasi bereaksi terhadap perubahan data secara otomatis
- **Low Latency**: Notifikasi perubahan diterima dalam hitungan milidetik
- **Scalable**: Mendukung monitoring multiple tables dan high-throughput scenarios
- **Reliable**: Mekanisme retry dan error handling yang robust

### Arsitektur CDC UDDIN-LANG

#### 1. Database Event Detection
- **PostgreSQL**: Menggunakan LISTEN/NOTIFY mechanism dengan trigger functions
- **MySQL**: Implementasi trigger-based change detection dengan polling optimization
- **Event Filtering**: Kemampuan untuk memfilter events berdasarkan kondisi tertentu

#### 2. Event Processing Pipeline
- **Event Capture**: Mendeteksi perubahan data (INSERT, UPDATE, DELETE)
- **Event Enrichment**: Menambahkan metadata seperti timestamp, user context
- **Event Routing**: Mengirim events ke callback functions yang sesuai
- **Error Handling**: Retry mechanism dan dead letter queue untuk failed events

#### 3. Application Integration
- **Callback Functions**: Handler functions untuk memproses events
- **Multi-table Support**: Monitoring multiple tables dalam satu stream
- **Payload Structure**: Standardized JSON payload dengan informasi lengkap

### Stream Tables Function

Fungsi `stream_tables()` adalah method utama untuk setup real-time data streaming:

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

PostgreSQL menggunakan LISTEN/NOTIFY mechanism yang native:
- **Real-time**: Notifikasi instan dengan latency < 100ms
- **Automatic Setup**: Otomatis membuat triggers dan notification functions
- **Efficient**: Tanpa overhead polling
- **Persistent**: Notifikasi bertahan meski koneksi terputus
- **Scalable**: Mendukung ribuan concurrent listeners

#### MySQL CDC

MySQL menggunakan trigger-based change detection:
- **Trigger-based**: Menggunakan MySQL triggers untuk capture changes
- **Optimized Polling**: Mekanisme polling yang dioptimasi
- **Compatible**: Bekerja dengan MySQL 5.7+ dan MariaDB
- **Reliable**: Menangani connection interruptions dengan graceful
- **Flexible**: Mendukung custom filtering dan transformation

#### Oracle CDC (Planned)

Rencana implementasi untuk Oracle Database:
- **Oracle Streams**: Menggunakan Oracle Streams API
- **LogMiner**: Integration dengan Oracle LogMiner
- **XStream**: Support untuk Oracle XStream Out

#### MongoDB CDC (Planned)

Rencana implementasi untuk MongoDB:
- **Change Streams**: Menggunakan MongoDB Change Streams
- **Oplog Tailing**: Monitoring oplog untuk changes
- **Aggregation Pipeline**: Custom change stream pipelines

### Use Cases CDC

#### 1. Real-time Analytics dan Dashboard
- **Live Metrics**: Update dashboard metrics secara real-time
- **User Activity Tracking**: Monitor user behavior dan engagement
- **Performance Monitoring**: Track application performance metrics

#### 2. Event-driven Microservices
- **Service Communication**: Trigger actions di microservices lain
- **Data Synchronization**: Sync data antar services
- **Event Sourcing**: Implement event sourcing patterns

#### 3. Business Process Automation
- **Order Processing**: Otomatis process orders dan inventory updates
- **Notification Systems**: Send real-time notifications ke users
- **Workflow Triggers**: Trigger business workflows berdasarkan data changes

#### 4. Fraud Detection dan Security
- **Anomaly Detection**: Detect suspicious patterns dalam real-time
- **Security Monitoring**: Monitor security events dan threats
- **Compliance Tracking**: Track compliance-related data changes

## Rencana Pengembangan CDC Masa Depan

### Advanced CDC Features (Roadmap)

#### 1. Multi-Database CDC Orchestration
- **Cross-Database Streaming**: Sync data changes antar multiple databases
- **Conflict Resolution**: Automatic conflict resolution untuk distributed systems
- **Global Transaction Coordination**: Coordinate transactions across databases

#### 2. Advanced Filtering dan Transformation
- **Conditional CDC**: Filter events berdasarkan complex conditions
- **Data Transformation**: Transform data sebelum dikirim ke callbacks
- **Schema Evolution**: Handle schema changes secara otomatis

#### 3. Performance Enhancements
- **Batch Processing**: Batch multiple events untuk improved performance
- **Compression**: Compress event payloads untuk reduced network overhead
- **Partitioning**: Partition events berdasarkan keys untuk parallel processing

#### 4. Enterprise Features
- **Dead Letter Queue**: Handle failed events dengan retry mechanisms
- **Event Replay**: Replay historical events untuk debugging atau recovery
- **Monitoring Dashboard**: Built-in monitoring dan alerting untuk CDC streams

### Integration dengan Cloud Services

#### AWS Integration (Planned)
- **Amazon RDS**: Native integration dengan RDS PostgreSQL dan MySQL
- **Amazon DynamoDB**: CDC support untuk DynamoDB Streams
- **Amazon Kinesis**: Stream events ke Kinesis untuk further processing

#### Google Cloud Integration (Planned)
- **Cloud SQL**: Integration dengan Cloud SQL PostgreSQL dan MySQL
- **Cloud Spanner**: CDC support untuk globally distributed databases
- **Pub/Sub**: Stream events ke Google Cloud Pub/Sub

#### Azure Integration (Planned)
- **Azure Database**: Support untuk Azure Database for PostgreSQL dan MySQL
- **Cosmos DB**: CDC integration dengan Cosmos DB Change Feed
- **Service Bus**: Stream events ke Azure Service Bus
### Best Practices untuk CDC Implementation

#### 1. Performance Optimization
- **Connection Pooling**: Gunakan connection pooling untuk multiple CDC streams
- **Batch Processing**: Process multiple events dalam batches untuk efficiency
- **Index Optimization**: Ensure proper indexing pada tables yang di-monitor
- **Memory Management**: Monitor memory usage untuk long-running CDC processes

#### 2. Error Handling dan Resilience
- **Retry Logic**: Implement exponential backoff untuk failed connections
- **Circuit Breaker**: Prevent cascade failures dengan circuit breaker pattern
- **Health Checks**: Regular health checks untuk CDC connections
- **Graceful Degradation**: Fallback mechanisms ketika CDC unavailable

#### 3. Security Considerations
- **Least Privilege**: Grant minimal permissions yang diperlukan untuk CDC
- **Data Encryption**: Encrypt sensitive data dalam CDC streams
- **Audit Logging**: Log semua CDC activities untuk compliance
- **Access Control**: Implement proper access control untuk CDC endpoints
```

## Troubleshooting CDC Issues

### Common Issues dan Solutions

#### 1. No Events Received
- **Check Database Permissions**: Ensure user memiliki permissions untuk LISTEN/NOTIFY
- **Verify Table Access**: Confirm access ke tables yang di-monitor
- **Network Connectivity**: Check network connection ke database
- **Firewall Settings**: Ensure ports terbuka untuk database connections

#### 2. Connection Drops
- **Connection Timeout**: Adjust connection timeout settings
- **Network Stability**: Monitor network stability dan latency
- **Database Load**: Check database performance dan resource usage
- **Reconnection Logic**: Implement automatic reconnection dengan exponential backoff

#### 3. High Memory Usage
- **Batch Size Optimization**: Reduce batch sizes untuk large datasets
- **Memory Limits**: Set memory limits untuk CDC processes
- **Garbage Collection**: Implement proper memory cleanup
- **Data Filtering**: Filter unnecessary data sebelum processing

#### 4. Performance Issues
- **Index Optimization**: Ensure proper indexing pada monitored tables
- **Query Optimization**: Optimize CDC queries untuk better performance
- **Connection Pooling**: Use connection pooling untuk multiple streams
- **Parallel Processing**: Process events secara parallel ketika memungkinkan

## Summary

UDDIN-LANG menyediakan kemampuan integrasi database yang powerful dengan fokus pada Change Data Capture (CDC) dan real-time streaming:

### Core Capabilities
- **Multi-database Support**: Koneksi ke PostgreSQL, MySQL, dan database lainnya
- **Change Data Capture (CDC)**: Monitor perubahan data secara real-time
- **Simple API**: Fungsi yang mudah digunakan untuk operasi database umum
- **Real-time Streaming**: Stream perubahan data dengan latency rendah

### CDC Features
- **Event-driven Architecture**: Respond to data changes secara otomatis
- **Multiple Event Types**: Handle INSERT, UPDATE, DELETE operations
- **Flexible Filtering**: Monitor specific tables atau columns
- **Scalable Processing**: Handle high-volume data changes efficiently

### Use Cases untuk CDC
- **Real-time Analytics**: Update dashboards dan reports secara instant
- **Event-driven Microservices**: Trigger business processes dari data changes
- **Data Synchronization**: Sync data antar systems secara real-time
- **Audit dan Compliance**: Track semua perubahan data untuk regulatory requirements
- **Cache Invalidation**: Update caches ketika underlying data berubah

### Rencana Pengembangan
UDDIN-LANG terus mengembangkan CDC capabilities dengan rencana support untuk:
- Oracle dan MongoDB CDC
- Advanced filtering dan transformation
- Cloud service integrations (AWS, GCP, Azure)
- Enterprise features seperti dead letter queues dan event replay

Dengan fokus pada CDC dan real-time streaming, UDDIN-LANG memungkinkan developers untuk membangun aplikasi yang responsive terhadap perubahan data, mendukung arsitektur modern yang event-driven dan scalable.