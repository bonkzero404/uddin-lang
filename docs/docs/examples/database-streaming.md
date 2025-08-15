---
sidebar_position: 2
---

# Database Streaming Examples

This page provides practical examples of using UddinLang's real-time database streaming capabilities.

## Basic Single Table Streaming

### PostgreSQL User Monitoring

```uddin
// Connect to PostgreSQL database
conn_result = db_connect("postgres", "localhost", 5432, "myapp", "postgres", "password")

if (not conn_result.success) then:
    print("❌ Failed to connect: " + conn_result.error)
    exit(1)
end

conn = conn_result.conn

// Define callback function for user changes
fun onUserChange(channel, payload):
    data = json_parse(payload)
    
    if (data.operation == "INSERT") then:
        print("👤 New user registered: " + data.data.username + " (" + data.data.email + ")")
    elif (data.operation == "UPDATE") then:
        print("✏️ User updated: " + data.data.username)
        if (data.old_data.email != data.data.email) then:
            print("   📧 Email changed: " + data.old_data.email + " → " + data.data.email)
        end
    elif (data.operation == "DELETE") then:
        print("🗑️ User deleted: " + data.old_data.username)
    end
end

// Start streaming
stream_id = stream_tables(conn, "users", onUserChange)
print("🔄 Started monitoring users table...")

// Keep the program running
while (true):
    sleep(1000)
end
```

### MySQL Product Inventory Monitoring

```uddin
// Connect to MySQL database
conn_result = db_connect("mysql", "localhost", 3306, "inventory", "root", "password")

if (not conn_result.success) then:
    print("❌ Failed to connect: " + conn_result.error)
    exit(1)
end

conn = conn_result.conn

// Define callback function for inventory changes
fun onInventoryChange(channel, payload):
    data = json_parse(payload)
    
    if (data.operation == "UPDATE") then:
        old_qty = data.old_data.quantity
        new_qty = data.data.quantity
        product_name = data.data.product_name
        
        if (new_qty < old_qty) then:
            print("📦 Stock decreased: " + product_name + " (" + str(old_qty) + " → " + str(new_qty) + ")")
            
            if (new_qty <= 10) then:
                print("⚠️ LOW STOCK ALERT: " + product_name + " has only " + str(new_qty) + " items left!")
            end
        elif (new_qty > old_qty) then:
            print("📈 Stock replenished: " + product_name + " (" + str(old_qty) + " → " + str(new_qty) + ")")
        end
    end
end

// Start streaming
stream_id = stream_tables(conn, "products", onInventoryChange)
print("🔄 Started monitoring product inventory...")

// Keep the program running
while (true):
    sleep(1000)
end
```

## Multi-Table Streaming

### E-commerce Real-time Dashboard

```uddin
// Connect to PostgreSQL database
conn_result = db_connect("postgres", "localhost", 5432, "ecommerce", "postgres", "password")

if (not conn_result.success) then:
    print("❌ Failed to connect: " + conn_result.error)
    exit(1)
end

conn = conn_result.conn

// Global statistics
static daily_orders = 0
static daily_revenue = 0.0
static active_users = 0
static low_stock_products = []

// Multi-table callback function
fun onEcommerceChange(channel, payload):
    data = json_parse(payload)
    table = data.table
    operation = data.operation
    
    if (table == "orders") then:
        if (operation == "INSERT") then:
            daily_orders = daily_orders + 1
            daily_revenue = daily_revenue + data.data.total_amount
            
            print("🛒 New order #" + str(data.data.id) + " - $" + str(data.data.total_amount))
            print("📊 Daily totals: " + str(daily_orders) + " orders, $" + str(daily_revenue) + " revenue")
            
            // Send notification for large orders
            if (data.data.total_amount > 1000) then:
                print("💰 HIGH VALUE ORDER: $" + str(data.data.total_amount) + " from customer " + str(data.data.customer_id))
            end
        end
        
    elif (table == "users") then:
        if (operation == "INSERT") then:
            print("👤 New user registered: " + data.data.email)
        elif (operation == "UPDATE" and data.data.last_login != data.old_data.last_login) then:
            print("🔐 User login: " + data.data.email)
        end
        
    elif (table == "products") then:
        if (operation == "UPDATE" and data.data.stock_quantity != data.old_data.stock_quantity) then:
            product_name = data.data.name
            new_stock = data.data.stock_quantity
            
            if (new_stock <= 5) then:
                if (not contains(low_stock_products, product_name)) then:
                    append(low_stock_products, product_name)
                    print("⚠️ CRITICAL STOCK: " + product_name + " has only " + str(new_stock) + " items left!")
                end
            elif (new_stock > 5 and contains(low_stock_products, product_name)) then:
                low_stock_products = filter(low_stock_products, fun(p): return p != product_name end)
                print("✅ Stock replenished: " + product_name + " now has " + str(new_stock) + " items")
            end
        end
        
    elif (table == "reviews") then:
        if (operation == "INSERT") then:
            rating = data.data.rating
            product_id = data.data.product_id
            
            if (rating <= 2) then:
                print("⭐ Low rating alert: Product #" + str(product_id) + " received " + str(rating) + "/5 stars")
            elif (rating >= 5) then:
                print("🌟 Excellent review: Product #" + str(product_id) + " received 5/5 stars!")
            end
        end
    end
end

// Start multi-table streaming
tables_to_monitor = ["orders", "users", "products", "reviews"]
stream_id = stream_tables(conn, tables_to_monitor, onEcommerceChange)

print("🔄 Started monitoring e-commerce tables: " + str(tables_to_monitor))
print("📊 Real-time dashboard active...")

// Keep the program running
while (true):
    sleep(1000)
end
```

### Financial Transaction Monitoring

```uddin
// Connect to PostgreSQL database
conn_result = db_connect("postgres", "localhost", 5432, "banking", "postgres", "password")

if (not conn_result.success) then:
    print("❌ Failed to connect: " + conn_result.error)
    exit(1)
end

conn = conn_result.conn

// Fraud detection thresholds
MAX_TRANSACTION_AMOUNT = 10000.0
MAX_DAILY_TRANSACTIONS = 20
SUSPICIOUS_COUNTRIES = ["XX", "YY", "ZZ"]

// Transaction tracking
static daily_transactions = {}
static suspicious_activities = []

// Multi-table callback for financial monitoring
fun onFinancialChange(channel, payload):
    data = json_parse(payload)
    table = data.table
    operation = data.operation
    
    if (table == "transactions") then:
        if (operation == "INSERT") then:
            amount = data.data.amount
            user_id = str(data.data.user_id)
            country = data.data.country_code
            
            print("💳 New transaction: $" + str(amount) + " by user " + user_id)
            
            // Track daily transactions per user
            if (not has_key(daily_transactions, user_id)) then:
                daily_transactions[user_id] = 0
            end
            daily_transactions[user_id] = daily_transactions[user_id] + 1
            
            // Fraud detection checks
            suspicious = false
            
            // Check for large amounts
            if (amount > MAX_TRANSACTION_AMOUNT) then:
                print("🚨 LARGE TRANSACTION ALERT: $" + str(amount) + " by user " + user_id)
                suspicious = true
            end
            
            // Check for too many transactions
            if (daily_transactions[user_id] > MAX_DAILY_TRANSACTIONS) then:
                print("🚨 HIGH FREQUENCY ALERT: User " + user_id + " has made " + str(daily_transactions[user_id]) + " transactions today")
                suspicious = true
            end
            
            // Check for suspicious countries
            if (contains(SUSPICIOUS_COUNTRIES, country)) then:
                print("🚨 LOCATION ALERT: Transaction from suspicious country " + country + " by user " + user_id)
                suspicious = true
            end
            
            if (suspicious) then:
                append(suspicious_activities, {
                    "user_id": user_id,
                    "amount": amount,
                    "country": country,
                    "timestamp": data.timestamp
                })
            end
        end
        
    elif (table == "accounts") then:
        if (operation == "UPDATE") then:
            old_balance = data.old_data.balance
            new_balance = data.data.balance
            account_id = str(data.data.id)
            
            if (new_balance < 0 and old_balance >= 0) then:
                print("⚠️ OVERDRAFT ALERT: Account " + account_id + " is now overdrawn ($" + str(new_balance) + ")")
            elif (new_balance >= 0 and old_balance < 0) then:
                print("✅ Account " + account_id + " is no longer overdrawn")
            end
        end
        
    elif (table == "users") then:
        if (operation == "UPDATE" and data.data.status != data.old_data.status) then:
            user_id = str(data.data.id)
            new_status = data.data.status
            
            if (new_status == "suspended") then:
                print("🔒 User " + user_id + " account suspended")
            elif (new_status == "active" and data.old_data.status == "suspended") then:
                print("🔓 User " + user_id + " account reactivated")
            end
        end
    end
end

// Start multi-table streaming
financial_tables = ["transactions", "accounts", "users"]
stream_id = stream_tables(conn, financial_tables, onFinancialChange)

print("🔄 Started monitoring financial tables: " + str(financial_tables))
print("🛡️ Fraud detection system active...")

// Keep the program running
while (true):
    sleep(1000)
end
```

## Advanced Examples

### Resilient Streaming with Error Handling

```uddin
// Robust connection setup with retry logic
fun setup_resilient_connection(config):
    max_retries = 3
    retry_delay = 5000  // 5 seconds
    
    for (i in range(max_retries)):
        print("🔄 Attempting database connection (attempt " + str(i + 1) + "/" + str(max_retries) + ")...")
        
        conn_result = db_connect(
            config.driver, config.host, config.port,
            config.database, config.username, config.password
        )
        
        if (conn_result.success) then:
            print("✅ Database connected successfully")
            return conn_result.conn
        else:
            print("❌ Connection failed: " + conn_result.error)
            if (i < max_retries - 1) then:
                print("⏳ Waiting " + str(retry_delay / 1000) + " seconds before retry...")
                sleep(retry_delay)
            end
        end
    end
    
    print("💥 Failed to connect after " + str(max_retries) + " attempts")
    return null
end

// Robust callback with error handling
fun robust_callback(channel, payload):
    try:
        data = json_parse(payload)
        
        // Validate required fields
        if (not has_key(data, "operation") or not has_key(data, "table")) then:
            print("⚠️ Invalid payload structure: missing required fields")
            return
        end
        
        // Process the change
        process_database_change(data)
        
    catch (json_error):
        print("❌ JSON parsing error: " + str(json_error))
        print("📄 Raw payload: " + payload)
    catch (processing_error):
        print("❌ Processing error: " + str(processing_error))
    end
end

fun process_database_change(data):
    operation = data.operation
    table = data.table
    
    print("📝 Processing " + operation + " on " + table + " table")
    
    // Add your business logic here
    if (table == "orders" and operation == "INSERT") then:
        // Handle new order
        handle_new_order(data.data)
    elif (table == "users" and operation == "UPDATE") then:
        // Handle user update
        handle_user_update(data.data, data.old_data)
    end
end

fun handle_new_order(order_data):
    print("🛒 Processing new order: #" + str(order_data.id))
    // Add order processing logic
end

fun handle_user_update(new_data, old_data):
    print("👤 Processing user update: " + new_data.email)
    // Add user update logic
end

// Main program with error handling
config = {
    "driver": "postgres",
    "host": "localhost",
    "port": 5432,
    "database": "myapp",
    "username": "postgres",
    "password": "password"
}

conn = setup_resilient_connection(config)
if (conn == null) then:
    exit(1)
end

try:
    // Start streaming with error handling
    tables = ["orders", "users", "products"]
    stream_id = stream_tables(conn, tables, robust_callback)
    
    print("🔄 Started resilient streaming for tables: " + str(tables))
    print("🛡️ Error handling enabled")
    
    // Keep running with periodic health checks
    while (true):
        sleep(10000)  // Check every 10 seconds
        
        // Perform health check (optional)
        try:
            result = db_query(conn, "SELECT 1")
            if (not result.success) then:
                print("⚠️ Database health check failed: " + result.error)
            end
        catch (health_error):
            print("❌ Health check error: " + str(health_error))
        end
    end
    
catch (streaming_error):
    print("💥 Streaming setup failed: " + str(streaming_error))
    exit(1)
finally:
    // Cleanup
    if (stream_id != null) then:
        db_stop_stream(stream_id)
        print("🛑 Streaming stopped")
    end
    
    if (conn != null) then:
        db_close(conn)
        print("🔌 Database connection closed")
    end
end
```

### Performance-Optimized Batch Processing

```uddin
// Configuration
BATCH_SIZE = 100
BATCH_TIMEOUT = 5000  // 5 seconds
MAX_MEMORY_USAGE = 1000000  // 1MB

// Global batch storage
static change_batch = []
static last_batch_process = time_now()
static total_memory_usage = 0

// Optimized callback with batching
fun optimized_callback(channel, payload):
    try:
        data = json_parse(payload)
        
        // Add to batch
        append(change_batch, data)
        total_memory_usage = total_memory_usage + len(payload)
        
        current_time = time_now()
        time_since_last_batch = current_time - last_batch_process
        
        // Process batch if conditions are met
        should_process = (
            len(change_batch) >= BATCH_SIZE or
            time_since_last_batch >= BATCH_TIMEOUT or
            total_memory_usage >= MAX_MEMORY_USAGE
        )
        
        if (should_process) then:
            process_batch()
        end
        
    catch (error):
        print("❌ Callback error: " + str(error))
    end
end

fun process_batch():
    if (len(change_batch) == 0) then:
        return
    end
    
    start_time = time_now()
    batch_size = len(change_batch)
    
    print("📦 Processing batch of " + str(batch_size) + " changes...")
    
    // Group changes by table and operation
    grouped_changes = group_changes_by_table(change_batch)
    
    // Process each group
    for (table in keys(grouped_changes)):
        table_changes = grouped_changes[table]
        print("  📋 Processing " + str(len(table_changes)) + " changes for " + table + " table")
        
        // Process table-specific changes
        process_table_changes(table, table_changes)
    end
    
    // Clear batch
    change_batch = []
    total_memory_usage = 0
    last_batch_process = time_now()
    
    processing_time = time_now() - start_time
    print("✅ Batch processed in " + str(processing_time) + "ms")
end

fun group_changes_by_table(changes):
    grouped = {}
    
    for (change in changes):
        table = change.table
        
        if (not has_key(grouped, table)) then:
            grouped[table] = []
        end
        
        append(grouped[table], change)
    end
    
    return grouped
end

fun process_table_changes(table, changes):
    // Count operations
    inserts = filter(changes, fun(c): return c.operation == "INSERT" end)
    updates = filter(changes, fun(c): return c.operation == "UPDATE" end)
    deletes = filter(changes, fun(c): return c.operation == "DELETE" end)
    
    if (len(inserts) > 0) then:
        print("    ➕ " + str(len(inserts)) + " inserts")
    end
    
    if (len(updates) > 0) then:
        print("    ✏️ " + str(len(updates)) + " updates")
    end
    
    if (len(deletes) > 0) then:
        print("    🗑️ " + str(len(deletes)) + " deletes")
    end
    
    // Add your batch processing logic here
    // For example: bulk updates, cache invalidation, etc.
end

// Setup optimized streaming
conn_result = db_connect("postgres", "localhost", 5432, "myapp", "postgres", "password")

if (not conn_result.success) then:
    print("❌ Failed to connect: " + conn_result.error)
    exit(1)
end

conn = conn_result.conn

// Start optimized streaming
tables = ["orders", "users", "products", "inventory"]
stream_id = stream_tables(conn, tables, optimized_callback)

print("🔄 Started optimized streaming with batching")
print("📊 Batch size: " + str(BATCH_SIZE) + ", Timeout: " + str(BATCH_TIMEOUT) + "ms")

// Keep running with periodic batch processing
while (true):
    sleep(1000)
    
    // Force process batch if timeout reached
    if (len(change_batch) > 0 and (time_now() - last_batch_process) >= BATCH_TIMEOUT) then:
        process_batch()
    end
end
```

These examples demonstrate various patterns and best practices for implementing real-time database streaming in UddinLang applications, from basic monitoring to advanced production-ready systems with error handling and performance optimization.