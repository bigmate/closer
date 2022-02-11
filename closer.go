package closer

import (
	"context"
	"log"
	"sync"

	"github.com/bigmate/app"
)

var cl *closer

func init() {
	cl = &closer{
		resourceClosers: make([]func() error, 0, 10),
	}
}

type closer struct {
	rw              sync.RWMutex
	resourceClosers []func() error
}

// NewCloser is a constructor of closer
func NewCloser() app.App {
	return cl
}

func (c *closer) close() {
	c.rw.RLock()
	defer c.rw.RUnlock()

	for _, closeResource := range c.resourceClosers {
		if err := closeResource(); err != nil {
			log.Printf("failed to close: %v\n", err)
		}
	}
}

// Run implements app.App interface
func (c *closer) Run(ctx context.Context) error {
	<-ctx.Done()
	c.close()
	return nil
}

// Add adds resource releaser func to the collections of functions
// that are run in FIFO order before application exits
func Add(c func() error) {
	cl.rw.Lock()
	defer cl.rw.Unlock()

	cl.resourceClosers = append(cl.resourceClosers, c)
}
