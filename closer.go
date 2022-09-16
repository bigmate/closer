package closer

import (
	"context"
	"sync"
)

//Adder is only for dependency injection
type Adder interface {
	Add(closer func() error)
}

type Closer interface {
	Add(closer func() error)
	Close(ctx context.Context) error
}

var (
	globalCloser = &closer{logger: stdErrLogger()}
)

type closer struct {
	rw        sync.RWMutex
	closeOnce sync.Once
	closers   []func() error
	logger    Logger
}

type Option func(c *closer)

func WithCloser(c func() error) Option {
	return func(cl *closer) {
		cl.closers = append([]func() error{c}, cl.closers...)
	}
}

func WithLogger(logger Logger) Option {
	return func(c *closer) {
		c.logger = logger
	}
}

// New is a constructor of closer.
func New(options ...Option) Closer {
	c := &closer{logger: stdErrLogger()}

	for _, apply := range options {
		apply(c)
	}

	return c
}

func (c *closer) Close(ctx context.Context) (err error) {
	c.closeOnce.Do(func() {
		err = c.close(ctx)
	})

	return err
}

func (c *closer) close(ctx context.Context) error {
	done := make(chan struct{}, 1)

	go func() {
		c.rw.RLock()
		defer c.rw.RUnlock()

		wg := sync.WaitGroup{}

		for _, f := range c.closers {
			wg.Add(1)

			go func(closeFunc func() error) {
				defer wg.Done()

				if err := closeFunc(); err != nil {
					c.logger.Errorf("failed to close: %v", err)
				}
			}(f)
		}

		wg.Wait()

		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func (c *closer) reconfigure(options ...Option) {
	c.rw.Lock()
	defer c.rw.Unlock()

	for _, apply := range options {
		apply(c)
	}
}

func (c *closer) Add(closer func() error) {
	c.rw.Lock()
	c.closers = append([]func() error{closer}, c.closers...)
	c.rw.Unlock()
}

// Add adds closer func to the global Closer
// that are run in LIFO order when Close is called.
// Caution: Try not to pass closer func for Logger that is used by Closer,
// otherwise if logger closed before other closers, there is a risk calling closed logger
// if any error occurs down the way closing other closers or Add Logger closer first.
func Add(closer func() error) {
	globalCloser.Add(closer)
}

// Close closes global Closer
// Caution: Only single call makes an affect.
func Close(ctx context.Context) error {
	return globalCloser.Close(ctx)
}

// Reconfigure global closer with provided options
func Reconfigure(options ...Option) {
	globalCloser.reconfigure(options...)
}

// Global returns globally initialized Closer.
func Global() Closer {
	return globalCloser
}
