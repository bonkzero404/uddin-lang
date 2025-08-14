package main

import (
	"fmt"
	"sync"
	"time"

	uddin "github.com/bonkzero404/uddin-lang"
)

// Task represents a concurrent task
type Task struct {
	ID   int
	Code string
	Vars map[string]interface{}
}

// Result represents the result of a task execution
type Result struct {
	TaskID   int
	Success  bool
	Output   string
	Stats    *uddin.Stats
	Error    error
	Duration time.Duration
}

// Worker represents a worker that executes UDDIN-LANG tasks
type Worker struct {
	ID     int
	engine *uddin.Engine
}

// NewWorker creates a new worker
func NewWorker(id int) *Worker {
	return &Worker{
		ID:     id,
		engine: uddin.New(),
	}
}

// ExecuteTask executes a task and returns the result
func (w *Worker) ExecuteTask(task Task) Result {
	start := time.Now()
	result := Result{
		TaskID: task.ID,
	}

	// Set variables if provided
	if task.Vars != nil {
		w.engine.SetVariables(task.Vars)
	}

	// Execute the code
	stats, err := w.engine.ExecuteString(task.Code)
	result.Duration = time.Since(start)

	if err != nil {
		result.Success = false
		result.Error = err
	} else {
		result.Success = true
		result.Stats = stats
	}

	return result
}

// WorkerPool manages a pool of workers
type WorkerPool struct {
	workers   []*Worker
	taskChan  chan Task
	resultChan chan Result
	wg        sync.WaitGroup
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(numWorkers int) *WorkerPool {
	pool := &WorkerPool{
		workers:    make([]*Worker, numWorkers),
		taskChan:   make(chan Task, numWorkers*2),
		resultChan: make(chan Result, numWorkers*2),
	}

	// Create workers
	for i := 0; i < numWorkers; i++ {
		pool.workers[i] = NewWorker(i + 1)
	}

	return pool
}

// Start starts the worker pool
func (wp *WorkerPool) Start() {
	for _, worker := range wp.workers {
		wp.wg.Add(1)
		go func(w *Worker) {
			defer wp.wg.Done()
			for task := range wp.taskChan {
				fmt.Printf("Worker %d executing task %d\n", w.ID, task.ID)
				result := w.ExecuteTask(task)
				wp.resultChan <- result
			}
		}(worker)
	}
}

// Stop stops the worker pool
func (wp *WorkerPool) Stop() {
	close(wp.taskChan)
	wp.wg.Wait()
	close(wp.resultChan)
}

// AddTask adds a task to the pool
func (wp *WorkerPool) AddTask(task Task) {
	wp.taskChan <- task
}

// GetResult gets a result from the pool
func (wp *WorkerPool) GetResult() Result {
	return <-wp.resultChan
}

// concurrentExecutionExample demonstrates concurrent execution
func concurrentExecutionExample() {
	fmt.Println("=== Concurrent Execution Example ===")

	// Create a worker pool with 3 workers
	pool := NewWorkerPool(3)
	pool.Start()

	// Create tasks
	tasks := []Task{
		{
			ID:   1,
			Code: "x = 10\ny = 20\nprint(\"Task 1 - Sum:\", x + y)",
		},
		{
			ID:   2,
			Code: "for (i in range(1, 6)): print(\"Task 2 - Count:\", i) end",
		},
		{
			ID: 3,
			Code: `fun factorial(n):
				if (n <= 1) then:
					return 1
				else:
					return n * factorial(n - 1)
				end
			end
			print("Task 3 - Factorial of 5:", factorial(5))`,
		},
		{
			ID:   4,
			Code: "arr = [1, 2, 3, 4, 5]\nsum = 0\nfor (item in arr): sum = sum + item end\nprint(\"Task 4 - Array sum:\", sum)",
		},
		{
			ID: 5,
			Code: "name = vars[\"name\"]\nage = vars[\"age\"]\nprint(\"Task 5 - Hello\", name, \"age\", age)",
			Vars: map[string]interface{}{
				"name": "Alice",
				"age":  30,
			},
		},
	}

	// Submit tasks
	for _, task := range tasks {
		pool.AddTask(task)
	}

	// Collect results
	results := make([]Result, len(tasks))
	for i := 0; i < len(tasks); i++ {
		result := pool.GetResult()
		results[result.TaskID-1] = result
	}

	// Stop the pool
	pool.Stop()

	// Display results
	fmt.Println("\n=== Execution Results ===")
	for _, result := range results {
		fmt.Printf("\nTask %d:\n", result.TaskID)
		fmt.Printf("  Success: %v\n", result.Success)
		fmt.Printf("  Duration: %v\n", result.Duration)
		if result.Success {
			fmt.Printf("  Operations: %d\n", result.Stats.Ops)
			fmt.Printf("  User Calls: %d\n", result.Stats.UserCalls)
			fmt.Printf("  Builtin Calls: %d\n", result.Stats.BuiltinCalls)
		} else {
			fmt.Printf("  Error: %v\n", result.Error)
		}
	}
}

// raceConditionExample demonstrates potential race conditions and how to avoid them
func raceConditionExample() {
	fmt.Println("\n=== Race Condition Example ===")

	// Shared state that could cause race conditions
	sharedCounter := 0
	var mutex sync.Mutex

	// Number of goroutines
	numGoroutines := 10
	var wg sync.WaitGroup

	fmt.Printf("Starting %d goroutines with shared counter...\n", numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine gets its own engine instance
			engine := uddin.New()

			// Set variables including the goroutine ID
			engine.SetVariables(map[string]interface{}{
				"goroutine_id": id,
				"counter":      sharedCounter,
			})

			// Execute code that simulates work
			code := `
				print("Goroutine", goroutine_id, "starting with counter:", counter)
				for (i in range(1, 4)):
					print("Goroutine", goroutine_id, "iteration", i)
				end
				print("Goroutine", goroutine_id, "completed")
			`

			_, err := engine.ExecuteString(code)
			if err != nil {
				fmt.Printf("Error in goroutine %d: %v\n", id, err)
			}

			// Safely update shared counter
			mutex.Lock()
			sharedCounter++
			mutex.Unlock()
		}(i + 1)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	fmt.Printf("\nAll goroutines completed. Final counter value: %d\n", sharedCounter)
}

// benchmarkExample demonstrates performance benchmarking
func benchmarkExample() {
	fmt.Println("\n=== Benchmark Example ===")

	// Test different scenarios
	scenarios := []struct {
		name string
		code string
	}{
		{
			name: "Simple arithmetic",
			code: "result = 2 + 3 * 4",
		},
		{
			name: "Loop execution",
			code: "sum = 0\nfor (i in range(1, 101)): sum = sum + i end",
		},
		{
			name: "Function calls",
			code: "fun square(x): return x * x end\nresult = square(10)",
		},
		{
			name: "Array operations",
			code: "arr = [1, 2, 3, 4, 5]\nsum = 0\nfor (item in arr): sum = sum + item end",
		},
	}

	for _, scenario := range scenarios {
		fmt.Printf("\nBenchmarking: %s\n", scenario.name)

		// Run multiple iterations
		iterations := 1000
		totalDuration := time.Duration(0)
		totalOps := 0

		for i := 0; i < iterations; i++ {
			engine := uddin.New()
			start := time.Now()
			stats, err := engine.ExecuteString(scenario.code)
			duration := time.Since(start)

			if err != nil {
				fmt.Printf("Error in iteration %d: %v\n", i, err)
				continue
			}

			totalDuration += duration
			totalOps += stats.Ops
		}

		avgDuration := totalDuration / time.Duration(iterations)
		avgOps := float64(totalOps) / float64(iterations)

		fmt.Printf("  Iterations: %d\n", iterations)
		fmt.Printf("  Average duration: %v\n", avgDuration)
		fmt.Printf("  Average operations: %.2f\n", avgOps)
		fmt.Printf("  Operations per second: %.2f\n", avgOps/avgDuration.Seconds())
	}
}

// concurrentUsageExample demonstrates all concurrent usage patterns
func main() {
	fmt.Println("=== UDDIN-LANG Concurrent Usage Examples ===")

	// Run concurrent execution example
	concurrentExecutionExample()

	// Run race condition example
	raceConditionExample()

	// Run benchmark example
	benchmarkExample()

	fmt.Println("\n=== All concurrent usage examples completed ===")
}