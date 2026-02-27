package task

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
)

// Result is an internal carrier for goroutine results.
type Result[T any] struct {
	Value T
	Err   error
	Label string
}

// Run executes fn in a goroutine and returns a tea.Cmd that resolves to
// DoneMsg[T] on success or ErrMsg on failure/cancellation.
// If ctx is cancelled before fn returns, ErrMsg{Err: ctx.Err()} is sent.
func Run[T any](ctx context.Context, label string, fn func(context.Context) (T, error)) tea.Cmd {
	return func() tea.Msg {
		done := make(chan Result[T], 1)
		go func() {
			v, err := fn(ctx)
			done <- Result[T]{Value: v, Err: err, Label: label}
		}()
		select {
		case r := <-done:
			if r.Err != nil {
				return ErrMsg{Label: label, Err: r.Err}
			}
			return DoneMsg[T]{Label: label, Value: r.Value}
		case <-ctx.Done():
			return ErrMsg{Label: label, Err: ctx.Err()}
		}
	}
}

// RunWithTimeout is like Run but derives a timeout context from ctx.
// The timeout context is cancelled when fn returns or after d, whichever comes first.
func RunWithTimeout[T any](ctx context.Context, label string, d time.Duration, fn func(context.Context) (T, error)) tea.Cmd {
	return func() tea.Msg {
		tctx, cancel := context.WithTimeout(ctx, d)
		defer cancel()
		done := make(chan Result[T], 1)
		go func() {
			v, err := fn(tctx)
			done <- Result[T]{Value: v, Err: err, Label: label}
		}()
		select {
		case r := <-done:
			if r.Err != nil {
				return ErrMsg{Label: label, Err: r.Err}
			}
			return DoneMsg[T]{Label: label, Value: r.Value}
		case <-tctx.Done():
			return ErrMsg{Label: label, Err: tctx.Err()}
		}
	}
}
