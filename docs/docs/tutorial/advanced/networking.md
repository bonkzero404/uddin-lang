---
sidebar_position: 3
---

# Networking and HTTP

Uddin-Lang provides comprehensive networking support, including HTTP client/server, TCP/UDP sockets, and other network utilities. These features allow you to build web applications, APIs, and network services.

## HTTP Client Functions

### Basic HTTP Requests

```uddin
import "http"

// GET request
response = http.get("https://api.github.com/users/octocat")
print("Status: " + str(response.status))
print("Body: " + str(response.body))
print("Headers: " + str(response.headers))
```

### HTTP Methods

```uddin
import "http"

// GET request
get_response = http.get("https://httpbin.org/get")

// POST request dengan data
post_data = {"name": "John", "age": 30}
post_response = http.post("https://httpbin.org/post", post_data)

// PUT request
put_data = {"id": 1, "name": "Updated Name"}
put_response = http.put("https://httpbin.org/put", put_data)

// DELETE request
delete_response = http.delete("https://httpbin.org/delete")
```

### Advanced HTTP Requests

```uddin
import "http"
import "json"

// Custom HTTP request dengan headers
headers = {
    "Authorization": "Bearer your-token",
    "Content-Type": "application/json",
    "User-Agent": "Uddin-Lang/1.0"
}

data = {"message": "Hello API"}

response = http.request("POST", "https://api.example.com/data", data, headers)

if (response.status == 200) then:
    print("Request successful!")
    result = json.parse(response.body)
    print(result)
else:
    print("Request failed: " + str(response.status))
end
```

## HTTP Server Functions

### Basic HTTP Server

```uddin
import "http"

// Create HTTP server
server_id = "my-server"
port = 8080

// Start server
http.server_start(port, server_id)
print("Server running on port " + str(port))

// Route handler function
fun homeHandler(req, res):
    http.response(res, 200, {"Content-Type": "text/plain"}, "Hello, World!")
end

// Register route
http.server_route("GET", "/", homeHandler, server_id)

print("Server ready to accept requests...")
print("Access http://localhost:" + str(port))

// Keep server running indefinitely
while (true):
    // Server is running in background
    // Use Ctrl+C to stop the server
end
```

### Advanced HTTP Server with Multiple Routes

```uddin
import "http"
import "json"
import "datetime"

server_id = "api-server"
port = 8080

// Start server
http.server_start(port, server_id)

// Handler for GET /users
fun getUsersHandler(request, response):
    users = [
        {"id": 1, "name": "Alice", "email": "alice@example.com"},
        {"id": 2, "name": "Bob", "email": "bob@example.com"}
    ]

    headers = {"Content-Type": "application/json"}
    http.response(response, 200, headers, json.stringify(users))
end

// Handler for POST /users
fun createUserHandler(request, response):
    try:
        user_data = json.parse(request["body"])

        // Validate data - check if properties exist and are not empty
        if (not("name" in user_data) or not("email" in user_data)) then:
            error_response = {"error": "Name and email are required"}
            return http.response(response, 400, {"Content-Type": "application/json"}, json.stringify(error_response))
        end


        // Simulate user creation
        new_user = {
            id: random_int(1000, 9999),
            name: user_data.name,
            email: user_data.email,
            created_at: datetime.now()
        }

        return http.response(response, 201, {"Content-Type": "application/json"}, json.stringify(new_user))

    catch (error):
        error_response = {"error": "Invalid JSON data"}
        return http.response(response, 400, {"Content-Type": "application/json"}, json.stringify(error_response))
    end
end

// Handler for GET /health
fun healthHandler(request, response):
    health_data = {
        "status": "healthy",
        "timestamp": datetime.now(),
        "uptime": "running"
    }

    http.response(response, 200, {"Content-Type": "application/json"}, json.stringify(health_data))
end

// Register routes
http.server_route("GET", "/api/users", getUsersHandler, server_id)
http.server_route("POST", "/api/users", createUserHandler, server_id)
http.server_route("GET", "/api/health", healthHandler, server_id)

print("API Server running at http://localhost:" + str(port))
print("Endpoints:")
print("  GET  /users  - List users")
print("  POST /users  - Create new user")
print("  GET  /health - Health check")

while (true):
    // Server is running in background
    // Use Ctrl+C to stop the server
end
```

## TCP Socket Programming

### TCP Client

```uddin
// Connect to TCP server
connection = tcp_connect("localhost", 8080)

if (connection.success) then:
    print("Connected to server")

    // Send data
    message = "Hello from TCP client!"
    tcp_write(connection.socket, message)

    // Read response
    response = tcp_read(connection.socket, 1024)
    print("Response from server: " + response)

    // Close connection
    tcp_close(connection.socket)
else:
    print("Failed to connect: " + connection.error)
end
```

### TCP Server

```uddin
// Create TCP server
server = tcp_listen(8080)

if (server.success) then:
    print("TCP Server listening on port 8080")

    while (true):
        // Accept connection
        client = tcp_accept(server.socket)

        if (client.success) then:
            print("Client connected: " + client.address)

            // Read data from client
            data = tcp_read(client.socket, 1024)
            print("Received: " + data)

            // Send response
            response = "Echo: " + data
            tcp_write(client.socket, response)

            // Close client connection
            tcp_close(client.socket)
        end
    end
else:
    print("Failed to create server: " + server.error)
end
```

## UDP Socket Programming

### UDP Client

```uddin
// Create UDP socket
socket = udp_connect("localhost", 9090)

if (socket.success) then:
    // Send data
    message = "Hello UDP!"
    udp_write(socket.socket, message, "localhost", 9090)

    // Read response
    response = udp_read(socket.socket, 1024)
    print("Response: " + response.data)
    print("From: " + response.address)

    udp_close(socket.socket)
end
```

### UDP Server

```uddin
// Create UDP server
server = udp_listen(9090)

if (server.success) then:
    print("UDP Server listening on port 9090")

    while (true):
        // Read data
        data = udp_read(server.socket, 1024)
        print("Received from " + data.address + ": " + data.data)

        // Send response
        response = "Echo: " + data.data
        udp_write(server.socket, response, data.address, data.port)
    end
end
```

## Network Utilities

### DNS Resolution

```uddin
import "http"

// Resolve hostname to IP
result = http.net_resolve("google.com")
if (result.success) then:
    print("IP addresses for google.com:")
    for (ip in result.ips):
        print("  " + ip)
    end
else:
    print("DNS resolution failed: " + result.error)
end
```

### Network Ping

```uddin
import "http"

// Ping host (using port 53 for DNS and 5000ms timeout)
ping_result = http.net_ping("8.8.8.8", 53, 5000)

if (ping_result.success) then:
    print("Ping successful!")
    print("  Response time: " + str(ping_result.time_ms) + "ms")
else:
    print("Ping failed: " + ping_result.error)
end

```

## JSON Processing

### JSON Parsing and Serialization

```uddin
import "json"

// Parse JSON string
json_string = '{"name": "John", "age": 30, "city": "New York"}'
data = json.parse(json_string)

print("Name: " + data.name)
print("Age: " + str(data.age))
print("City: " + data.city)

// Convert object to JSON
user = {
    "id": 123,
    "username": "johndoe",
    "profile": {
        "email": "john@example.com",
        "preferences": ["coding", "music", "travel"]
    }
}

json_output = json.stringify(user)
print(json_output)
```

## Practical Application Examples

### REST API Client

```uddin
import "http"
import "json"

fun APIClient(base_url, api_key):
    return {
        "base_url": base_url,
        "headers": {
            "Authorization": "Bearer " + api_key,
            "Content-Type": "application/json"
        },

        "get": fun(endpoint):
            url = this.base_url + endpoint
            return http.request("GET", url, null, this.headers)
        end,

        "post": fun(endpoint, data):
            url = this.base_url + endpoint
            return http.request("POST", url, data, this.headers)
        end,

        "put": fun(endpoint, data):
            url = this.base_url + endpoint
            return http.request("PUT", url, data, this.headers)
        end,

        "delete": fun(endpoint):
            url = this.base_url + endpoint
            return http.request("DELETE", url, null, this.headers)
        end
    }
end

// Usage
api = APIClient("https://api.example.com", "your-api-key")

// GET request
users = api.get("/users")
if (users.status == 200) then:
    user_list = json.parse(users.body)
    print("Found " + str(len(user_list)) + " users")
end

// POST request
new_user = {"name": "Alice", "email": "alice@example.com"}
response = api.post("/users", new_user)
if (response.status == 201) then:
    print("User created successfully")
end
```

### Web Scraper

```uddin
import "http"
import "regex"

fun scrapeWebsite(url):
    try:
        response = http.get(url)

        if (response.status != 200) then:
            return {"error": "HTTP " + str(response.status)}
        end

        html = response.body

        // Extract title using split method
        if (contains(html, "<title>") and contains(html, "</title>")) then:
            parts1 = split(html, "<title>")
            if (len(parts1) > 1) then:
                parts2 = split(parts1[1], "</title>")
                title = parts2[0]
            else:
                title = "No title found"
            end
        else:
            title = "No title found"
        end

        // Extract all links
        link_matches = regex.find_all(html, 'href="([^"]+)"')
        links = []
        if (len(link_matches) > 0) then:
            for (match in link_matches):
                links = append(links, match[1])
            end
        end

        return {
            "url": url,
            "title": title,
            "links": links,
            "content_length": len(html)
        }

    catch (error):
        return {"error": str(error)}
    end
end

// Usage
result = scrapeWebsite("https://example.com")
if ("error" in result) then:
    print("Error: " + result.error)
else:
    print("Title: " + str(result.title))
    print("Found links in the page")
    print("Content length: " + str(result.content_length) + " bytes")
end
```

### Chat Server

```uddin
// Simple chat server using TCP
clients = []
server_running = true

fun broadcastMessage(message, sender_socket):
    for (client in clients):
        if (client.socket != sender_socket) then:
            tcp_write(client.socket, message)
        end
    end
end

fun handleClient(client_socket, client_address):
    // Add client to list
    client_info = {"socket": client_socket, "address": client_address}
    clients = append(clients, client_info)

    print("Client connected: " + client_address)
    broadcastMessage("User joined the chat", client_socket)

    while (server_running):
        try:
            message = tcp_read(client_socket, 1024)
            if (len(message) > 0) then:
                formatted_message = client_address + ": " + message
                print(formatted_message)
                broadcastMessage(formatted_message, client_socket)
            end
        catch (error):
            print("Client disconnected: " + client_address)
            break
        end
    end

    tcp_close(client_socket)
end

// Start chat server
server = tcp_listen(8888)
if (server.success) then:
    print("Chat server started on port 8888")

    while (server_running):
        client = tcp_accept(server.socket)
        if (client.success) then:
            // Handle client in background (simplified)
            handleClient(client.socket, client.address)
        end
    end
end
```

## Error Handling and Best Practices

### Robust HTTP Client

```uddin
import "http"
import "json"

fun safeHttpRequest(method, url, data, headers, retries):
    for (attempt in range(1, retries + 1)):
        try:
            response = http.request(method, url, data, headers)

            // Check for successful status codes
            if (response.status >= 200 and response.status < 300) then:
                return {"success": true, "data": response}
            else if (response.status >= 500 and attempt < retries) then:
                // Retry on server errors
                print("Server error, retrying... (" + str(attempt) + "/" + str(retries) + ")")
                continue
            else:
                return {"success": false, "error": "HTTP " + str(response.status)}
            end

        catch (error):
            if (attempt < retries) then:
                print("Request failed, retrying... (" + str(attempt) + "/" + str(retries) + ")")
            else:
                return {"success": false, "error": str(error)}
            end
        end
    end

    return {"success": false, "error": "Max retries exceeded"}
end

// Usage
result = safeHttpRequest("GET", "https://rickandmortyapi.com/api", null, {}, 3)
if (result.success) then:
    print("Request successful")
    data = result.data.body
    print(json.stringify(data))
else:
    print("Request failed: " + result.error)
end
```

## Tips and Best Practices

1. **Error Handling**: Always use try-catch for network operations
2. **Timeouts**: Implement timeouts to prevent hanging
3. **Rate Limiting**: Respect API rate limits with delays
4. **Connection Pooling**: Reuse connections for better performance
5. **Security**: Validate input and use HTTPS for sensitive data
6. **Logging**: Log all network activity for debugging

With these comprehensive networking features, Uddin-Lang can be used to build various network applications, from simple web scrapers to complex distributed systems.
