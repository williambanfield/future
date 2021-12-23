package future

import (
	"context"
	"errors"
)

var (
	ErrAlreadyComplete = errors.New("future already complete")
	Canceled           = errors.New("future canceled")
)

var (
	_ Future[any] = (*Completeable[any])(nil)
)

// Future is an interface for interacting with data that might not exist yet.
type Future[V any] interface {
	Get(context.Context) (V, error)
	IsCancelled() bool
	Cancel()
}

// Completeable provides a container for data that may not exist yet.
// Completable satisfies the Future interface and extends it with the `Complete`
// method that allows for the value to be filled in and for anyone waiting on the Future
// to retrieve the value.
type Completeable[V any] struct {
	coch chan struct{}
	cach chan struct{}
	val  V
}

func NewCompleatableFuture[V any]() *Completeable[V] {
	return &Completeable[V]{
		coch: make(chan struct{}),
		cach: make(chan struct{}),
	}
}

// Get retrieves the value held by the Completeable. If the Completeable is 
// canceled, it will return an error indicating that it was canceled. 
// The context parameter passed to Get allows the caller to bail out of their
// call to Get before receiving the value.
//
// Canceling the context passed to Get will not cancel the entire
// Completeable. 
func (c *Completeable[V]) Get(ctx context.Context) (V, error) {
	select {
	case <-c.coch:
		return c.val, nil
	case <-c.cach:
		var res V
		return res, Canceled
	case <-ctx.Done():
		var res V
		return res, ctx.Err()
	}
}

// Cancel marks the completeable as canceled and notifies any other callers waiting
// for the Completeable that it is canceled. 
func (c *Completeable[V]) Cancel() {
	close(c.cach)
}

// Complete marks the Completeable as complete and notifies any callers waiting
// for the completeable that its value is ready. 
func (c *Completeable[V]) Complete(val V) error {
	if c.isComplete() {
		return ErrAlreadyComplete
	}
	if c.IsCancelled() {
		return Canceled
	}
	c.val = val
	close(c.coch)
	return nil
}

// IsCancelled allows consumers of a Completeable to check if it was canceled.
// It is useful to check if the completeable was canceled before performing a
// slow or blocking operations to prevent a routine with a completeable from
// doig work that will ultimately not be observed by any other routine. 
func (c *Completeable[V]) IsCancelled() bool {
	select {
	case <-c.cach:
		return true
	default:
		return false
	}
}

func (c *Completeable[V]) isComplete() bool {
	select {
	case <-c.coch:
		return true
	default:
		return false
	}
}
