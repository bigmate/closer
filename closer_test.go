package closer //nolint // api restrictions

import (
	"context"
	"errors"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	lg := stdErrLogger()
	closerFunc := func() error {
		return errors.New("connection closed")
	}

	c := New(
		WithLogger(lg),
		WithCloser(closerFunc),
	)

	closerObj, ok := c.(*closer)

	if !ok {
		t.Fatalf("expected *closer implementation, got: %T", c)
	}

	if !reflect.DeepEqual(lg, closerObj.logger) {
		t.Error("invalid logger stored")
	}

	if len(closerObj.closers) != 1 {
		t.Fatalf("expected exactly 1 closer func to be stored, got: %v", len(closerObj.closers))
	}
}

func TestReplaces(t *testing.T) {
	ReplaceLogger(nil)
	ReplaceLogger(stdErrLogger())

	if reflect.DeepEqual(globalCloser.logger, stdErrLogger()) {
		t.Errorf("expected testBufLogger logger, got: %v", globalCloser.logger)
	}
}

func TestAdd(t *testing.T) {
	closerFunc := func() error {
		return nil
	}

	Add(closerFunc)
	Add(closerFunc)
	Add(closerFunc)

	if len(globalCloser.closers) != 3 {
		t.Errorf("expected exacly 3 closer, have: %v", len(globalCloser.closers))
	}
}

type nopWriter struct{}

func (nopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestCloseTimeout(t *testing.T) {
	counter := uint32(0)

	closerFunc := func() error {
		time.Sleep(time.Second * 10)
		atomic.AddUint32(&counter, 1)

		return errors.New("connection closed")
	}

	c := New(
		WithCloser(closerFunc),
		WithCloser(closerFunc),
		WithCloser(closerFunc),
		WithLogger(logger{nopWriter{}}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := c.Close(ctx); err != ctx.Err() {
		t.Errorf("expcted error, got: %v", err)
	}

	if err := c.Close(ctx); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := c.Close(ctx); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	readCounter := atomic.LoadUint32(&counter)
	if readCounter != 0 {
		t.Errorf("expected counter to be 0, got: %v", readCounter)
	}
}

func TestCloseWithoutTimeout(t *testing.T) {
	counter := uint32(0)
	expected := 10

	closerFunc := func() error {
		time.Sleep(time.Second * 2)
		atomic.AddUint32(&counter, 1)

		return nil
	}

	c := New(WithLogger(logger{nopWriter{}}))

	for i := 0; i < expected; i++ {
		c.Add(closerFunc)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if err := c.Close(ctx); err != ctx.Err() {
		t.Errorf("expected error, got: %v", err)
	}

	if err := c.Close(ctx); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	readCounter := atomic.LoadUint32(&counter)
	if readCounter != uint32(expected) {
		t.Errorf("expected counter to be %v, got: %v", expected, readCounter)
	}
}
