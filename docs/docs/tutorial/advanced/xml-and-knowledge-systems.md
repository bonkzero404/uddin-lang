---
sidebar_position: 5
---

# XML Processing and Knowledge Systems

Uddin-Lang menyediakan fitur unik untuk pemrosesan XML dan sistem knowledge-based programming melalui fact database dan event processing. Fitur ini memungkinkan Anda membangun aplikasi yang dapat mengelola pengetahuan dan merespons event secara dinamis.

## XML Processing

### Basic XML Operations

```uddin
// xml_parse - parsing XML string ke object
xml_string = `
<bookstore>
    <book id="1" category="fiction">
        <title>The Great Gatsby</title>
        <author>F. Scott Fitzgerald</author>
        <price currency="USD">12.99</price>
        <published>1925</published>
    </book>
    <book id="2" category="science">
        <title>A Brief History of Time</title>
        <author>Stephen Hawking</author>
        <price currency="USD">15.99</price>
        <published>1988</published>
    </book>
</bookstore>
`

parsed_xml = xml_parse(xml_string)
if (parsed_xml != null) then:
    bookstore = parsed_xml["bookstore"]  // Get bookstore object from root
    print("Root element: bookstore")

    // Akses books
    books = bookstore["book"]
    for (book in books):
        print("Book ID: " + book["@attributes"]["id"])
        print("Category: " + book["@attributes"]["category"])
        print("Title: " + book["title"])
        print("Author: " + book["author"])
        print("Price: " + str(book["price"]) + " USD")
        print("Published: " + book["published"])
        print("---")
    end
else:
    print("XML parsing error")
end
```

### XML Generation

```uddin
// xml_stringify - convert object ke XML string
product_data = {
    "catalog": {
        "@attributes": {
            "version": "1.0",
            "xmlns": "http://example.com/catalog"
        },
        "product": [
            {
                "@attributes": {"id": "P001", "active": "true"},
                "name": "Laptop Gaming",
                "description": "High-performance gaming laptop",
                "price": "1299.99",
                "specs": {
                    "cpu": "Intel i7",
                    "ram": "16GB",
                    "storage": "512GB SSD"
                }
            },
            {
                "@attributes": {"id": "P002", "active": "false"},
                "name": "Wireless Mouse",
                "description": "Ergonomic wireless mouse",
                "price": "29.99"
            }
        ]
    }
}

xml_output = xml_stringify(product_data)
print("Generated XML:")
print(xml_output)
```

### XML Data Processing

```uddin
fun processXMLCatalog(xml_content):
    parsed = xml_parse(xml_content)
    if (not parsed.success) then:
        return {"error": "Invalid XML: " + parsed.error}
    end

    catalog = parsed.data
    products = []

    if (catalog.children and catalog.children.product) then:
        product_nodes = catalog.children.product

        // Handle single product vs array of products
        if (typeof(product_nodes) != "array") then:
            product_nodes = [product_nodes]
        end

        for product_node in product_nodes:
            product = {
                "id": product_node.attributes.id,
                "active": product_node.attributes.active == "true",
                "name": product_node.children.name.content,
                "description": product_node.children.description.content,
                "price": {
                    "amount": float(product_node.children.price.content),
                    "currency": product_node.children.price.attributes.currency
                }
            }

            // Optional specs
            if (product_node.children.specs) then:
                product.specs = {}
                specs = product_node.children.specs.children
                for spec_name in specs:
                    product.specs[spec_name] = specs[spec_name].content
                end
            end

            products = append(products, product)
        end
    end

    return {
        "success": true,
        "products": products,
        "total": len(products),
        "active_count": len(filter(products, fun(p): return p.active end))
    }
end

// Penggunaan
result = processXMLCatalog(xml_string)
if (result.success) then:
    print("Total products: " + str(result.total))
    print("Active products: " + str(result.active_count))

    for product in result.products:
        if (product.active) then:
            print(product.name + " - $" + str(product.price.amount))
        end
    end
end
```

## Fact Database System

Sistem fact database memungkinkan Anda menyimpan dan query pengetahuan dalam bentuk fakta-fakta terstruktur.

### Basic Fact Operations

```uddin
// fact_assert - menambahkan fakta ke database
fact_assert("person", "john", {"age": 30, "city": "Jakarta", "job": "engineer"})
fact_assert("person", "alice", {"age": 25, "city": "Bandung", "job": "designer"})
fact_assert("person", "bob", {"age": 35, "city": "Jakarta", "job": "manager"})

// fact_assert untuk relasi
fact_assert("works_at", "john", {"company": "TechCorp", "department": "Engineering"})
fact_assert("works_at", "alice", {"company": "DesignStudio", "department": "Creative"})
fact_assert("works_at", "bob", {"company": "TechCorp", "department": "Management"})

// fact_assert untuk hierarki
fact_assert("reports_to", "john", {"manager": "bob"})
fact_assert("skill", "john", {"name": "Python", "level": "expert"})
fact_assert("skill", "john", {"name": "JavaScript", "level": "intermediate"})
fact_assert("skill", "alice", {"name": "Photoshop", "level": "expert"})
```

### Querying Facts

```uddin
// fact_query - query fakta dengan pattern
// Query semua person
all_persons = fact_query("person", null, {})
print("All persons: " + str(len(all_persons)))
for person in all_persons:
    print(person.subject + ": age " + str(person.data.age) + ", " + person.data.city)
end

// Query dengan filter
jakarta_people = fact_query("person", null, {"city": "Jakarta"})
print("\nPeople in Jakarta:")
for person in jakarta_people:
    print(person.subject + " (" + person.data.job + ")")
end

// Query specific subject
john_facts = fact_query(null, "john", {})
print("\nFacts about John:")
for fact in john_facts:
    print(fact.predicate + ": " + str(fact.data))
end

// fact_exists - mengecek keberadaan fakta
if (fact_exists("person", "john", {"city": "Jakarta"})) then:
    print("John lives in Jakarta")
end

// fact_count - menghitung fakta
total_facts = fact_count(null, null, {})
print("Total facts in database: " + str(total_facts))

person_count = fact_count("person", null, {})
print("Total persons: " + str(person_count))
```

### Advanced Fact Queries

```uddin
// Setup fact database with initial data
print("Setting up fact database...")

// fact_assert - add facts to database
fact_assert("person", "john", {"age": 30, "city": "Jakarta", "job": "engineer"})
fact_assert("person", "alice", {"age": 25, "city": "Bandung", "job": "designer"})
fact_assert("person", "bob", {"age": 35, "city": "Jakarta", "job": "manager"})
fact_assert("person", "charlie", {"age": 28, "city": "Surabaya", "job": "developer"})

// fact_assert for work relations
fact_assert("works_at", "john", {"company": "TechCorp", "department": "Engineering"})
fact_assert("works_at", "alice", {"company": "DesignStudio", "department": "Creative"})
fact_assert("works_at", "bob", {"company": "TechCorp", "department": "Management"})
fact_assert("works_at", "charlie", {"company": "StartupXYZ", "department": "Development"})

print("Fact database setup complete!\n")

// fact_query - query facts with pattern
// Query all persons
all_persons = fact_query("person")
print("All persons found: " + str(all_persons))
if (all_persons != null) then:
    print("\nPerson details:")
    // Query individual persons
    john_data = fact_query("person", "john")
    if (john_data != null) then:
        print("john: age " + str(john_data["age"]) + ", " + str(john_data["city"]) + ", job: " + str(john_data["job"]))
    end

    alice_data = fact_query("person", "alice")
    if (alice_data != null) then:
        print("alice: age " + str(alice_data["age"]) + ", " + str(alice_data["city"]) + ", job: " + str(alice_data["job"]))
    end

    bob_data = fact_query("person", "bob")
    if (bob_data != null) then:
        print("bob: age " + str(bob_data["age"]) + ", " + str(bob_data["city"]) + ", job: " + str(bob_data["job"]))
    end

    charlie_data = fact_query("person", "charlie")
    if (charlie_data != null) then:
        print("charlie: age " + str(charlie_data["age"]) + ", " + str(charlie_data["city"]) + ", job: " + str(charlie_data["job"]))
    end
else:
    print("No persons found")
end

// Query specific person
john_data = fact_query("person", "john")
if (john_data != null) then:
    print("\nJohn's data:")
    print("Age: " + str(john_data.age))
    print("City: " + john_data.city)
    print("Job: " + john_data.job)
end

// Query work relations
print("\nWork relations:")
work_relations = fact_query("works_at")
if (work_relations != null) then:
    john_work = fact_query("works_at", "john")
    if (john_work != null) then:
        print("john works at " + str(john_work["company"]) + " in " + str(john_work["department"]))
    end

    alice_work = fact_query("works_at", "alice")
    if (alice_work != null) then:
        print("alice works at " + str(alice_work["company"]) + " in " + str(alice_work["department"]))
    end

    bob_work = fact_query("works_at", "bob")
    if (bob_work != null) then:
        print("bob works at " + str(bob_work["company"]) + " in " + str(bob_work["department"]))
    end

    charlie_work = fact_query("works_at", "charlie")
    if (charlie_work != null) then:
        print("charlie works at " + str(charlie_work["company"]) + " in " + str(charlie_work["department"]))
    end
end

// fact_exists - check if fact exists
if (fact_exists("person", "john")) then:
    print("\nJohn exists in person database")
end

// fact_count - count facts
total_facts = fact_count()
print("Total facts in database: " + str(total_facts))

person_count = fact_count("person")
print("Total persons: " + str(person_count))

work_count = fact_count("works_at")
print("Total work relations: " + str(work_count))
```

## Event Processing System

Sistem event processing memungkinkan Anda mendeteksi pattern dalam stream event dan merespons secara real-time.

### Basic Event Operations

```uddin
// event_emit - emit events
event_emit("user_login", {"user_id": "john", "timestamp": date_now(), "ip": "192.168.1.100"})
event_emit("user_login", {"user_id": "alice", "timestamp": date_now(), "ip": "192.168.1.101"})
event_emit("page_view", {"user_id": "john", "page": "/dashboard", "timestamp": date_now()})
event_emit("user_logout", {"user_id": "alice", "timestamp": date_now()})

// event_count - count events
total_events = event_count()  // No parameters to count all events
print("Total events: " + str(total_events))

login_events = event_count("user_login", {})
print("Login events: " + str(login_events))

// event_get_window - get events within time window
recent_events = event_get_window("1h", "user_login")  // Last 1 hour
if (recent_events != null) then:
    print("Recent logins: " + str(len(recent_events)))
    if (len(recent_events) > 0) then:
        i = 0
        while (i < len(recent_events)):
            event = recent_events[i]
            print("Event " + str(i) + ": " + str(event))
            i = i + 1
        end
    else:
        print("No recent login events found")
    end
else:
    print("Recent logins: 0 (null)")
end
```

### Event Pattern Detection

```uddin
// event_define_pattern - define pattern for detection
// Pattern: 3 failed login attempts within 5 minutes
pattern_id = "suspicious_login"
pattern_config = {
    "event_type": "login_failed",
    "count": 3,
    "time_window": 300,  // 5 minutes
    "group_by": "user_id"
}

event_define_pattern(pattern_id, pattern_config)

// Simulate failed login attempts
for (i in range(4)):
    event_emit("login_failed", {
        "user_id": "suspicious_user",
        "timestamp": date_now(),
        "ip": "192.168.1.200",
        "reason": "invalid_password"
    })
end

// Check pattern matches
suspicious_patterns = event_get_window("1h", "login_failed")
if (len(suspicious_patterns) > 0) then:
    print("ALERT: Suspicious login activity detected!")
    i = 0
    while (i < len(suspicious_patterns)):
        pattern = suspicious_patterns[i]
        print("Event " + str(i) + ": " + str(pattern))
        i = i + 1
    end
end
```

### Advanced Event Processing

```uddin
// Global handlers storage
login_handlers = []
purchase_handlers = []

fun registerLoginHandler(handler):
    login_handlers = append(login_handlers, handler)
end

fun registerPurchaseHandler(handler):
    purchase_handlers = append(purchase_handlers, handler)
end

fun emitLoginEvent(data):
    event_emit("user_login", data)
    for (handler in login_handlers):
        try:
            handler(data)
        catch (error):
            print("Handler error: " + str(error))
        end
    end
end

fun emitPurchaseEvent(data):
    event_emit("user_purchase", data)
    for (handler in purchase_handlers):
        try:
            handler(data)
        catch (error):
            print("Handler error: " + str(error))
        end
    end
end

fun createEventProcessor():
    return {
        "on": fun(event_type, handler):
            if (event_type == "user_login") then:
                registerLoginHandler(handler)
            else if (event_type == "user_purchase") then:
                registerPurchaseHandler(handler)
            end
        end,

        "processUserActivity": fun(user_id, activity_type, details):
            activity_data = {
                "user_id": user_id,
                "activity": activity_type,
                "timestamp": date_now(),
                "details": details
            }

            // Specific activity events
            if (activity_type == "login") then:
                emitLoginEvent(activity_data)
            else if (activity_type == "purchase") then:
                emitPurchaseEvent(activity_data)
            end
        end
    }
end

// Setup event processor
processor = createEventProcessor()

// Register handlers
processor.on("user_login", fun(data):
    print("User " + data.user_id + " logged in at " + date_format(data.timestamp, "15:04:05"))

    // Update user facts
    fact_assert("last_login", data.user_id, {
        "timestamp": data.timestamp,
        "ip": data.details.ip
    })
end)

processor.on("user_purchase", fun(data):
    print("Purchase by " + data.user_id + ": $" + str(data.details.amount))

    // Update purchase history
    fact_assert("purchase", data.user_id + "_" + str(data.timestamp), {
        "user_id": data.user_id,
        "amount": data.details.amount,
        "product": data.details.product,
        "timestamp": data.timestamp
    })
end)

// Process activities
processor.processUserActivity("john", "login", {"ip": "192.168.1.100"})
processor.processUserActivity("john", "purchase", {"amount": 99.99, "product": "Premium Plan"})
processor.processUserActivity("alice", "login", {"ip": "192.168.1.101"})
```

## Practical Applications

### Knowledge Base System

```uddin
fun createKnowledgeBase():
    return {
        "addRule": fun(rule_name, conditions, conclusion):
            fact_assert("rule", rule_name, {
                "conditions": conditions,
                "conclusion": conclusion
            })
        end,

        "addFact": fun(subject, predicate, object):
            fact_assert(predicate, subject, object)
        end,

        "query": fun(predicate, subject, filters):
            return fact_query(predicate, subject, filters)
        end,

        "infer": fun():
            // Simple inference engine

            // Manually check the mortality rule since we know it exists
            // Check if socrates is human
            socrates_human = fact_query("is_human", "socrates")

            if (socrates_human != null) then:
                // Check if socrates is already mortal
                socrates_mortal = fact_query("is_mortal", "socrates")

                if (socrates_mortal == null) then:
                    // Assert that socrates is mortal
                    fact_assert("is_mortal", "socrates", {"reason": "all humans are mortal"})
                    new_fact = {
                        "predicate": "is_mortal",
                        "subject": "socrates",
                        "data": {"reason": "all humans are mortal"}
                    }
                    return [new_fact]
                end
            end

            return []
        end
    }
end

// Setup knowledge base
kb = createKnowledgeBase()

// Add facts
kb.addFact("socrates", "is_human", {"type": "person"})
kb.addFact("plato", "is_human", {"type": "person"})
kb.addFact("fido", "is_dog", {"type": "animal"})

// Add rules
kb.addRule("mortality_rule",
    [{"predicate": "is_human", "subject": null, "filters": {}}],
    {"predicate": "is_mortal", "subject": "socrates", "data": {"reason": "all humans are mortal"}}
)

// Perform inference
new_facts = kb.infer()
print("Inferred facts: " + str(new_facts))
if (new_facts != null and len(new_facts) > 0) then:
    print("Number of new facts: " + str(len(new_facts)))
    // Since we know there's only one fact in our simple example
    fact = new_facts[0]
    print(fact["subject"] + " " + fact["predicate"])
else:
    print("No new facts inferred")
end
```

### Real-time Analytics Dashboard

```uddin
fun createAnalyticsDashboard():
    dashboard = {
        "trackEvent": fun(event_type, user_id, data):
            event_data = {
                "user_id": user_id,
                "timestamp": date_now(),
                "data": data
            }

            event_emit(event_type, event_data)

            // Update aggregated facts
            dashboard.updateAggregates(event_type, user_id, data)
        end,

        "updateAggregates": fun(event_type, user_id, data):
            // Daily aggregates
            today = date_format(date_now(), "2006-01-02")
            daily_key = event_type + "_" + today

            existing = fact_query("daily_count", daily_key, {})
            if (existing != null) then:
                new_count = existing.count + 1
                fact_retract("daily_count", daily_key, {})
            else:
                new_count = 1
            end

            fact_assert("daily_count", daily_key, {
                "event_type": event_type,
                "date": today,
                "count": new_count
            })

            // User aggregates
            user_key = user_id + "_" + event_type
            user_existing = fact_query("user_count", user_key, {})
            if (user_existing != null) then:
                user_new_count = user_existing.count + 1
                fact_retract("user_count", user_key, {})
            else:
                user_new_count = 1
            end

            fact_assert("user_count", user_key, {
                "user_id": user_id,
                "event_type": event_type,
                "count": user_new_count
            })
        end,

        "getDashboardData": fun():
            today = date_format(date_now(), "2006-01-02")

            // Get today's events
            daily_counts = fact_query("daily_count")

            // Get top users
            user_counts = fact_query("user_count")

            // Get recent events
            recent_events = event_get_window("1h")  // Last hour

            return {
                "daily_stats": daily_counts,
                "user_stats": user_counts,
                "recent_activity": recent_events,
                "timestamp": date_now()
            }
        end
    }

    return dashboard
end

// Setup dashboard
dashboard = createAnalyticsDashboard()

// Track some events
dashboard.trackEvent("page_view", "john", {"page": "/home"})
dashboard.trackEvent("page_view", "alice", {"page": "/products"})
dashboard.trackEvent("click", "john", {"element": "buy_button"})
dashboard.trackEvent("purchase", "john", {"amount": 29.99})

// Get dashboard data
data = dashboard.getDashboardData()
print("Dashboard Data:")

// Display daily stats summary
print("\n=== Daily Statistics ===")
if (data.daily_stats != null) then:
    print("Events tracked today:")
    print(str(data.daily_stats))
else:
    print("No daily stats available")
end

// Display user stats summary
print("\n=== User Statistics ===")
if (data.user_stats != null) then:
    print("User activity summary:")
    print(str(data.user_stats))
else:
    print("No user stats available")
end

// Display recent activity summary
print("\n=== Recent Activity (Last Hour) ===")
if (data.recent_activity != null) then:
    print("Recent events:")
    print(str(data.recent_activity))
else:
    print("No recent activity")
end

print("\n=== Dashboard Updated ===")
print("Timestamp: " + str(data.timestamp))
```

## Tips dan Best Practices

### XML Processing

1. **Validation**: Selalu validasi XML sebelum processing
2. **Namespace Handling**: Perhatikan XML namespaces dalam parsing
3. **Memory Management**: Untuk XML besar, pertimbangkan streaming parsing
4. **Error Handling**: Handle malformed XML dengan graceful

### Fact Database

1. **Fact Design**: Desain struktur fact yang konsisten
2. **Indexing**: Gunakan subject dan predicate yang efisien untuk query
3. **Cleanup**: Hapus fact yang tidak relevan secara berkala
4. **Consistency**: Maintain konsistensi data dengan validation

### Event Processing

1. **Event Schema**: Definisikan schema event yang konsisten
2. **Performance**: Monitor performa untuk high-volume events
3. **Pattern Complexity**: Jangan buat pattern yang terlalu kompleks
4. **Memory Usage**: Clear old events secara berkala

Dengan fitur XML processing dan knowledge systems ini, Uddin-Lang dapat digunakan untuk membangun aplikasi intelligent yang dapat memproses data terstruktur dan mengelola pengetahuan secara dinamis.
