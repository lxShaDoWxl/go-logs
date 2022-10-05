package logs

import (
	"context"
	"github.com/cockroachdb/errors"
	"testing"
)

func TestFatalError(t *testing.T) {
	defer recoverTest(context.Background())
	one()
}
func TestError(t *testing.T) {
	Error(context.Background(), errors.New("TestError"))
}
func recoverTest(ctx context.Context) {
	if err := recover(); err != nil {
		FatalError(ctx, errors.WithStackDepth(err.(error), 2))
	}
}
func one() {
	two()
}
func two() {
	panic(errors.New("test"))
}
