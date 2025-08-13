package interpreter

import (
	"context"
	"runtime"
	"sync"
	"time"
)

// ConcurrentExecutor provides concurrent execution capabilities
type ConcurrentExecutor struct {
	workerPool   *WorkerPool
	asyncIOPool  *AsyncIOPool
	maxWorkers   int
	maxQueueSize int
	stats        ConcurrentStats
	mutex        sync.RWMutex
}

// ConcurrentStats tracks concurrent execution statistics
type ConcurrentStats struct {
	TasksExecuted     int64
	TasksQueued       int64
	TasksFailed       int64
	ParallelOpsCount  int64
	AsyncIOOpsCount   int64
	WorkerUtilization float64
}

// NewConcurrentExecutor creates a new concurrent executor
func NewConcurrentExecutor(maxWorkers, maxQueueSize int) *ConcurrentExecutor {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}
	if maxQueueSize <= 0 {
		maxQueueSize = maxWorkers * 10
	}
	
	return &ConcurrentExecutor{
		workerPool:   NewWorkerPool(maxWorkers, maxQueueSize),
		asyncIOPool:  NewAsyncIOPool(maxWorkers/2, maxQueueSize),
		maxWorkers:   maxWorkers,
		maxQueueSize: maxQueueSize,
	}
}

// Global concurrent executor
var globalConcurrentExecutor = NewConcurrentExecutor(0, 0)

// GetGlobalConcurrentExecutor returns the global concurrent executor
func GetGlobalConcurrentExecutor() *ConcurrentExecutor {
	return globalConcurrentExecutor
}

// Task represents a unit of work to be executed
type Task struct {
	ID       string
	Function func() (Value, error)
	Result   chan TaskResult
	Context  context.Context
	Timeout  time.Duration
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	Value Value
	Error error
}

// WorkerPool manages a pool of workers for parallel execution
type WorkerPool struct {
	workers     []*Worker
	taskQueue   chan *Task
	quitChannel chan bool
	wg          sync.WaitGroup
	active      bool
	mutex       sync.RWMutex
}

// Worker represents a single worker in the pool
type Worker struct {
	id          int
	taskQueue   chan *Task
	quitChannel chan bool
	wg          *sync.WaitGroup
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(numWorkers, queueSize int) *WorkerPool {
	pool := &WorkerPool{
		workers:     make([]*Worker, numWorkers),
		taskQueue:   make(chan *Task, queueSize),
		quitChannel: make(chan bool),
		active:      true,
	}
	
	// Create and start workers
	for i := 0; i < numWorkers; i++ {
		worker := &Worker{
			id:          i,
			taskQueue:   pool.taskQueue,
			quitChannel: pool.quitChannel,
			wg:          &pool.wg,
		}
		pool.workers[i] = worker
		pool.wg.Add(1)
		go worker.start()
	}
	
	return pool
}

// SubmitTask submits a task to the worker pool
func (wp *WorkerPool) SubmitTask(task *Task) error {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()
	
	if !wp.active {
		return &RuntimeError{Message: "Worker pool is not active"}
	}
	
	select {
	case wp.taskQueue <- task:
		return nil
	default:
		return &RuntimeError{Message: "Task queue is full"}
	}
}

// Shutdown gracefully shuts down the worker pool
func (wp *WorkerPool) Shutdown() {
	wp.mutex.Lock()
	wp.active = false
	close(wp.quitChannel)
	wp.mutex.Unlock()
	
	wp.wg.Wait()
	close(wp.taskQueue)
}

// start starts the worker
func (w *Worker) start() {
	defer w.wg.Done()
	
	for {
		select {
		case task := <-w.taskQueue:
			w.executeTask(task)
		case <-w.quitChannel:
			return
		}
	}
}

// executeTask executes a single task
func (w *Worker) executeTask(task *Task) {
	defer func() {
		if r := recover(); r != nil {
			task.Result <- TaskResult{
				Value: nil,
				Error: &RuntimeError{Message: "Task panicked"},
			}
		}
	}()
	
	// Set up timeout if specified
	ctx := task.Context
	if task.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, task.Timeout)
		defer cancel()
	}
	
	// Execute task with timeout
	done := make(chan TaskResult, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- TaskResult{
					Value: nil,
					Error: &RuntimeError{Message: "Task panicked"},
				}
			}
		}()
		value, err := task.Function()
		done <- TaskResult{Value: value, Error: err}
	}()
	
	select {
	case result := <-done:
		task.Result <- result
	case <-ctx.Done():
		task.Result <- TaskResult{
			Value: nil,
			Error: &RuntimeError{Message: "Task timed out"},
		}
	}
}

// AsyncIOPool manages asynchronous I/O operations
type AsyncIOPool struct {
	workers     []*AsyncIOWorker
	ioQueue     chan *IOTask
	quitChannel chan bool
	wg          sync.WaitGroup
	active      bool
	mutex       sync.RWMutex
}

// AsyncIOWorker represents a worker for async I/O operations
type AsyncIOWorker struct {
	id          int
	ioQueue     chan *IOTask
	quitChannel chan bool
	wg          *sync.WaitGroup
}

// IOTask represents an I/O task
type IOTask struct {
	ID       string
	Function func() (Value, error)
	Result   chan TaskResult
	Context  context.Context
}

// NewAsyncIOPool creates a new async I/O pool
func NewAsyncIOPool(numWorkers, queueSize int) *AsyncIOPool {
	pool := &AsyncIOPool{
		workers:     make([]*AsyncIOWorker, numWorkers),
		ioQueue:     make(chan *IOTask, queueSize),
		quitChannel: make(chan bool),
		active:      true,
	}
	
	// Create and start workers
	for i := 0; i < numWorkers; i++ {
		worker := &AsyncIOWorker{
			id:          i,
			ioQueue:     pool.ioQueue,
			quitChannel: pool.quitChannel,
			wg:          &pool.wg,
		}
		pool.workers[i] = worker
		pool.wg.Add(1)
		go worker.start()
	}
	
	return pool
}

// SubmitIOTask submits an I/O task to the async I/O pool
func (aio *AsyncIOPool) SubmitIOTask(task *IOTask) error {
	aio.mutex.RLock()
	defer aio.mutex.RUnlock()
	
	if !aio.active {
		return &RuntimeError{Message: "Async I/O pool is not active"}
	}
	
	select {
	case aio.ioQueue <- task:
		return nil
	default:
		return &RuntimeError{Message: "I/O task queue is full"}
	}
}

// Shutdown gracefully shuts down the async I/O pool
func (aio *AsyncIOPool) Shutdown() {
	aio.mutex.Lock()
	aio.active = false
	close(aio.quitChannel)
	aio.mutex.Unlock()
	
	aio.wg.Wait()
	close(aio.ioQueue)
}

// start starts the async I/O worker
func (w *AsyncIOWorker) start() {
	defer w.wg.Done()
	
	for {
		select {
		case task := <-w.ioQueue:
			w.executeIOTask(task)
		case <-w.quitChannel:
			return
		}
	}
}

// executeIOTask executes a single I/O task
func (w *AsyncIOWorker) executeIOTask(task *IOTask) {
	defer func() {
		if r := recover(); r != nil {
			task.Result <- TaskResult{
				Value: nil,
				Error: &RuntimeError{Message: "I/O task panicked"},
			}
		}
	}()
	
	select {
	case <-task.Context.Done():
		task.Result <- TaskResult{
			Value: nil,
			Error: &RuntimeError{Message: "I/O task cancelled"},
		}
	default:
		value, err := task.Function()
		task.Result <- TaskResult{Value: value, Error: err}
	}
}

// ParallelArrayProcessor processes arrays in parallel
type ParallelArrayProcessor struct {
	executor    *ConcurrentExecutor
	chunkSize   int
	maxParallel int
}

// NewParallelArrayProcessor creates a new parallel array processor
func NewParallelArrayProcessor(executor *ConcurrentExecutor, chunkSize, maxParallel int) *ParallelArrayProcessor {
	if chunkSize <= 0 {
		chunkSize = 100
	}
	if maxParallel <= 0 {
		maxParallel = runtime.NumCPU()
	}
	
	return &ParallelArrayProcessor{
		executor:    executor,
		chunkSize:   chunkSize,
		maxParallel: maxParallel,
	}
}

// ProcessArray processes an array in parallel using the given function
func (pap *ParallelArrayProcessor) ProcessArray(array []Value, processor func(Value) Value) ([]Value, error) {
	if len(array) == 0 {
		return array, nil
	}
	
	// Calculate number of chunks
	numChunks := (len(array) + pap.chunkSize - 1) / pap.chunkSize
	if numChunks > pap.maxParallel {
		numChunks = pap.maxParallel
		actualChunkSize := (len(array) + numChunks - 1) / numChunks
		pap.chunkSize = actualChunkSize
	}
	
	// Create result array
	result := make([]Value, len(array))
	var wg sync.WaitGroup
	errorChan := make(chan error, numChunks)
	
	// Process chunks in parallel
	for i := 0; i < numChunks; i++ {
		start := i * pap.chunkSize
		end := start + pap.chunkSize
		if end > len(array) {
			end = len(array)
		}
		
		wg.Add(1)
		go func(chunkStart, chunkEnd int) {
			defer wg.Done()
			
			for j := chunkStart; j < chunkEnd; j++ {
				result[j] = processor(array[j])
			}
		}(start, end)
	}
	
	// Wait for all chunks to complete
	wg.Wait()
	close(errorChan)
	
	// Check for errors
	for err := range errorChan {
		if err != nil {
			return nil, err
		}
	}
	
	// Update stats
	pap.executor.mutex.Lock()
	pap.executor.stats.ParallelOpsCount++
	pap.executor.mutex.Unlock()
	
	return result, nil
}

// ExecuteTask executes a task using the concurrent executor
func (ce *ConcurrentExecutor) ExecuteTask(taskFunc func() (Value, error), timeout time.Duration) (Value, error) {
	task := &Task{
		ID:       "task",
		Function: taskFunc,
		Result:   make(chan TaskResult, 1),
		Context:  context.Background(),
		Timeout:  timeout,
	}
	
	err := ce.workerPool.SubmitTask(task)
	if err != nil {
		ce.mutex.Lock()
		ce.stats.TasksFailed++
		ce.mutex.Unlock()
		return nil, err
	}
	
	ce.mutex.Lock()
	ce.stats.TasksQueued++
	ce.mutex.Unlock()
	
	// Wait for result
	result := <-task.Result
	
	ce.mutex.Lock()
	if result.Error != nil {
		ce.stats.TasksFailed++
	} else {
		ce.stats.TasksExecuted++
	}
	ce.mutex.Unlock()
	
	return result.Value, result.Error
}

// ExecuteIOTask executes an I/O task asynchronously
func (ce *ConcurrentExecutor) ExecuteIOTask(ioFunc func() (Value, error)) (Value, error) {
	task := &IOTask{
		ID:       "io_task",
		Function: ioFunc,
		Result:   make(chan TaskResult, 1),
		Context:  context.Background(),
	}
	
	err := ce.asyncIOPool.SubmitIOTask(task)
	if err != nil {
		ce.mutex.Lock()
		ce.stats.TasksFailed++
		ce.mutex.Unlock()
		return nil, err
	}
	
	// Wait for result
	result := <-task.Result
	
	ce.mutex.Lock()
	if result.Error != nil {
		ce.stats.TasksFailed++
	} else {
		ce.stats.AsyncIOOpsCount++
	}
	ce.mutex.Unlock()
	
	return result.Value, result.Error
}

// GetStats returns current concurrent execution statistics
func (ce *ConcurrentExecutor) GetStats() ConcurrentStats {
	ce.mutex.RLock()
	defer ce.mutex.RUnlock()
	return ce.stats
}

// ResetStats resets all concurrent execution statistics
func (ce *ConcurrentExecutor) ResetStats() {
	ce.mutex.Lock()
	ce.stats = ConcurrentStats{}
	ce.mutex.Unlock()
}

// Shutdown gracefully shuts down the concurrent executor
func (ce *ConcurrentExecutor) Shutdown() {
	ce.workerPool.Shutdown()
	ce.asyncIOPool.Shutdown()
}

// ParallelMapOperation performs a map operation on an array in parallel
func (ce *ConcurrentExecutor) ParallelMapOperation(array []Value, mapFunc func(Value) Value) ([]Value, error) {
	processor := NewParallelArrayProcessor(ce, 100, ce.maxWorkers)
	return processor.ProcessArray(array, mapFunc)
}

// ParallelFilterOperation performs a filter operation on an array in parallel
func (ce *ConcurrentExecutor) ParallelFilterOperation(array []Value, filterFunc func(Value) bool) ([]Value, error) {
	if len(array) == 0 {
		return array, nil
	}
	
	// First pass: determine which elements to keep
	keepFlags := make([]bool, len(array))
	processor := NewParallelArrayProcessor(ce, 100, ce.maxWorkers)
	
	// Use ProcessArray to evaluate filter function in parallel
	filterResults, err := processor.ProcessArray(array, func(v Value) Value {
		return filterFunc(v)
	})
	if err != nil {
		return nil, err
	}
	
	// Convert filter results to boolean flags
	for i, result := range filterResults {
		if boolResult, ok := result.(bool); ok {
			keepFlags[i] = boolResult
		}
	}
	
	// Second pass: collect kept elements
	var result []Value
	for i, keep := range keepFlags {
		if keep {
			result = append(result, array[i])
		}
	}
	
	ce.mutex.Lock()
	ce.stats.ParallelOpsCount++
	ce.mutex.Unlock()
	
	return result, nil
}

// ParallelReduceOperation performs a reduce operation on an array in parallel
func (ce *ConcurrentExecutor) ParallelReduceOperation(array []Value, reduceFunc func(Value, Value) Value, initialValue Value) (Value, error) {
	if len(array) == 0 {
		return initialValue, nil
	}
	
	// For small arrays, use sequential processing
	if len(array) < 100 {
		result := initialValue
		for _, v := range array {
			result = reduceFunc(result, v)
		}
		return result, nil
	}
	
	// Parallel reduce using divide and conquer
	chunkSize := len(array) / ce.maxWorkers
	if chunkSize < 1 {
		chunkSize = 1
	}
	
	numChunks := (len(array) + chunkSize - 1) / chunkSize
	partialResults := make([]Value, numChunks)
	var wg sync.WaitGroup
	
	// Process chunks in parallel
	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(array) {
			end = len(array)
		}
		
		wg.Add(1)
		go func(chunkIndex, chunkStart, chunkEnd int) {
			defer wg.Done()
			
			result := initialValue
			for j := chunkStart; j < chunkEnd; j++ {
				result = reduceFunc(result, array[j])
			}
			partialResults[chunkIndex] = result
		}(i, start, end)
	}
	
	wg.Wait()
	
	// Combine partial results
	finalResult := initialValue
	for _, partial := range partialResults {
		finalResult = reduceFunc(finalResult, partial)
	}
	
	ce.mutex.Lock()
	ce.stats.ParallelOpsCount++
	ce.mutex.Unlock()
	
	return finalResult, nil
}