package workman

import (
	"errors"
	"reflect"
	"sync"
)

var (
	// The user provided an invalid number of workers
	ErrTooFewWorkers = errors.New("number of workers must exceed 0")

	// The user provided an invalid work function
	ErrInvalidWorkFuncType = errors.New("invalid work function type")

	// Workers have already been started and cannot start again
	ErrAlreadyStarted = errors.New("workers already started")

	// Work has already completed and no more work can be done
	ErrWorkCompleted = errors.New("work has completed")

	// Some workers encountered errors during processing
	ErrWorkerErrors = errors.New("some workers encountered errors")
)

// WorkManagerError stores a list of errors encountered during worker processing.
type WorkManagerError struct {
	error
	errs []error
}

// All returns a list of all errors encountered by workers.
func (e *WorkManagerError) All() []error { return e.errs }

// add is used by the WorkManager to append errors to the error list (not thread-safe)
func (e *WorkManagerError) add(err error) { e.errs = append(e.errs, err) }

// A task is the job given to a worker. It stores the function args
// and the returned error from the work function.
type task struct {
	args []interface{}
	err  error
}

// WorkManager manages a pool of parallel workers and provides
// an API for sending work and collecting errors. It offers a layer
// of abstraction over the concept of workers and the asynchronous
// nature of its processing.
//
// A WorkManager is required to call three methods:
//  - StartWorkers(func (args...))
// 	Pass a function (or method with pointer receiver) to inform
//	the WorkManager how to process the work you will send it.
//	This is a normal function that takes some arbitrary number
//	of arguments. Note: the function passed *must* only return
//	an error. It is also unsafe for the function passed to
//	read or write to any shared state.
//  - SendWork(args...)
//	Pass arguments as you would normally pass to the function
//	previously given to StartWorkers.
//  - WaitForCompletion()
//	Wait for all workers to complete their tasks and receive a
//	list of any errors that were encountered during processing.
//
// The WorkManager is stateful. It is an error to run its methods
// out of order or to send it work after it has already completed
// all work.
type WorkManager struct {
	// wg manages final coordination among workers
	wg sync.WaitGroup
	// nWorkers is the number of simultaneous goroutines to run
	nWorkers int

	// tasks holds the new work, waiting to be consumed
	tasks chan task

	// results holds the final work product from the workers
	results chan task

	// errors holds a list of errors from workers
	errors *WorkManagerError

	// hasStarted indicates whether the manager has started its workers
	hasStarted bool

	// hasCompleted indicates whether all workers have finished their work
	hasCompleted bool
}

// start changes the work manager's state to started
func (wm *WorkManager) start() { wm.hasStarted = true }

// finish changes the work manager's state to finished
func (wm *WorkManager) finish() { wm.hasCompleted = true }

// NewWorkManager returns a WorkManager with n parallel workers.
func NewWorkManager(n int) (WorkManager, error) {
	if n < 1 {
		return WorkManager{}, ErrTooFewWorkers
	}
	wm := WorkManager{}
	wm.nWorkers = n
	wm.tasks = make(chan task, wm.nWorkers)
	wm.results = make(chan task, wm.nWorkers)
	return wm, nil
}

// StartWorkers starts a pool of workers that will run workFunc.
func (wm *WorkManager) StartWorkers(workFunc interface{}) error {
	if wm.hasStarted {
		return ErrAlreadyStarted
	}

	// Initialize the worker function from the passed interface
	worker := reflect.ValueOf(workFunc)
	if worker.Kind() != reflect.Func {
		return ErrInvalidWorkFuncType
	}

	// Start goroutine workers that receive on the task queue and push to the output queue
	for i := 0; i < wm.nWorkers; i++ {
		go func() {
			for t := range wm.tasks {
				// Convert the variadic task args into a slice of reflected Values
				args := make([]reflect.Value, len(t.args))
				for i := range t.args {
					args[i] = reflect.ValueOf(t.args[i])
				}

				// Call the work function with the args
				rv := worker.Call(args)

				// Parse the returned error if there is one
				if len(rv) > 0 {
					err := rv[0]
					if !err.IsNil() {
						t.err = err.Interface().(error)
					}
				}

				// Send the task into the results queue for futher processing
				wm.results <- t
			}
		}()
	}

	// Start output manager to serialize access to the resulting errors slice
	go func() {
		for output := range wm.results {
			if output.err != nil {
				if wm.errors == nil {
					wm.errors = &WorkManagerError{error: ErrWorkerErrors}
				}
				wm.errors.add(output.err)
			}
			wm.wg.Done()
		}
	}()

	wm.start()
	return nil
}

// SendWork provides the args necessary for the workers to run their workFunc.
func (wm *WorkManager) SendWork(args ...interface{}) error {
	if wm.hasCompleted {
		return ErrWorkCompleted
	}
	wm.wg.Add(1)
	go func() {
		wm.tasks <- task{args: args}
	}()
	return nil
}

// WaitForCompletion blocks until all workers have completed their work.
// The returned error is non-nil if any worker encountered an error. The
// exact error list is returned by calling err.All()
func (wm *WorkManager) WaitForCompletion() *WorkManagerError {
	if wm.hasCompleted {
		return &WorkManagerError{error: ErrWorkCompleted}
	}
	defer wm.finish()
	defer close(wm.tasks)
	defer close(wm.results)
	wm.wg.Wait()
	return wm.errors
}
