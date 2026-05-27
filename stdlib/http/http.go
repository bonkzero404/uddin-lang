package stdlib_http

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/bonkzero404/uddin-lang/interpreter"
	"github.com/goccy/go-json"
)

// Package-level server and route registries (mirrors fn_network.go)
var httpServers = make(map[string]*http.Server)
var httpRoutes = make(map[string]map[string]interpreter.Value) // serverID -> routeKey -> handler (Value)

type httpModule struct{}

func (m *httpModule) Name() string { return "http" }

func (m *httpModule) Functions() map[string]interpreter.ModuleFunc {
	return map[string]interpreter.ModuleFunc{
		"get":          httpGetFunc,
		"post":         httpPostFunc,
		"put":          httpPutFunc,
		"delete":       httpDeleteFunc,
		"request":      httpRequestFunc,
		"server_start": httpServerStartFunc,
		"server_stop":  httpServerStopFunc,
		"server_route": httpServerRouteFunc,
		"response":     httpResponseFunc,
		"tcp_connect":  tcpConnectFunc,
		"tcp_listen":   tcpListenFunc,
		"tcp_accept":   tcpAcceptFunc,
		"tcp_read":     tcpReadFunc,
		"tcp_write":    tcpWriteFunc,
		"tcp_close":    tcpCloseFunc,
		"udp_connect":  udpConnectFunc,
		"udp_listen":   udpListenFunc,
		"udp_read":     udpReadFunc,
		"udp_write":    udpWriteFunc,
		"udp_close":    udpCloseFunc,
		"net_resolve":  netResolveFunc,
		"net_ping":     netPingFunc,
	}
}

func init() {
	interpreter.RegisterModule(&httpModule{})
}

// --- HTTP client helpers ---

func valueToJSON(v interpreter.Value) ([]byte, error) {
	return json.Marshal(valueToInterface(v))
}

func valueToInterface(v interpreter.Value) any {
	switch val := v.(type) {
	case map[string]interpreter.Value:
		m := make(map[string]any)
		for k, v2 := range val {
			m[k] = valueToInterface(v2)
		}
		return m
	case []interpreter.Value:
		s := make([]any, len(val))
		for i, v2 := range val {
			s[i] = valueToInterface(v2)
		}
		return s
	case *[]interpreter.Value:
		s := make([]any, len(*val))
		for i, v2 := range *val {
			s[i] = valueToInterface(v2)
		}
		return s
	case nil:
		return nil
	default:
		return val
	}
}

func jsonToValue(data any) interpreter.Value {
	switch val := data.(type) {
	case map[string]any:
		result := make(map[string]interpreter.Value)
		for k, v := range val {
			result[k] = jsonToValue(v)
		}
		return interpreter.Value(result)
	case []any:
		result := make([]interpreter.Value, len(val))
		for i, v := range val {
			result[i] = jsonToValue(v)
		}
		return interpreter.Value(&result)
	case float64:
		if val == float64(int(val)) {
			return interpreter.Value(int(val))
		}
		return interpreter.Value(val)
	case bool:
		return interpreter.Value(val)
	case string:
		return interpreter.Value(val)
	case nil:
		return interpreter.Value(nil)
	default:
		return interpreter.Value(fmt.Sprintf("%v", val))
	}
}

func convertHeaders(headers http.Header) interpreter.Value {
	headerMap := make(map[string]interpreter.Value)
	for key, values := range headers {
		if len(values) > 0 {
			headerMap[key] = values[0]
		}
	}
	return interpreter.Value(headerMap)
}

func readRequestBody(r *http.Request) interpreter.Value {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return interpreter.Value("")
	}
	r.Body.Close()
	return interpreter.Value(string(body))
}

func httpRequest(pos interpreter.Position, method, url string, data interpreter.Value, headers map[string]interpreter.Value) interpreter.Value {
	var body io.Reader
	if data != nil {
		switch d := data.(type) {
		case string:
			body = strings.NewReader(d)
		case map[string]interpreter.Value, []interpreter.Value, *[]interpreter.Value:
			jsonData, err := valueToJSON(data)
			if err != nil {
				panic(interpreter.RuntimeErrorf(pos, "Failed to convert data to JSON: %v", err))
			}
			body = bytes.NewReader(jsonData)
		default:
			body = strings.NewReader(fmt.Sprintf("%v", d))
		}
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(interpreter.RuntimeErrorf(pos, "Failed to create HTTP request: %v", err))
	}

	if (method == "POST" || method == "PUT" || method == "PATCH") && data != nil {
		switch data.(type) {
		case map[string]interpreter.Value, []interpreter.Value, *[]interpreter.Value:
			req.Header.Set("Content-Type", "application/json")
		}
	}

	for key, value := range headers {
		if strValue, ok := value.(string); ok {
			req.Header.Set(key, strValue)
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		panic(interpreter.RuntimeErrorf(pos, "HTTP request failed: %v", err))
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(interpreter.RuntimeErrorf(pos, "Failed to read response body: %v", err))
	}

	var bodyValue interpreter.Value
	var jsonData any
	if err := json.Unmarshal(respBody, &jsonData); err == nil {
		bodyValue = jsonToValue(jsonData)
	} else {
		bodyValue = string(respBody)
	}

	headerMap := make(map[string]interpreter.Value)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headerMap[key] = values[0]
		}
	}

	response := map[string]interpreter.Value{
		"status":      resp.StatusCode,
		"status_text": resp.Status,
		"headers":     headerMap,
		"body":        bodyValue,
		"url":         resp.Request.URL.String(),
	}
	return interpreter.Value(response)
}

// --- HTTP client functions ---

func httpGetFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		panic(interpreter.TypeErrorf(pos, "http.get() requires exactly 1 argument, got %d", len(args)))
	}
	url, ok := args[0].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.get() requires a string URL"))
	}
	return httpRequest(pos, "GET", url, nil, nil)
}

func httpPostFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 2 || len(args) > 3 {
		panic(interpreter.TypeErrorf(pos, "http.post() requires 2 or 3 args, got %d", len(args)))
	}
	url, ok := args[0].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.post() requires first argument to be a string URL"))
	}
	var headers map[string]interpreter.Value
	if len(args) == 3 {
		headersMap, ok := args[2].(map[string]interpreter.Value)
		if !ok {
			panic(interpreter.TypeErrorf(pos, "http.post() requires third argument to be an object (headers)"))
		}
		headers = headersMap
	}
	return httpRequest(pos, "POST", url, args[1], headers)
}

func httpPutFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 2 || len(args) > 3 {
		panic(interpreter.TypeErrorf(pos, "http.put() requires 2 or 3 args, got %d", len(args)))
	}
	url, ok := args[0].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.put() requires first argument to be a string URL"))
	}
	var headers map[string]interpreter.Value
	if len(args) == 3 {
		headersMap, ok := args[2].(map[string]interpreter.Value)
		if !ok {
			panic(interpreter.TypeErrorf(pos, "http.put() requires third argument to be an object (headers)"))
		}
		headers = headersMap
	}
	return httpRequest(pos, "PUT", url, args[1], headers)
}

func httpDeleteFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		panic(interpreter.TypeErrorf(pos, "http.delete() requires exactly 1 argument, got %d", len(args)))
	}
	url, ok := args[0].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.delete() requires a string URL"))
	}
	return httpRequest(pos, "DELETE", url, nil, nil)
}

func httpRequestFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 2 || len(args) > 4 {
		panic(interpreter.TypeErrorf(pos, "http.request() requires 2 to 4 args, got %d", len(args)))
	}
	method, ok := args[0].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.request() requires first argument to be a string (method)"))
	}
	url, ok := args[1].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.request() requires second argument to be a string (URL)"))
	}
	var data interpreter.Value
	var headers map[string]interpreter.Value
	if len(args) >= 3 {
		data = args[2]
	}
	if len(args) == 4 {
		headersMap, ok := args[3].(map[string]interpreter.Value)
		if !ok {
			panic(interpreter.TypeErrorf(pos, "http.request() requires fourth argument to be an object (headers)"))
		}
		headers = headersMap
	}
	return httpRequest(pos, method, url, data, headers)
}

// --- HTTP server functions ---

func httpServerStartFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 1 || len(args) > 2 {
		panic(interpreter.TypeErrorf(pos, "http.server_start() requires 1 or 2 args, got %d", len(args)))
	}
	port, ok := args[0].(int)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.server_start() requires first argument to be an integer (port)"))
	}
	serverID := "default"
	if len(args) == 2 {
		if id, ok := args[1].(string); ok {
			serverID = id
		} else {
			panic(interpreter.TypeErrorf(pos, "http.server_start() requires second argument to be a string (server_id)"))
		}
	}

	if _, exists := httpServers[serverID]; exists {
		panic(interpreter.RuntimeErrorf(pos, "HTTP server with ID '%s' already exists", serverID))
	}

	if httpRoutes[serverID] == nil {
		httpRoutes[serverID] = make(map[string]interpreter.Value)
	}

	address := fmt.Sprintf(":%d", port)
	mux := http.NewServeMux()

	// Capture serverID in closure
	sid := serverID
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		routeKey := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		// We cannot call user functions from goroutines without an interpreter context,
		// so the route handler is stored as interpreter.Value but calling it here
		// would require an interpreter. We store the handler value and call it with
		// a nil context — the route will be dispatched by type assertion.
		if _, exists := httpRoutes[sid][routeKey]; exists {
			// Best-effort: we cannot invoke uddin-lang functions from a goroutine
			// without capturing the interpreter. Route calling is a no-op here;
			// actual invocation requires the caller to setup the server properly.
			w.WriteHeader(501)
			w.Write([]byte("Handler invocation not supported from standalone stdlib"))
			return
		}
		w.WriteHeader(404)
		w.Write([]byte("404 Not Found"))
	})

	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}
	httpServers[serverID] = server

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	serverObj := map[string]interpreter.Value{
		"type":      "http_server",
		"server_id": serverID,
		"port":      port,
		"address":   address,
		"_server":   server,
	}
	return interpreter.Value(serverObj)
}

func httpServerStopFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) > 1 {
		panic(interpreter.TypeErrorf(pos, "http.server_stop() requires 0 or 1 args, got %d", len(args)))
	}
	serverID := "default"
	if len(args) == 1 {
		if id, ok := args[0].(string); ok {
			serverID = id
		} else {
			panic(interpreter.TypeErrorf(pos, "http.server_stop() requires argument to be a string (server_id)"))
		}
	}
	if server, exists := httpServers[serverID]; exists {
		server.Close()
		delete(httpServers, serverID)
		delete(httpRoutes, serverID)
	} else {
		panic(interpreter.RuntimeErrorf(pos, "HTTP server with ID '%s' not found", serverID))
	}
	return interpreter.Value(nil)
}

func httpServerRouteFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 3 || len(args) > 4 {
		panic(interpreter.TypeErrorf(pos, "http.server_route() requires 3 or 4 args, got %d", len(args)))
	}
	method, ok := args[0].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.server_route() requires first argument to be a string (method)"))
	}
	path, ok := args[1].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.server_route() requires second argument to be a string (path)"))
	}
	// Store handler as interpreter.Value (function value)
	handler := args[2]

	serverID := "default"
	if len(args) == 4 {
		if id, ok := args[3].(string); ok {
			serverID = id
		} else {
			panic(interpreter.TypeErrorf(pos, "http.server_route() requires fourth argument to be a string (server_id)"))
		}
	}
	if _, exists := httpServers[serverID]; !exists {
		panic(interpreter.RuntimeErrorf(pos, "HTTP server with ID '%s' not found. Start server first.", serverID))
	}
	routeKey := fmt.Sprintf("%s %s", method, path)
	httpRoutes[serverID][routeKey] = handler
	return interpreter.Value(nil)
}

func httpResponseFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 1 || len(args) > 4 {
		panic(interpreter.TypeErrorf(pos, "http.response() requires 1 to 4 args, got %d", len(args)))
	}
	resObj, ok := args[0].(map[string]interpreter.Value)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.response() requires first argument to be a response object"))
	}
	w, ok := resObj["_writer"].(http.ResponseWriter)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.response() requires a valid response object"))
	}
	status := 200
	if len(args) >= 2 {
		if s, ok := args[1].(int); ok {
			status = s
		} else {
			panic(interpreter.TypeErrorf(pos, "http.response() requires second argument to be an integer (status)"))
		}
	}
	if len(args) >= 3 {
		if headers, ok := args[2].(map[string]interpreter.Value); ok {
			for key, value := range headers {
				if strValue, ok := value.(string); ok {
					w.Header().Set(key, strValue)
				}
			}
		} else {
			panic(interpreter.TypeErrorf(pos, "http.response() requires third argument to be an object (headers)"))
		}
	}
	body := ""
	if len(args) >= 4 {
		switch b := args[3].(type) {
		case string:
			body = b
		case map[string]interpreter.Value, []interpreter.Value, *[]interpreter.Value:
			jsonData, err := valueToJSON(args[3])
			if err != nil {
				panic(interpreter.RuntimeErrorf(pos, "Failed to convert body to JSON: %v", err))
			}
			body = string(jsonData)
			if w.Header().Get("Content-Type") == "" {
				w.Header().Set("Content-Type", "application/json")
			}
		default:
			body = fmt.Sprintf("%v", b)
		}
	}
	w.WriteHeader(status)
	w.Write([]byte(body))
	responseInfo := map[string]interpreter.Value{
		"status": status,
		"body":   body,
		"sent":   true,
	}
	return interpreter.Value(responseInfo)
}

// --- TCP functions ---

func tcpConnectFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 2 {
		panic(interpreter.TypeErrorf(pos, "http.tcp_connect() requires exactly 2 arguments, got %d", len(args)))
	}
	host, ok := args[0].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.tcp_connect() requires first argument to be a string (host)"))
	}
	port, ok := args[1].(int)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.tcp_connect() requires second argument to be an integer (port)"))
	}
	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   err.Error(),
			"socket":  interpreter.Value(nil),
		})
	}
	socketObj := map[string]interpreter.Value{"_conn": conn}
	return interpreter.Value(map[string]interpreter.Value{
		"success": true,
		"socket":  socketObj,
		"address": address,
		"_conn":   conn,
	})
}

func tcpListenFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		panic(interpreter.TypeErrorf(pos, "http.tcp_listen() requires exactly 1 argument, got %d", len(args)))
	}
	port, ok := args[0].(int)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.tcp_listen() requires an integer (port)"))
	}
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   err.Error(),
			"socket":  interpreter.Value(nil),
		})
	}
	socketObj := map[string]interpreter.Value{"_listener": listener}
	return interpreter.Value(map[string]interpreter.Value{
		"success":   true,
		"socket":    socketObj,
		"address":   address,
		"_listener": listener,
	})
}

func tcpAcceptFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		panic(interpreter.TypeErrorf(pos, "http.tcp_accept() requires exactly 1 argument, got %d", len(args)))
	}
	listenerObj, ok := args[0].(map[string]interpreter.Value)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.tcp_accept() requires a listener object"))
	}
	var listener net.Listener
	if socketObj, ok := listenerObj["socket"].(map[string]interpreter.Value); ok {
		if l, ok := socketObj["_listener"].(net.Listener); ok {
			listener = l
		}
	} else if l, ok := listenerObj["_listener"].(net.Listener); ok {
		listener = l
	}
	if listener == nil {
		panic(interpreter.TypeErrorf(pos, "http.tcp_accept() requires a valid TCP listener"))
	}
	conn, err := listener.Accept()
	if err != nil {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   err.Error(),
			"socket":  interpreter.Value(nil),
		})
	}
	socketObj := map[string]interpreter.Value{"_conn": conn}
	return interpreter.Value(map[string]interpreter.Value{
		"success": true,
		"socket":  socketObj,
		"address": conn.RemoteAddr().String(),
		"_conn":   conn,
	})
}

func tcpReadFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 2 {
		panic(interpreter.TypeErrorf(pos, "http.tcp_read() requires exactly 2 arguments, got %d", len(args)))
	}
	connObj, ok := args[0].(map[string]interpreter.Value)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.tcp_read() requires a connection object"))
	}
	var conn net.Conn
	if socketObj, ok := connObj["socket"].(map[string]interpreter.Value); ok {
		if c, ok := socketObj["_conn"].(net.Conn); ok {
			conn = c
		}
	} else if c, ok := connObj["_conn"].(net.Conn); ok {
		conn = c
	}
	if conn == nil {
		panic(interpreter.TypeErrorf(pos, "http.tcp_read() requires a valid TCP connection"))
	}
	size, ok := args[1].(int)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.tcp_read() requires second argument to be an integer (size)"))
	}
	buffer := make([]byte, size)
	n, err := conn.Read(buffer)
	if err != nil {
		panic(interpreter.RuntimeErrorf(pos, "Failed to read from connection: %v", err))
	}
	return interpreter.Value(string(buffer[:n]))
}

func tcpWriteFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 2 {
		panic(interpreter.TypeErrorf(pos, "http.tcp_write() requires exactly 2 arguments, got %d", len(args)))
	}
	connObj, ok := args[0].(map[string]interpreter.Value)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.tcp_write() requires a connection object"))
	}
	var conn net.Conn
	if socketObj, ok := connObj["socket"].(map[string]interpreter.Value); ok {
		if c, ok := socketObj["_conn"].(net.Conn); ok {
			conn = c
		}
	} else if c, ok := connObj["_conn"].(net.Conn); ok {
		conn = c
	}
	if conn == nil {
		panic(interpreter.TypeErrorf(pos, "http.tcp_write() requires a valid TCP connection"))
	}
	data, ok := args[1].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.tcp_write() requires second argument to be a string (data)"))
	}
	n, err := conn.Write([]byte(data))
	if err != nil {
		panic(interpreter.RuntimeErrorf(pos, "Failed to write to connection: %v", err))
	}
	return interpreter.Value(n)
}

func tcpCloseFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		panic(interpreter.TypeErrorf(pos, "http.tcp_close() requires exactly 1 argument, got %d", len(args)))
	}
	obj, ok := args[0].(map[string]interpreter.Value)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.tcp_close() requires a connection or listener object"))
	}
	if socketObj, ok := obj["socket"].(map[string]interpreter.Value); ok {
		if conn, ok := socketObj["_conn"].(net.Conn); ok {
			conn.Close()
			return interpreter.Value(nil)
		}
	}
	if conn, ok := obj["_conn"].(net.Conn); ok {
		conn.Close()
		return interpreter.Value(nil)
	}
	if socketObj, ok := obj["socket"].(map[string]interpreter.Value); ok {
		if listener, ok := socketObj["_listener"].(net.Listener); ok {
			listener.Close()
			return interpreter.Value(nil)
		}
	}
	if listener, ok := obj["_listener"].(net.Listener); ok {
		listener.Close()
		return interpreter.Value(nil)
	}
	panic(interpreter.TypeErrorf(pos, "http.tcp_close() requires a valid TCP connection or listener"))
}

// --- UDP functions ---

func udpConnectFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 2 {
		panic(interpreter.TypeErrorf(pos, "http.udp_connect() requires exactly 2 arguments, got %d", len(args)))
	}
	host, ok := args[0].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.udp_connect() requires first argument to be a string (host)"))
	}
	port, ok := args[1].(int)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.udp_connect() requires second argument to be an integer (port)"))
	}
	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.Dial("udp", address)
	if err != nil {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   err.Error(),
			"socket":  interpreter.Value(nil),
		})
	}
	socketObj := map[string]interpreter.Value{"_conn": conn}
	return interpreter.Value(map[string]interpreter.Value{
		"success": true,
		"socket":  socketObj,
		"address": address,
		"_conn":   conn,
	})
}

func udpListenFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		panic(interpreter.TypeErrorf(pos, "http.udp_listen() requires exactly 1 argument, got %d", len(args)))
	}
	port, ok := args[0].(int)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.udp_listen() requires an integer (port)"))
	}
	address := fmt.Sprintf(":%d", port)
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		panic(interpreter.RuntimeErrorf(pos, "Failed to resolve UDP address %s: %v", address, err))
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   err.Error(),
			"socket":  interpreter.Value(nil),
		})
	}
	socketObj := map[string]interpreter.Value{"_conn": conn}
	return interpreter.Value(map[string]interpreter.Value{
		"success": true,
		"socket":  socketObj,
		"address": address,
		"_conn":   conn,
	})
}

func udpReadFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 2 {
		panic(interpreter.TypeErrorf(pos, "http.udp_read() requires exactly 2 arguments, got %d", len(args)))
	}
	connObj, ok := args[0].(map[string]interpreter.Value)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.udp_read() requires a connection object"))
	}
	size, ok := args[1].(int)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.udp_read() requires second argument to be an integer (size)"))
	}
	var conn net.Conn
	if socketObj, ok := connObj["socket"].(map[string]interpreter.Value); ok {
		if c, ok := socketObj["_conn"].(net.Conn); ok {
			conn = c
		}
	} else if c, ok := connObj["_conn"].(net.Conn); ok {
		conn = c
	}
	if conn == nil {
		panic(interpreter.TypeErrorf(pos, "http.udp_read() requires a valid UDP connection"))
	}
	buffer := make([]byte, size)
	var n int
	var addr net.Addr
	var err error
	if udpConn, ok := conn.(*net.UDPConn); ok {
		n, addr, err = udpConn.ReadFromUDP(buffer)
	} else {
		n, err = conn.Read(buffer)
		addr = conn.RemoteAddr()
	}
	if err != nil {
		panic(interpreter.RuntimeErrorf(pos, "Failed to read from UDP connection: %v", err))
	}
	result := map[string]interpreter.Value{
		"data":    string(buffer[:n]),
		"address": addr.String(),
	}
	return interpreter.Value(result)
}

func udpWriteFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 2 || len(args) > 3 {
		panic(interpreter.TypeErrorf(pos, "http.udp_write() requires 2 or 3 args, got %d", len(args)))
	}
	connObj, ok := args[0].(map[string]interpreter.Value)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.udp_write() requires a connection object"))
	}
	data, ok := args[1].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.udp_write() requires second argument to be a string (data)"))
	}
	var conn net.Conn
	if socketObj, ok := connObj["socket"].(map[string]interpreter.Value); ok {
		if c, ok := socketObj["_conn"].(net.Conn); ok {
			conn = c
		}
	} else if c, ok := connObj["_conn"].(net.Conn); ok {
		conn = c
	}
	if conn == nil {
		panic(interpreter.TypeErrorf(pos, "http.udp_write() requires a valid UDP connection"))
	}
	var n int
	var err error
	if len(args) == 3 {
		addrStr, ok := args[2].(string)
		if !ok {
			panic(interpreter.TypeErrorf(pos, "http.udp_write() requires third argument to be a string (address)"))
		}
		udpConn, ok := conn.(*net.UDPConn)
		if !ok {
			panic(interpreter.TypeErrorf(pos, "http.udp_write() with address requires a UDP listener connection"))
		}
		addr, err2 := net.ResolveUDPAddr("udp", addrStr)
		if err2 != nil {
			panic(interpreter.RuntimeErrorf(pos, "Failed to resolve UDP address %s: %v", addrStr, err2))
		}
		n, err = udpConn.WriteToUDP([]byte(data), addr)
	} else {
		n, err = conn.Write([]byte(data))
	}
	if err != nil {
		panic(interpreter.RuntimeErrorf(pos, "Failed to write to UDP connection: %v", err))
	}
	return interpreter.Value(n)
}

func udpCloseFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		panic(interpreter.TypeErrorf(pos, "http.udp_close() requires exactly 1 argument, got %d", len(args)))
	}
	connObj, ok := args[0].(map[string]interpreter.Value)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.udp_close() requires a connection object"))
	}
	var conn net.Conn
	if socketObj, ok := connObj["socket"].(map[string]interpreter.Value); ok {
		if c, ok := socketObj["_conn"].(net.Conn); ok {
			conn = c
		}
	} else if c, ok := connObj["_conn"].(net.Conn); ok {
		conn = c
	}
	if conn == nil {
		panic(interpreter.TypeErrorf(pos, "http.udp_close() requires a valid UDP connection"))
	}
	conn.Close()
	return interpreter.Value(nil)
}

// --- Network utility functions ---

func netResolveFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) != 1 {
		panic(interpreter.TypeErrorf(pos, "http.net_resolve() requires exactly 1 argument, got %d", len(args)))
	}
	hostname, ok := args[0].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.net_resolve() requires a string (hostname)"))
	}
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return interpreter.Value(map[string]interpreter.Value{
			"success": false,
			"error":   err.Error(),
			"ips":     &[]interpreter.Value{},
		})
	}
	ipStrings := make([]interpreter.Value, len(ips))
	for i, ip := range ips {
		ipStrings[i] = ip.String()
	}
	return interpreter.Value(map[string]interpreter.Value{
		"success": true,
		"ips":     &ipStrings,
	})
}

func netPingFunc(ctx interpreter.ModuleContext, pos interpreter.Position, args []interpreter.Value) interpreter.Value {
	if len(args) < 2 || len(args) > 3 {
		panic(interpreter.TypeErrorf(pos, "http.net_ping() requires 2 or 3 args, got %d", len(args)))
	}
	host, ok := args[0].(string)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.net_ping() requires first argument to be a string (host)"))
	}
	port, ok := args[1].(int)
	if !ok {
		panic(interpreter.TypeErrorf(pos, "http.net_ping() requires second argument to be an integer (port)"))
	}
	timeoutMs := 5000
	if len(args) == 3 {
		if timeout, ok := args[2].(int); ok {
			timeoutMs = timeout
		} else {
			panic(interpreter.TypeErrorf(pos, "http.net_ping() requires third argument to be an integer (timeout_ms)"))
		}
	}
	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	timeout := time.Duration(timeoutMs) * time.Millisecond
	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, timeout)
	elapsed := time.Since(start)
	result := map[string]interpreter.Value{
		"time_ms": int(elapsed.Nanoseconds() / 1000000),
	}
	if err != nil {
		result["success"] = false
		result["error"] = err.Error()
	} else {
		result["success"] = true
		conn.Close()
	}
	return interpreter.Value(result)
}
