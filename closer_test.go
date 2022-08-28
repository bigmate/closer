package closer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	ctx := context.Background()
	logger := stdErrLogger{}
	closerFunc := func() error {
		return errors.New("connection closed")
	}

	c := New(
		WithContext(ctx),
		WithLogger(logger),
		WithCloser(closerFunc),
	)

	closerObj, ok := c.(*closer)

	if !ok {
		t.Fatalf("expected *closer implementation, got: %T", c)
	}

	if !reflect.DeepEqual(ctx, closerObj.ctx) {
		t.Error("invalid context stored")
	}

	if !reflect.DeepEqual(logger, closerObj.logger) {
		t.Error("invalid logger stored")
	}

	if len(closerObj.closers) != 1 {
		t.Fatalf("expected exactly 1 closer func to be stored, got: %v", len(closerObj.closers))
	}
}

type testBufLogger struct {
	bytes.Buffer
}

func (t *testBufLogger) Errorf(_ context.Context, format string, args ...interface{}) {
	t.WriteString(fmt.Sprintf(format, args...))
	t.WriteByte('\n')
}

func TestReplaces(t *testing.T) {
	ctxFormer, ctxLater := context.TODO(), context.Background()
	ReplaceContext(ctxFormer)
	ReplaceContext(ctxLater)
	if !reflect.DeepEqual(ctxFormer, cl.ctx) {
		t.Errorf("expected TODO context, got: %v", cl.ctx)
	}

	ReplaceLogger(&testBufLogger{})
	ReplaceLogger(stdErrLogger{})
	if reflect.DeepEqual(cl.logger, stdErrLogger{}) {
		t.Errorf("expected testBufLogger logger, got: %v", cl.logger)
	}
}

func TestAddClose(t *testing.T) {
	logger := &testBufLogger{}
	ReplaceLogger(logger)

	closerFunc := func() error {
		return errors.New("connection closed")
	}

	Add(closerFunc)
	Add(closerFunc)
	Add(closerFunc)

	if len(cl.closers) != 3 {
		t.Errorf("expected exacly 3 closer, have: %v", len(cl.closers))
	}

	Close()
	Close()
	Close()

	logs := strings.Split(strings.TrimRight(logger.String(), "\n"), "\n")

	if len(logs) != 3 {
		t.Errorf("expected exactly 3 lines to be logged, got: %v", len(logs))
		for _, log := range logs {
			t.Log(log)
		}
	}
}
