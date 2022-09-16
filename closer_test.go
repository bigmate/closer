package closer_test

import (
	"context"
	"fmt"
	"github.com/bigmate/closer"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type logger struct {
	mu   sync.Mutex
	dest io.Writer
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	fmt.Fprintf(l.dest, format, args...)
	fmt.Fprintln(l.dest)
}

func ensureWithLogger(t *testing.T, linesCount int, buf *strings.Builder) {
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")

	if len(lines) != linesCount {
		t.Fatalf("expected %v lines got: %v", linesCount, len(lines))
	}

	for _, line := range lines {
		if line != "failed to close: connection closed" {
			t.Errorf("unexpected log: %s", line)
		}
	}
}

func ensureNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func ensureError(t *testing.T, err error) {
	if err == nil {
		t.Errorf("expected err, got nil")
	}
}

func closerErrorFunc() error {
	return fmt.Errorf("connection closed")
}

func TestOptions(t *testing.T) {
	buf := &strings.Builder{}
	log := &logger{dest: buf}

	c := closer.New(
		closer.WithLogger(log),
		closer.WithCloser(closerErrorFunc),
	)

	c.Add(closerErrorFunc)

	ensureNoError(t, c.Close(context.Background()))
	ensureWithLogger(t, 2, buf)
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

		return fmt.Errorf("connection closed")
	}

	c := closer.New(
		closer.WithCloser(closerFunc),
		closer.WithCloser(closerFunc),
		closer.WithCloser(closerFunc),
		closer.WithLogger(&logger{dest: nopWriter{}}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	ensureError(t, c.Close(ctx))
	ensureNoError(t, c.Close(ctx))
	ensureNoError(t, c.Close(ctx))

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

	c := closer.New(closer.WithLogger(&logger{dest: nopWriter{}}))

	for i := 0; i < expected; i++ {
		c.Add(closerFunc)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	ensureNoError(t, c.Close(ctx))
	ensureNoError(t, c.Close(ctx))
	ensureNoError(t, c.Close(ctx))

	readCounter := atomic.LoadUint32(&counter)
	if readCounter != uint32(expected) {
		t.Errorf("expected counter to be %v, got: %v", expected, readCounter)
	}
}

func TestReconfigure(t *testing.T) {
	buf := &strings.Builder{}
	log := &logger{dest: buf}

	closer.Reconfigure(
		closer.WithCloser(closerErrorFunc),
		closer.WithCloser(closerErrorFunc),
		closer.WithCloser(closerErrorFunc),
		closer.WithLogger(log),
	)

	ensureNoError(t, closer.Close(context.Background()))
	ensureWithLogger(t, 3, buf)
}

func TestAdd(t *testing.T) {
	buf := &strings.Builder{}
	log := &logger{dest: buf}
	count := 17

	closer.Reconfigure(closer.WithLogger(log))

	for i := 0; i < count; i++ {
		closer.Add(closerErrorFunc)
	}

	ensureNoError(t, closer.Close(context.Background()))
	ensureWithLogger(t, count, buf)
}
