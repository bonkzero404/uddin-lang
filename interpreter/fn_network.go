package interpreter

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/goccy/go-json"
)

// Global HTTP server registry to manage multiple servers
var httpServers = make(map[string]*http.Server)
var httpRoutes = make(map[string]map[string]functionType) // serverID -> route -> handler

// httpGetFunc implements the http_get() built-in function
// http_get(url) -> object
// Example: response = http_get("https://api.example.com/data")
func httpGetFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "http_get", args, 1); err != nil {
		return Value(err)
	}
	url, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "http_get() requires a string URL"))
	}

	return httpRequest(pos, "GET", url, nil, nil)
}

// httpPostFunc implements the http_post() built-in function
// http_post(url, data) -> object
// Example: response = http_post("https://api.example.com/data", {"name": "John"})
func httpPostFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		panic(typeError(pos, "http_post() requires 2 or 3 args, got %d", len(args)))
	}
	url, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "http_post() requires first argument to be a string URL"))
	}

	var headers map[string]Value
	if len(args) == 3 {
		headersMap, ok := args[2].(map[string]Value)
		if !ok {
			panic(typeError(pos, "http_post() requires third argument to be an object (headers)"))
		}
		headers = headersMap
	}

	return httpRequest(pos, "POST", url, args[1], headers)
}

// httpPutFunc implements the http_put() built-in function
// http_put(url, data) -> object
// Example: response = http_put("https://api.example.com/data/1", {"name": "Jane"})
func httpPutFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		panic(typeError(pos, "http_put() requires 2 or 3 args, got %d", len(args)))
	}
	url, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "http_put() requires first argument to be a string URL"))
	}

	var headers map[string]Value
	if len(args) == 3 {
		headersMap, ok := args[2].(map[string]Value)
		if !ok {
			panic(typeError(pos, "http_put() requires third argument to be an object (headers)"))
		}
		headers = headersMap
	}

	return httpRequest(pos, "PUT", url, args[1], headers)
}

// httpDeleteFunc implements the http_delete() built-in function
// http_delete(url) -> object
// Example: response = http_delete("https://api.example.com/data/1")
func httpDeleteFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "http_delete", args, 1); err != nil {
		return Value(err)
	}
	url, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "http_delete() requires a string URL"))
	}

	return httpRequest(pos, "DELETE", url, nil, nil)
}

// httpRequestFunc implements the http_request() built-in function
// http_request(method, url, data, headers) -> object
// Example: response = http_request("PATCH", "https://api.example.com/data/1", {"status": "active"}, {"Authorization": "Bearer token"})
func httpRequestFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 || len(args) > 4 {
		panic(typeError(pos, "http_request() requires 2 to 4 args, got %d", len(args)))
	}

	method, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "http_request() requires first argument to be a string (method)"))
	}

	url, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "http_request() requires second argument to be a string (URL)"))
	}

	var data Value
	var headers map[string]Value

	if len(args) >= 3 {
		data = args[2]
	}

	if len(args) == 4 {
		headersMap, ok := args[3].(map[string]Value)
		if !ok {
			panic(typeError(pos, "http_request() requires fourth argument to be an object (headers)"))
		}
		headers = headersMap
	}

	return httpRequest(pos, method, url, data, headers)
}

// httpRequest is a helper function that performs the actual HTTP request
func httpRequest(pos Position, method, url string, data Value, headers map[string]Value) Value {
	var body io.Reader

	// Prepare request body if data is provided
	if data != nil {
		switch d := data.(type) {
		case string:
			body = strings.NewReader(d)
		case map[string]Value, []Value:
			// Convert to JSON
			jsonData, err := valueToJSON(data)
			if err != nil {
				panic(valueError(pos, "Failed to convert data to JSON: %v", err))
			}
			body = bytes.NewReader(jsonData)
		default:
			// Convert other types to string
			body = strings.NewReader(fmt.Sprintf("%v", data))
		}
	}

	// Create HTTP request
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(valueError(pos, "Failed to create HTTP request: %v", err))
	}

	// Set default Content-Type for POST/PUT requests with JSON data
	if (method == "POST" || method == "PUT" || method == "PATCH") && data != nil {
		if _, isMap := data.(map[string]Value); isMap {
			req.Header.Set("Content-Type", "application/json")
		} else if _, isArray := data.([]Value); isArray {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	// Set custom headers
	for key, value := range headers {
		if strValue, ok := value.(string); ok {
			req.Header.Set(key, strValue)
		}
	}

	// Perform the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		panic(valueError(pos, "HTTP request failed: %v", err))
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(valueError(pos, "Failed to read response body: %v", err))
	}

	// Parse response body as JSON if possible
	var bodyValue Value
	var jsonData any
	if err := json.Unmarshal(respBody, &jsonData); err == nil {
		bodyValue = jsonToValue(jsonData)
	} else {
		// If not JSON, return as string
		bodyValue = string(respBody)
	}

	// Convert headers to map
	headerMap := make(map[string]Value)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headerMap[key] = values[0]
		}
	}

	// Create response object
	response := map[string]Value{
		"status":      resp.StatusCode,
		"status_text": resp.Status,
		"headers":     headerMap,
		"body":        bodyValue,
		"url":         resp.Request.URL.String(),
	}

	return Value(response)
}

// httpServerStartFunc implements the http_server_start() built-in function
// http_server_start(port, server_id?) -> server_object
// Example: server = http_server_start(8080, "main")
func httpServerStartFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 1 || len(args) > 2 {
		panic(typeError(pos, "http_server_start() requires 1 or 2 args, got %d", len(args)))
	}

	port, ok := args[0].(int)
	if !ok {
		panic(typeError(pos, "http_server_start() requires first argument to be an integer (port)"))
	}

	serverID := "default"
	if len(args) == 2 {
		if id, ok := args[1].(string); ok {
			serverID = id
		} else {
			panic(typeError(pos, "http_server_start() requires second argument to be a string (server_id)"))
		}
	}

	// Check if server already exists
	if _, exists := httpServers[serverID]; exists {
		panic(valueError(pos, "HTTP server with ID '%s' already exists", serverID))
	}

	// Initialize routes for this server
	if httpRoutes[serverID] == nil {
		httpRoutes[serverID] = make(map[string]functionType)
	}

	// Create HTTP server
	address := fmt.Sprintf(":%d", port)
	mux := http.NewServeMux()

	// Default handler that looks up routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Look for exact route match first
		routeKey := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		if handler, exists := httpRoutes[serverID][routeKey]; exists {
			// Create request object for the handler
			reqObj := map[string]Value{
				"method":  r.Method,
				"path":    r.URL.Path,
				"query":   r.URL.RawQuery,
				"headers": convertHeaders(r.Header),
				"body":    readRequestBody(r),
			}

			// Create response object
			resObj := map[string]Value{
				"status":  200,
				"headers": make(map[string]Value),
				"body":    "",
				"_writer": w,
			}

			// Call the handler
			handler.call(interp, pos, []Value{reqObj, resObj})
			return
		}

		// No route found
		w.WriteHeader(404)
		w.Write([]byte("404 Not Found"))
	})

	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	// Store server
	httpServers[serverID] = server

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	// Return server object
	serverObj := map[string]Value{
		"type":      "http_server",
		"server_id": serverID,
		"port":      port,
		"address":   address,
		"_server":   server,
	}
	return Value(serverObj)
}

// httpServerStopFunc implements the http_server_stop() built-in function
// http_server_stop(server_id?) -> null
// Example: http_server_stop("main")
func httpServerStopFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) > 1 {
		panic(typeError(pos, "http_server_stop() requires 0 or 1 args, got %d", len(args)))
	}

	serverID := "default"
	if len(args) == 1 {
		if id, ok := args[0].(string); ok {
			serverID = id
		} else {
			panic(typeError(pos, "http_server_stop() requires argument to be a string (server_id)"))
		}
	}

	// Find and stop server
	if server, exists := httpServers[serverID]; exists {
		server.Close()
		delete(httpServers, serverID)
		delete(httpRoutes, serverID)
	} else {
		panic(valueError(pos, "HTTP server with ID '%s' not found", serverID))
	}

	return Value(nil)
}

// httpServerRouteFunc implements the http_server_route() built-in function
// http_server_route(method, path, handler, server_id?) -> null
// Example: http_server_route("GET", "/api/users", my_handler, "main")
func httpServerRouteFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 3 || len(args) > 4 {
		panic(typeError(pos, "http_server_route() requires 3 or 4 args, got %d", len(args)))
	}

	method, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "http_server_route() requires first argument to be a string (method)"))
	}

	path, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "http_server_route() requires second argument to be a string (path)"))
	}

	handler, ok := args[2].(functionType)
	if !ok {
		panic(typeError(pos, "http_server_route() requires third argument to be a function (handler)"))
	}

	serverID := "default"
	if len(args) == 4 {
		if id, ok := args[3].(string); ok {
			serverID = id
		} else {
			panic(typeError(pos, "http_server_route() requires fourth argument to be a string (server_id)"))
		}
	}

	// Check if server exists
	if _, exists := httpServers[serverID]; !exists {
		panic(valueError(pos, "HTTP server with ID '%s' not found. Start server first.", serverID))
	}

	// Register route
	routeKey := fmt.Sprintf("%s %s", method, path)
	httpRoutes[serverID][routeKey] = handler

	return Value(nil)
}

// httpResponseFunc implements the http_response() built-in function
// http_response(response_obj, status?, headers?, body?) -> response_info_object
// Returns: {"status": status_code, "body": response_body, "sent": true}
// Example: result = http_response(res, 200, {"Content-Type": "application/json"}, '{"message": "Hello"}')
func httpResponseFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 1 || len(args) > 4 {
		panic(typeError(pos, "http_response() requires 1 to 4 args, got %d", len(args)))
	}

	resObj, ok := args[0].(map[string]Value)
	if !ok {
		panic(typeError(pos, "http_response() requires first argument to be a response object"))
	}

	w, ok := resObj["_writer"].(http.ResponseWriter)
	if !ok {
		panic(typeError(pos, "http_response() requires a valid response object"))
	}

	// Set status code
	status := 200
	if len(args) >= 2 {
		if s, ok := args[1].(int); ok {
			status = s
		} else {
			panic(typeError(pos, "http_response() requires second argument to be an integer (status)"))
		}
	}

	// Set headers
	if len(args) >= 3 {
		if headers, ok := args[2].(map[string]Value); ok {
			for key, value := range headers {
				if strValue, ok := value.(string); ok {
					w.Header().Set(key, strValue)
				}
			}
		} else {
			panic(typeError(pos, "http_response() requires third argument to be an object (headers)"))
		}
	}

	// Set body
	body := ""
	if len(args) >= 4 {
		switch b := args[3].(type) {
		case string:
			body = b
		case map[string]Value, []Value:
			// Convert to JSON
			jsonData, err := valueToJSON(args[3])
			if err != nil {
				panic(valueError(pos, "Failed to convert body to JSON: %v", err))
			}
			body = string(jsonData)
			if w.Header().Get("Content-Type") == "" {
				w.Header().Set("Content-Type", "application/json")
			}
		default:
			body = fmt.Sprintf("%v", args[3])
		}
	}

	// Write response
	w.WriteHeader(status)
	w.Write([]byte(body))

	// Return response info object
	responseInfo := map[string]Value{
		"status": status,
		"body":   body,
		"sent":   true,
	}
	return Value(responseInfo)
}

// Helper functions for HTTP server

func convertHeaders(headers http.Header) Value {
	headerMap := make(map[string]Value)
	for key, values := range headers {
		if len(values) > 0 {
			headerMap[key] = values[0]
		}
	}
	return Value(headerMap)
}

func readRequestBody(r *http.Request) Value {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return Value("")
	}
	r.Body.Close()
	return Value(string(body))
}

// tcpConnectFunc implements the tcp_connect() built-in function
// tcp_connect(host, port) -> connection_object
// Example: conn = tcp_connect("localhost", 8080)
func tcpConnectFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "tcp_connect", args, 2); err != nil {
		return Value(err)
	}
	host, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "tcp_connect() requires first argument to be a string (host)"))
	}
	port, ok := args[1].(int)
	if !ok {
		panic(typeError(pos, "tcp_connect() requires second argument to be an integer (port)"))
	}

	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.Dial("tcp", address)
	if err != nil {
		// Return connection object with success=false and error message
		connObj := map[string]Value{
			"success": false,
			"error":   err.Error(),
			"socket":  Value(nil),
		}
		return Value(connObj)
	}

	// Create socket object that contains the connection
	socketObj := map[string]Value{
		"_conn": conn,
	}

	// Return connection object with success=true
	connObj := map[string]Value{
		"success": true,
		"socket":  socketObj, // Socket object for operations
		"address": address,
		"_conn":   conn, // Internal connection object for backward compatibility
	}
	return Value(connObj)
}

// tcpListenFunc implements the tcp_listen() built-in function
// tcp_listen(port) -> listener_object
// Example: listener = tcp_listen(8080)
func tcpListenFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "tcp_listen", args, 1); err != nil {
		return Value(err)
	}
	port, ok := args[0].(int)
	if !ok {
		panic(typeError(pos, "tcp_listen() requires an integer (port)"))
	}

	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		// Return listener object with success=false and error message
		listenerObj := map[string]Value{
			"success": false,
			"error":   err.Error(),
			"socket":  Value(nil),
		}
		return Value(listenerObj)
	}

	// Create socket object that contains the listener
	socketObj := map[string]Value{
		"_listener": listener,
	}

	// Return listener object with success=true
	listenerObj := map[string]Value{
		"success":   true,
		"socket":    socketObj, // Socket object for operations
		"address":   address,
		"_listener": listener, // Internal listener object for backward compatibility
	}
	return Value(listenerObj)
}

// tcpAcceptFunc implements the tcp_accept() built-in function
// tcp_accept(listener) -> connection_object
// Example: conn = tcp_accept(listener)
func tcpAcceptFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "tcp_accept", args, 1); err != nil {
		return Value(err)
	}
	listenerObj, ok := args[0].(map[string]Value)
	if !ok {
		panic(typeError(pos, "tcp_accept() requires a listener object"))
	}

	// Try to get listener from either "socket" or "_listener" property
	var listener net.Listener
	if socketObj, ok := listenerObj["socket"].(map[string]Value); ok {
		if l, ok := socketObj["_listener"].(net.Listener); ok {
			listener = l
		}
	} else if l, ok := listenerObj["_listener"].(net.Listener); ok {
		listener = l
	}
	if listener == nil {
		panic(typeError(pos, "tcp_accept() requires a valid TCP listener"))
	}

	conn, err := listener.Accept()
	if err != nil {
		// Return connection object with success=false and error message
		connObj := map[string]Value{
			"success": false,
			"error":   err.Error(),
			"socket":  Value(nil),
		}
		return Value(connObj)
	}

	// Create socket object that contains the connection
	socketObj := map[string]Value{
		"_conn": conn,
	}

	// Return connection object with success=true
	connObj := map[string]Value{
		"success": true,
		"socket":  socketObj, // Socket object for operations
		"address": conn.RemoteAddr().String(),
		"_conn":   conn, // Internal connection object for backward compatibility
	}
	return Value(connObj)
}

// tcpReadFunc implements the tcp_read() built-in function
// tcp_read(connection, size) -> string
// Example: data = tcp_read(conn, 1024)
func tcpReadFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "tcp_read", args, 2); err != nil {
		return Value(err)
	}
	connObj, ok := args[0].(map[string]Value)
	if !ok {
		panic(typeError(pos, "tcp_read() requires a connection object"))
	}

	// Try to get connection from either "socket" or "_conn" property
	var conn net.Conn
	if socketObj, isSocketObj := connObj["socket"].(map[string]Value); isSocketObj {
		if c, isConn := socketObj["_conn"].(net.Conn); isConn {
			conn = c
		}
	} else if c, isConn := connObj["_conn"].(net.Conn); isConn {
		conn = c
	}
	if conn == nil {
		panic(typeError(pos, "tcp_read() requires a valid TCP connection"))
	}

	size, ok := args[1].(int)
	if !ok {
		panic(typeError(pos, "tcp_read() requires second argument to be an integer (size)"))
	}

	buffer := make([]byte, size)
	n, err := conn.Read(buffer)
	if err != nil {
		panic(valueError(pos, "Failed to read from connection: %v", err))
	}

	return Value(string(buffer[:n]))
}

// tcpWriteFunc implements the tcp_write() built-in function
// tcp_write(connection, data) -> bytes_written
// Example: written = tcp_write(conn, "Hello, World!")
func tcpWriteFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "tcp_write", args, 2); err != nil {
		return Value(err)
	}
	connObj, ok := args[0].(map[string]Value)
	if !ok {
		panic(typeError(pos, "tcp_write() requires a connection object"))
	}

	// Try to get connection from either "socket" or "_conn" property
	var conn net.Conn
	if socketObj, isSocketObj := connObj["socket"].(map[string]Value); isSocketObj {
		if c, isConn := socketObj["_conn"].(net.Conn); isConn {
			conn = c
		}
	} else if c, isConn := connObj["_conn"].(net.Conn); isConn {
		conn = c
	}
	if conn == nil {
		panic(typeError(pos, "tcp_write() requires a valid TCP connection"))
	}

	data, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "tcp_write() requires second argument to be a string (data)"))
	}

	n, err := conn.Write([]byte(data))
	if err != nil {
		panic(valueError(pos, "Failed to write to connection: %v", err))
	}

	return Value(n)
}

// tcpCloseFunc implements the tcp_close() built-in function
// tcp_close(connection_or_listener) -> null
// Example: tcp_close(conn)
func tcpCloseFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "tcp_close", args, 1); err != nil {
		return Value(err)
	}
	obj, ok := args[0].(map[string]Value)
	if !ok {
		panic(typeError(pos, "tcp_close() requires a connection or listener object"))
	}

	// Try to close as connection first - check both "socket" and "_conn"
	if socketObj, ok := obj["socket"].(map[string]Value); ok {
		if conn, ok := socketObj["_conn"].(net.Conn); ok {
			err := conn.Close()
			if err != nil {
				panic(valueError(pos, "Failed to close connection: %v", err))
			}
			return Value(nil)
		}
	}
	if conn, ok := obj["_conn"].(net.Conn); ok {
		err := conn.Close()
		if err != nil {
			panic(valueError(pos, "Failed to close connection: %v", err))
		}
		return Value(nil)
	}

	// Try to close as listener - check both "socket" and "_listener"
	if socketObj, ok := obj["socket"].(map[string]Value); ok {
		if listener, ok := socketObj["_listener"].(net.Listener); ok {
			err := listener.Close()
			if err != nil {
				panic(valueError(pos, "Failed to close listener: %v", err))
			}
			return Value(nil)
		}
	}
	if listener, ok := obj["_listener"].(net.Listener); ok {
		err := listener.Close()
		if err != nil {
			panic(valueError(pos, "Failed to close listener: %v", err))
		}
		return Value(nil)
	}

	panic(typeError(pos, "tcp_close() requires a valid TCP connection or listener"))
}

// UDP Connection Functions

// udpConnectFunc implements the udp_connect() built-in function
// udp_connect(host, port) -> connection_object
// Example: conn = udp_connect("localhost", 8080)
func udpConnectFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "udp_connect", args, 2); err != nil {
		return Value(err)
	}
	host, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "udp_connect() requires first argument to be a string (host)"))
	}
	port, ok := args[1].(int)
	if !ok {
		panic(typeError(pos, "udp_connect() requires second argument to be an integer (port)"))
	}

	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.Dial("udp", address)
	if err != nil {
		// Return connection object with success=false and error message
		connObj := map[string]Value{
			"success": false,
			"error":   err.Error(),
			"socket":  Value(nil),
		}
		return Value(connObj)
	}

	// Create socket object that contains the connection
	socketObj := map[string]Value{
		"_conn": conn,
	}

	// Return connection object with success=true
	connObj := map[string]Value{
		"success": true,
		"socket":  socketObj, // Socket object for operations
		"address": address,
		"_conn":   conn, // Internal connection object for backward compatibility
	}
	return Value(connObj)
}

// udpListenFunc implements the udp_listen() built-in function
// udp_listen(port) -> connection_object
// Example: conn = udp_listen(8080)
func udpListenFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "udp_listen", args, 1); err != nil {
		return Value(err)
	}
	port, ok := args[0].(int)
	if !ok {
		panic(typeError(pos, "udp_listen() requires an integer (port)"))
	}

	address := fmt.Sprintf(":%d", port)
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		panic(valueError(pos, "Failed to resolve UDP address %s: %v", address, err))
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		// Return connection object with success=false and error message
		connObj := map[string]Value{
			"success": false,
			"error":   err.Error(),
			"socket":  Value(nil),
		}
		return Value(connObj)
	}

	// Create socket object that contains the connection
	socketObj := map[string]Value{
		"_conn": conn,
	}

	// Return connection object with success=true
	connObj := map[string]Value{
		"success": true,
		"socket":  socketObj, // Socket object for operations
		"address": address,
		"_conn":   conn, // Internal connection object for backward compatibility
	}
	return Value(connObj)
}

// udpReadFunc implements the udp_read() built-in function
// udp_read(connection, size) -> {"data": string, "address": string}
// Example: result = udp_read(conn, 1024)
func udpReadFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "udp_read", args, 2); err != nil {
		return Value(err)
	}
	connObj, ok := args[0].(map[string]Value)
	if !ok {
		panic(typeError(pos, "udp_read() requires a connection object"))
	}

	size, ok := args[1].(int)
	if !ok {
		panic(typeError(pos, "udp_read() requires second argument to be an integer (size)"))
	}

	buffer := make([]byte, size)
	var n int
	var addr net.Addr
	var err error

	// Try to get connection from either "socket" or "_conn" property
	var conn net.Conn
	if socketObj, ok := connObj["socket"].(map[string]Value); ok {
		if c, ok := socketObj["_conn"].(net.Conn); ok {
			conn = c
		}
	} else if c, ok := connObj["_conn"].(net.Conn); ok {
		conn = c
	}
	if conn == nil {
		panic(typeError(pos, "udp_read() requires a valid UDP connection"))
	}

	// Handle both UDP connection types
	if udpConn, ok := conn.(*net.UDPConn); ok {
		n, addr, err = udpConn.ReadFromUDP(buffer)
	} else {
		n, err = conn.Read(buffer)
		addr = conn.RemoteAddr()
	}

	if err != nil {
		panic(valueError(pos, "Failed to read from UDP connection: %v", err))
	}

	result := map[string]Value{
		"data":    string(buffer[:n]),
		"address": addr.String(),
	}
	return Value(result)
}

// udpWriteFunc implements the udp_write() built-in function
// udp_write(connection, data, address?) -> bytes_written
// Example: written = udp_write(conn, "Hello, World!", "192.168.1.100:8080")
func udpWriteFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		panic(typeError(pos, "udp_write() requires 2 or 3 args, got %d", len(args)))
	}

	connObj, ok := args[0].(map[string]Value)
	if !ok {
		panic(typeError(pos, "udp_write() requires a connection object"))
	}

	data, ok := args[1].(string)
	if !ok {
		panic(typeError(pos, "udp_write() requires second argument to be a string (data)"))
	}

	var n int
	var err error

	// Try to get connection from either "socket" or "_conn" property
	var conn net.Conn
	if socketObj, ok := connObj["socket"].(map[string]Value); ok {
		if c, ok := socketObj["_conn"].(net.Conn); ok {
			conn = c
		}
	} else if c, ok := connObj["_conn"].(net.Conn); ok {
		conn = c
	}
	if conn == nil {
		panic(typeError(pos, "udp_write() requires a valid UDP connection"))
	}

	if len(args) == 3 {
		// Write to specific address
		addrStr, ok := args[2].(string)
		if !ok {
			panic(typeError(pos, "udp_write() requires third argument to be a string (address)"))
		}

		udpConn, ok := conn.(*net.UDPConn)
		if !ok {
			panic(typeError(pos, "udp_write() with address requires a UDP listener connection"))
		}

		var addr *net.UDPAddr
		addr, err = net.ResolveUDPAddr("udp", addrStr)
		if err != nil {
			panic(valueError(pos, "Failed to resolve UDP address %s: %v", addrStr, err))
		}

		n, err = udpConn.WriteToUDP([]byte(data), addr)
		if err != nil {
			panic(valueError(pos, "Failed to write to UDP address %s: %v", addrStr, err))
		}
	} else {
		// Write to connected address
		n, err = conn.Write([]byte(data))
	}

	if err != nil {
		panic(valueError(pos, "Failed to write to UDP connection: %v", err))
	}

	return Value(n)
}

// udpCloseFunc implements the udp_close() built-in function
// udp_close(connection) -> null
// Example: udp_close(conn)
func udpCloseFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "udp_close", args, 1); err != nil {
		return Value(err)
	}
	connObj, ok := args[0].(map[string]Value)
	if !ok {
		panic(typeError(pos, "udp_close() requires a connection object"))
	}

	// Try to get connection from either "socket" or "_conn" property
	var conn net.Conn
	if socketObj, ok := connObj["socket"].(map[string]Value); ok {
		if c, ok := socketObj["_conn"].(net.Conn); ok {
			conn = c
		}
	} else if c, ok := connObj["_conn"].(net.Conn); ok {
		conn = c
	}
	if conn == nil {
		panic(typeError(pos, "udp_close() requires a valid UDP connection"))
	}

	err := conn.Close()
	if err != nil {
		panic(valueError(pos, "Failed to close UDP connection: %v", err))
	}

	return Value(nil)
}

// Network Utility Functions

// netResolveFunc implements the net_resolve() built-in function
// net_resolve(hostname) -> {"success": bool, "ips": array, "error": string?}
// Example: result = net_resolve("google.com")
func netResolveFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "net_resolve", args, 1); err != nil {
		return Value(err)
	}
	hostname, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "net_resolve() requires a string (hostname)"))
	}

	ips, err := net.LookupIP(hostname)
	if err != nil {
		// Return error object with success: false
		result := map[string]Value{
			"success": false,
			"error":   err.Error(),
			"ips":     &[]Value{}, // Empty array
		}
		return Value(result)
	}

	// Convert IPs to string array
	ipStrings := make([]Value, len(ips))
	for i, ip := range ips {
		ipStrings[i] = Value(ip.String())
	}

	// Return success object
	result := map[string]Value{
		"success": true,
		"ips":     &ipStrings,
	}
	return Value(result)
}

// netPingFunc implements the net_ping() built-in function
// net_ping(host, port, timeout_ms?) -> {"success": bool, "time_ms": int, "error": string?}
// Example: result = net_ping("google.com", 80, 5000)
func netPingFunc(interp *interpreter, pos Position, args []Value) Value {
	if len(args) < 2 || len(args) > 3 {
		panic(typeError(pos, "net_ping() requires 2 or 3 args, got %d", len(args)))
	}

	host, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "net_ping() requires first argument to be a string (host)"))
	}

	port, ok := args[1].(int)
	if !ok {
		panic(typeError(pos, "net_ping() requires second argument to be an integer (port)"))
	}

	timeoutMs := 5000 // Default 5 seconds
	if len(args) == 3 {
		if timeout, ok := args[2].(int); ok {
			timeoutMs = timeout
		} else {
			panic(typeError(pos, "net_ping() requires third argument to be an integer (timeout_ms)"))
		}
	}

	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	timeout := time.Duration(timeoutMs) * time.Millisecond

	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, timeout)
	elapsed := time.Since(start)

	result := map[string]Value{
		"time_ms": int(elapsed.Nanoseconds() / 1000000),
	}

	if err != nil {
		result["success"] = false
		result["error"] = err.Error()
	} else {
		result["success"] = true
		conn.Close()
	}

	return Value(result)
}
