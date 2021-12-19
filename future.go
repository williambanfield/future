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

type Future[V any] interface {
	Get(context.Context) (V, error)
	IsCancelled() bool
	Cancel()
}

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

func (c *Completeable[V]) Cancel() {
	close(c.cach)
}

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
