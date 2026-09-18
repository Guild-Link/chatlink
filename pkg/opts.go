package chatlink

import "time"

type Option func(*Worker)

type MessageHandler func(*Client, string)

// WithTimeout sets the timeout for worker operations.
func WithTimeout(timeout time.Duration) Option {
	return func(w *Worker) {
		w.timeout = timeout
	}
}

// WithMessageHandler adds a message handler to the worker.
func WithMessageHandler(handler MessageHandler) Option {
	return func(w *Worker) {
		w.messageHandlers = append(w.messageHandlers, handler)
	}
}
