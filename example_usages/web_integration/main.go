package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	uddin "github.com/bonkzero404/uddin-lang"
)

// ScriptRequest represents a request to execute UDDIN-LANG code
type ScriptRequest struct {
	Code      string         `json:"code"`
	Variables map[string]any `json:"variables,omitempty"`
}

// ScriptResponse represents the response from script execution
type ScriptResponse struct {
	Success bool         `json:"success"`
	Result  any          `json:"result,omitempty"`
	Output  string       `json:"output,omitempty"`
	Stats   *uddin.Stats `json:"stats,omitempty"`
	Error   string       `json:"error,omitempty"`
	AST     string       `json:"ast,omitempty"`
}

// WebServer demonstrates UDDIN-LANG integration with web services
type WebServer struct {
	engine *uddin.Engine
}

// NewWebServer creates a new web server with UDDIN-LANG integration
func NewWebServer() *WebServer {
	return &WebServer{
		engine: uddin.New(),
	}
}

// executeHandler handles script execution requests
func (ws *WebServer) executeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ScriptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := ScriptResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid JSON: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create a new engine for each request to avoid state pollution
	engine := uddin.New()

	// Capture output
	var output strings.Builder
	engine.SetStdout(&output)

	// Set variables if provided
	if req.Variables != nil {
		engine.SetVariables(req.Variables)
	}

	// Execute the code
	stats, err := engine.ExecuteString(req.Code)
	if err != nil {
		response := ScriptResponse{
			Success: false,
			Error:   err.Error(),
			Output:  output.String(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ScriptResponse{
		Success: true,
		Output:  output.String(),
		Stats:   stats,
	}

	json.NewEncoder(w).Encode(response)
}

// evaluateHandler handles expression evaluation requests
func (ws *WebServer) evaluateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ScriptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := ScriptResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid JSON: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create a new engine for each request
	engine := uddin.New()

	// Set variables if provided
	if req.Variables != nil {
		engine.SetVariables(req.Variables)
	}

	// Evaluate the expression
	result, stats, err := engine.EvaluateString(req.Code)
	if err != nil {
		response := ScriptResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ScriptResponse{
		Success: true,
		Result:  result,
		Stats:   stats,
	}

	json.NewEncoder(w).Encode(response)
}

// astHandler handles AST conversion requests
func (ws *WebServer) astHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ScriptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := ScriptResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid JSON: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Convert code to AST JSON
	jsonData, err := uddin.ConvertStringToJSON(req.Code)
	if err != nil {
		response := ScriptResponse{
			Success: false,
			Error:   fmt.Sprintf("AST conversion failed: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ScriptResponse{
		Success: true,
		AST:     string(jsonData),
	}

	json.NewEncoder(w).Encode(response)
}

// healthHandler provides a health check endpoint
func (ws *WebServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "uddin-lang-api",
	})
}

// Start starts the web server
func (ws *WebServer) Start(port string) {
	http.HandleFunc("/execute", ws.executeHandler)
	http.HandleFunc("/evaluate", ws.evaluateHandler)
	http.HandleFunc("/ast", ws.astHandler)
	http.HandleFunc("/health", ws.healthHandler)

	fmt.Printf("UDDIN-LANG Web API Server starting on port %s\n", port)
	fmt.Println("Available endpoints:")
	fmt.Println("  POST /execute   - Execute UDDIN-LANG code")
	fmt.Println("  POST /evaluate  - Evaluate UDDIN-LANG expression")
	fmt.Println("  POST /ast       - Convert code to AST JSON")
	fmt.Println("  GET  /health    - Health check")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// webIntegrationExample demonstrates the web integration
func main() {
	fmt.Println("=== UDDIN-LANG Web Integration Example ===")
	fmt.Println("\nThis example shows how to integrate UDDIN-LANG with web services.")
	fmt.Println("\nTo run the web server, uncomment the following line:")
	fmt.Println("// server := NewWebServer()")
	fmt.Println("// server.Start(\"8080\")")

	// Example of how to use the web integration
	fmt.Println("\nExample API requests:")
	fmt.Println("\n1. Execute code:")
	fmt.Println(`curl -X POST http://localhost:8080/execute \`)
	fmt.Println(`  -H "Content-Type: application/json" \`)
	fmt.Println(`  -d '{"code": "x = 10\\ny = 20\\nprint(\\"Sum:\\", x + y)"}'`)

	fmt.Println("\n2. Evaluate expression:")
	fmt.Println(`curl -X POST http://localhost:8080/evaluate \`)
	fmt.Println(`  -H "Content-Type: application/json" \`)
	fmt.Println(`  -d '{"code": "2 + 3 * 4", "variables": {"x": 10, "y": 5}}'`)

	fmt.Println("\n3. Convert to AST:")
	fmt.Println(`curl -X POST http://localhost:8080/ast \`)
	fmt.Println(`  -H "Content-Type: application/json" \`)
	fmt.Println(`  -d '{"code": "fun add(a, b): return a + b end"}'`)

	fmt.Println("\n4. Health check:")
	fmt.Println(`curl http://localhost:8080/health`)

	// Uncomment to actually start the server
	// server := NewWebServer()
	// server.Start("8080")

	fmt.Println("\n=== Web integration examples completed ===")
}
