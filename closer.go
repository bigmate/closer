package closer

import (
	"context"
	"sync"
)

type Closer interface {
	Close()
	Add(closer func() error)
	Run(ctx context.Context) error
}

var (
	onceContext sync.Once
	onceLogger  sync.Once
	cl          *closer
)

func init() {
	cl = &closer{
		ctx:    context.Background(),
		logger: stdErrLogger{},
	}
}

type closer struct {
	ctx       context.Context
	rw        sync.RWMutex
	closeOnce sync.Once
	closers   []func() error
	logger    Logger
}

type Option func(c *closer)

func WithCloser(c func() error) Option {
	return func(cl *closer) {
		cl.closers = append(cl.closers, c)
	}
}

func WithLogger(logger Logger) Option {
	return func(c *closer) {
		c.logger = logger
	}
}

func WithContext(ctx context.Context) Option {
	return func(c *closer) {
		c.ctx = ctx
	}
}

// New is a constructor of closer
func New(options ...Option) Closer {
	c := &closer{
		ctx:    context.Background(),
		logger: stdErrLogger{},
	}

	for _, apply := range options {
		apply(c)
	}

	return c
}

func (c *closer) Close() {
	c.closeOnce.Do(func() {
		c.rw.RLock()
		defer c.rw.RUnlock()

		for _, closeResource := range c.closers {
			if err := closeResource(); err != nil {
				c.logger.Errorf(c.ctx, "failed to close: %v", err)
			}
		}
	})
}

func (c *closer) Add(closer func() error) {
	c.rw.Lock()
	c.closers = append(c.closers, closer)
	c.rw.Unlock()
}

// Run implements app.App interface
func (c *closer) Run(ctx context.Context) error {
	<-ctx.Done()
	c.Close()
	return nil
}

// Add adds closer func to the global Closer
// that are run in FIFO order when Close is called.
// Caution: Try not to pass closer func for Logger that is used by Closer,
// otherwise if logger closed before other closers, there is a risk calling closed logger
// if any error occurs down the way closing other closers
func Add(closer func() error) {
	cl.Add(closer)
}

// Close closes global Closer
// Caution: Only single call makes an affect
func Close() {
	cl.Close()
}

// ReplaceLogger replaces Logger of global Closer
// Caution: Only single call makes an affect
func ReplaceLogger(logger Logger) {
	onceLogger.Do(func() {
		cl.rw.Lock()
		cl.logger = logger
		cl.rw.Unlock()
	})
}

// ReplaceContext replaces context.Context of global Closer
// Caution: Only single call makes an affect
func ReplaceContext(ctx context.Context) {
	onceContext.Do(func() {
		cl.rw.Lock()
		cl.ctx = ctx
		cl.rw.Unlock()
	})
}

// GlobalCloser returns globally initialized Closer
func GlobalCloser() Closer {
	return cl
}
