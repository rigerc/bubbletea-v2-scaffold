// Package task provides primitives for running background work inside a
// BubbleTea program and routing results back through the message loop.
package task

// DoneMsg carries a successfully completed task result.
// T is the value type returned by the task function.
type DoneMsg[T any] struct {
	Label string
	Value T
}

// ErrMsg carries a failed or cancelled task error.
// Err may be context.Canceled if the root context was cancelled.
type ErrMsg struct {
	Label string
	Err   error
}

// ProgressMsg carries incremental progress updates (Progress in 0.0–1.0).
type ProgressMsg struct {
	Label    string
	Progress float64
}
