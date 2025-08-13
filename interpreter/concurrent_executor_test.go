package interpreter

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewConcurrentExecutor(t *testing.T) {
	executor := NewConcurrentExecutor(4, 20)
	defer executor.Shutdown()
	
	if executor == nil {
		t.Error("Expected executor to be created")
	}
	
	if executor.maxWorkers != 4 {
		t.Errorf("Expected 4 workers, got %d", executor.maxWorkers)
	}
	
	if executor.maxQueueSize != 20 {
		t.Errorf("Expected queue size 20, got %d", executor.maxQueueSize)
	}
}

func TestNewConcurrentExecutor_DefaultValues(t *testing.T) {
	executor := NewConcurrentExecutor(0, 0)
	defer executor.Shutdown()
	
	if executor.maxWorkers <= 0 {
		t.Error("Expected positive number of workers")
	}
	
	if executor.maxQueueSize <= 0 {
		t.Error("Expected positive queue size")
	}
}

func TestNewWorkerPool(t *testing.T) {
	pool := NewWorkerPool(2, 10)
	defer pool.Shutdown()
	
	if pool == nil {
		t.Error("Expected worker pool to be created")
	}
	
	if len(pool.workers) != 2 {
		t.Errorf("Expected 2 workers, got %d", len(pool.workers))
	}
	
	if !pool.active {
		t.Error("Expected worker pool to be active")
	}
}

func TestWorkerPool_SubmitTask(t *testing.T) {
	pool := NewWorkerPool(2, 10)
	defer pool.Shutdown()
	
	task := &Task{
		ID: "test_task",
		Function: func() (Value, error) {
			return "result", nil
		},
		Result:  make(chan TaskResult, 1),
		Context: context.Background(),
	}
	
	err := pool.SubmitTask(task)
	if err != nil {
		t.Errorf("Expected task submission to succeed, got error: %v", err)
	}
	
	// Wait for result
	select {
	case result := <-task.Result:
		if result.Error != nil {
			t.Errorf("Expected task to succeed, got error: %v", result.Error)
		}
		if result.Value != "result" {
			t.Errorf("Expected result 'result', got %v", result.Value)
		}
	case <-time.After(5 * time.Second):
		t.Error("Task execution timed out")
	}
}

func TestWorkerPool_SubmitTask_AfterShutdown(t *testing.T) {
	pool := NewWorkerPool(2, 10)
	pool.Shutdown()
	
	task := &Task{
		ID: "test_task",
		Function: func() (Value, error) {
			return "result", nil
		},
		Result:  make(chan TaskResult, 1),
		Context: context.Background(),
	}
	
	err := pool.SubmitTask(task)
	if err == nil {
		t.Error("Expected task submission to fail after shutdown")
	}
}

func TestWorker_ExecuteTask_WithTimeout(t *testing.T) {
	pool := NewWorkerPool(1, 10)
	defer pool.Shutdown()
	
	task := &Task{
		ID: "timeout_task",
		Function: func() (Value, error) {
			time.Sleep(2 * time.Second)
			return "result", nil
		},
		Result:  make(chan TaskResult, 1),
		Context: context.Background(),
		Timeout: 100 * time.Millisecond,
	}
	
	err := pool.SubmitTask(task)
	if err != nil {
		t.Errorf("Expected task submission to succeed, got error: %v", err)
	}
	
	// Wait for result
	select {
	case result := <-task.Result:
		if result.Error == nil {
			t.Error("Expected task to timeout")
		}
	case <-time.After(5 * time.Second):
		t.Error("Test timed out")
	}
}

func TestWorker_ExecuteTask_WithPanic(t *testing.T) {
	pool := NewWorkerPool(1, 10)
	defer pool.Shutdown()
	
	task := &Task{
		ID: "panic_task",
		Function: func() (Value, error) {
			panic("test panic")
		},
		Result:  make(chan TaskResult, 1),
		Context: context.Background(),
	}
	
	err := pool.SubmitTask(task)
	if err != nil {
		t.Errorf("Expected task submission to succeed, got error: %v", err)
	}
	
	// Wait for result
	select {
	case result := <-task.Result:
		if result.Error == nil {
			t.Error("Expected task to return error due to panic")
		}
	case <-time.After(5 * time.Second):
		t.Error("Test timed out")
	}
}

func TestNewAsyncIOPool(t *testing.T) {
	pool := NewAsyncIOPool(2, 10)
	defer pool.Shutdown()
	
	if pool == nil {
		t.Error("Expected async I/O pool to be created")
	}
	
	if len(pool.workers) != 2 {
		t.Errorf("Expected 2 workers, got %d", len(pool.workers))
	}
	
	if !pool.active {
		t.Error("Expected async I/O pool to be active")
	}
}

func TestAsyncIOPool_SubmitIOTask(t *testing.T) {
	pool := NewAsyncIOPool(2, 10)
	defer pool.Shutdown()
	
	task := &IOTask{
		ID: "test_io_task",
		Function: func() (Value, error) {
			return "io_result", nil
		},
		Result:  make(chan TaskResult, 1),
		Context: context.Background(),
	}
	
	err := pool.SubmitIOTask(task)
	if err != nil {
		t.Errorf("Expected I/O task submission to succeed, got error: %v", err)
	}
	
	// Wait for result
	select {
	case result := <-task.Result:
		if result.Error != nil {
			t.Errorf("Expected I/O task to succeed, got error: %v", result.Error)
		}
		if result.Value != "io_result" {
			t.Errorf("Expected result 'io_result', got %v", result.Value)
		}
	case <-time.After(5 * time.Second):
		t.Error("I/O task execution timed out")
	}
}

func TestAsyncIOPool_SubmitIOTask_WithCancellation(t *testing.T) {
	pool := NewAsyncIOPool(1, 10)
	defer pool.Shutdown()
	
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	task := &IOTask{
		ID: "cancelled_task",
		Function: func() (Value, error) {
			time.Sleep(1 * time.Second)
			return "result", nil
		},
		Result:  make(chan TaskResult, 1),
		Context: ctx,
	}
	
	err := pool.SubmitIOTask(task)
	if err != nil {
		t.Errorf("Expected I/O task submission to succeed, got error: %v", err)
	}
	
	// Wait for result
	select {
	case result := <-task.Result:
		if result.Error == nil {
			t.Error("Expected I/O task to be cancelled")
		}
	case <-time.After(5 * time.Second):
		t.Error("Test timed out")
	}
}

func TestNewParallelArrayProcessor(t *testing.T) {
	executor := NewConcurrentExecutor(4, 20)
	defer executor.Shutdown()
	
	processor := NewParallelArrayProcessor(executor, 50, 4)
	
	if processor == nil {
		t.Error("Expected processor to be created")
	}
	
	if processor.chunkSize != 50 {
		t.Errorf("Expected chunk size 50, got %d", processor.chunkSize)
	}
	
	if processor.maxParallel != 4 {
		t.Errorf("Expected max parallel 4, got %d", processor.maxParallel)
	}
}

func TestParallelArrayProcessor_ProcessArray(t *testing.T) {
	executor := NewConcurrentExecutor(4, 20)
	defer executor.Shutdown()
	
	processor := NewParallelArrayProcessor(executor, 10, 4)
	
	// Create test array
	input := make([]Value, 100)
	for i := 0; i < 100; i++ {
		input[i] = i
	}
	
	// Process array (double each value)
	result, err := processor.ProcessArray(input, func(v Value) Value {
		if i, ok := v.(int); ok {
			return i * 2
		}
		return v
	})
	
	if err != nil {
		t.Errorf("Expected processing to succeed, got error: %v", err)
	}
	
	if len(result) != len(input) {
		t.Errorf("Expected result length %d, got %d", len(input), len(result))
	}
	
	// Verify results
	for i, v := range result {
		expected := i * 2
		if v != expected {
			t.Errorf("Expected result[%d] = %d, got %v", i, expected, v)
		}
	}
}

func TestParallelArrayProcessor_ProcessArray_EmptyArray(t *testing.T) {
	executor := NewConcurrentExecutor(4, 20)
	defer executor.Shutdown()
	
	processor := NewParallelArrayProcessor(executor, 10, 4)
	
	result, err := processor.ProcessArray([]Value{}, func(v Value) Value {
		return v
	})
	
	if err != nil {
		t.Errorf("Expected processing to succeed, got error: %v", err)
	}
	
	if len(result) != 0 {
		t.Errorf("Expected empty result, got length %d", len(result))
	}
}

func TestConcurrentExecutor_ExecuteTask(t *testing.T) {
	executor := NewConcurrentExecutor(2, 10)
	defer executor.Shutdown()
	
	result, err := executor.ExecuteTask(func() (Value, error) {
		return "task_result", nil
	}, 5*time.Second)
	
	if err != nil {
		t.Errorf("Expected task execution to succeed, got error: %v", err)
	}
	
	if result != "task_result" {
		t.Errorf("Expected result 'task_result', got %v", result)
	}
	
	// Check stats
	stats := executor.GetStats()
	if stats.TasksExecuted != 1 {
		t.Errorf("Expected 1 executed task, got %d", stats.TasksExecuted)
	}
}

func TestConcurrentExecutor_ExecuteIOTask(t *testing.T) {
	executor := NewConcurrentExecutor(2, 10)
	defer executor.Shutdown()
	
	result, err := executor.ExecuteIOTask(func() (Value, error) {
		return "io_result", nil
	})
	
	if err != nil {
		t.Errorf("Expected I/O task execution to succeed, got error: %v", err)
	}
	
	if result != "io_result" {
		t.Errorf("Expected result 'io_result', got %v", result)
	}
	
	// Check stats
	stats := executor.GetStats()
	if stats.AsyncIOOpsCount != 1 {
		t.Errorf("Expected 1 async I/O operation, got %d", stats.AsyncIOOpsCount)
	}
}

func TestConcurrentExecutor_ParallelMapOperation(t *testing.T) {
	executor := NewConcurrentExecutor(4, 20)
	defer executor.Shutdown()
	
	// Create test array
	input := make([]Value, 50)
	for i := 0; i < 50; i++ {
		input[i] = i
	}
	
	// Map operation (square each value)
	result, err := executor.ParallelMapOperation(input, func(v Value) Value {
		if i, ok := v.(int); ok {
			return i * i
		}
		return v
	})
	
	if err != nil {
		t.Errorf("Expected map operation to succeed, got error: %v", err)
	}
	
	if len(result) != len(input) {
		t.Errorf("Expected result length %d, got %d", len(input), len(result))
	}
	
	// Verify results
	for i, v := range result {
		expected := i * i
		if v != expected {
			t.Errorf("Expected result[%d] = %d, got %v", i, expected, v)
		}
	}
}

func TestConcurrentExecutor_ParallelFilterOperation(t *testing.T) {
	executor := NewConcurrentExecutor(4, 20)
	defer executor.Shutdown()
	
	// Create test array
	input := make([]Value, 20)
	for i := 0; i < 20; i++ {
		input[i] = i
	}
	
	// Filter operation (keep even numbers)
	result, err := executor.ParallelFilterOperation(input, func(v Value) bool {
		if i, ok := v.(int); ok {
			return i%2 == 0
		}
		return false
	})
	
	if err != nil {
		t.Errorf("Expected filter operation to succeed, got error: %v", err)
	}
	
	// Should have 10 even numbers (0, 2, 4, ..., 18)
	expectedLength := 10
	if len(result) != expectedLength {
		t.Errorf("Expected result length %d, got %d", expectedLength, len(result))
	}
	
	// Verify all results are even
	for i, v := range result {
		if val, ok := v.(int); ok {
			if val%2 != 0 {
				t.Errorf("Expected even number, got %d at index %d", val, i)
			}
		} else {
			t.Errorf("Expected int value, got %v at index %d", v, i)
		}
	}
}

func TestConcurrentExecutor_ParallelReduceOperation(t *testing.T) {
	executor := NewConcurrentExecutor(4, 20)
	defer executor.Shutdown()
	
	// Create test array
	input := make([]Value, 100)
	for i := 0; i < 100; i++ {
		input[i] = i + 1 // 1, 2, 3, ..., 100
	}
	
	// Reduce operation (sum all values)
	result, err := executor.ParallelReduceOperation(input, func(acc, v Value) Value {
		if accInt, ok := acc.(int); ok {
			if vInt, ok := v.(int); ok {
				return accInt + vInt
			}
		}
		return acc
	}, 0)
	
	if err != nil {
		t.Errorf("Expected reduce operation to succeed, got error: %v", err)
	}
	
	// Sum of 1 to 100 is 5050
	expected := 5050
	if result != expected {
		t.Errorf("Expected result %d, got %v", expected, result)
	}
}

func TestConcurrentExecutor_ParallelReduceOperation_SmallArray(t *testing.T) {
	executor := NewConcurrentExecutor(4, 20)
	defer executor.Shutdown()
	
	// Small array should use sequential processing
	input := []Value{1, 2, 3, 4, 5}
	
	result, err := executor.ParallelReduceOperation(input, func(acc, v Value) Value {
		if accInt, ok := acc.(int); ok {
			if vInt, ok := v.(int); ok {
				return accInt + vInt
			}
		}
		return acc
	}, 0)
	
	if err != nil {
		t.Errorf("Expected reduce operation to succeed, got error: %v", err)
	}
	
	expected := 15 // 1+2+3+4+5
	if result != expected {
		t.Errorf("Expected result %d, got %v", expected, result)
	}
}

func TestConcurrentExecutor_GetStats(t *testing.T) {
	executor := NewConcurrentExecutor(2, 10)
	defer executor.Shutdown()
	
	// Execute some tasks
	executor.ExecuteTask(func() (Value, error) {
		return "result1", nil
	}, 5*time.Second)
	
	executor.ExecuteIOTask(func() (Value, error) {
		return "result2", nil
	})
	
	stats := executor.GetStats()
	
	if stats.TasksExecuted != 1 {
		t.Errorf("Expected 1 executed task, got %d", stats.TasksExecuted)
	}
	
	if stats.AsyncIOOpsCount != 1 {
		t.Errorf("Expected 1 async I/O operation, got %d", stats.AsyncIOOpsCount)
	}
}

func TestConcurrentExecutor_ResetStats(t *testing.T) {
	executor := NewConcurrentExecutor(2, 10)
	defer executor.Shutdown()
	
	// Execute a task
	executor.ExecuteTask(func() (Value, error) {
		return "result", nil
	}, 5*time.Second)
	
	// Reset stats
	executor.ResetStats()
	
	stats := executor.GetStats()
	if stats.TasksExecuted != 0 {
		t.Errorf("Expected stats to be reset, got %d executed tasks", stats.TasksExecuted)
	}
}

// Benchmark tests
func BenchmarkConcurrentExecutor_ExecuteTask(b *testing.B) {
	executor := NewConcurrentExecutor(4, 100)
	defer executor.Shutdown()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.ExecuteTask(func() (Value, error) {
			return i, nil
		}, 5*time.Second)
	}
}

func BenchmarkConcurrentExecutor_ParallelMapOperation(b *testing.B) {
	executor := NewConcurrentExecutor(4, 100)
	defer executor.Shutdown()
	
	input := make([]Value, 1000)
	for i := 0; i < 1000; i++ {
		input[i] = i
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.ParallelMapOperation(input, func(v Value) Value {
			if val, ok := v.(int); ok {
				return val * 2
			}
			return v
		})
	}
}

func BenchmarkParallelArrayProcessor_ProcessArray(b *testing.B) {
	executor := NewConcurrentExecutor(4, 100)
	defer executor.Shutdown()
	
	processor := NewParallelArrayProcessor(executor, 100, 4)
	input := make([]Value, 1000)
	for i := 0; i < 1000; i++ {
		input[i] = i
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor.ProcessArray(input, func(v Value) Value {
			if val, ok := v.(int); ok {
				return val * val
			}
			return v
		})
	}
}

// Test concurrent access
func TestConcurrentExecutor_ConcurrentAccess(t *testing.T) {
	executor := NewConcurrentExecutor(4, 100)
	defer executor.Shutdown()
	
	// Run multiple goroutines concurrently
	var wg sync.WaitGroup
	numGoroutines := 10
	tasksPerGoroutine := 50
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			for j := 0; j < tasksPerGoroutine; j++ {
				result, err := executor.ExecuteTask(func() (Value, error) {
					return id*1000 + j, nil
				}, 5*time.Second)
				
				if err != nil {
					t.Errorf("Task failed: %v", err)
					return
				}
				
				expected := id*1000 + j
				if result != expected {
					t.Errorf("Expected %d, got %v", expected, result)
					return
				}
			}
		}(i)
	}
	
	// Wait for all goroutines to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		// Success
	case <-time.After(30 * time.Second):
		t.Fatal("Test timed out")
	}
	
	// Check final stats
	stats := executor.GetStats()
	expectedTasks := int64(numGoroutines * tasksPerGoroutine)
	if stats.TasksExecuted != expectedTasks {
		t.Errorf("Expected %d executed tasks, got %d", expectedTasks, stats.TasksExecuted)
	}
}