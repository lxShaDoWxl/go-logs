package logs

import (
	"context"
	"os"
	"testing"

	"github.com/go-errors/errors"
)

func TestFatalError(t *testing.T) {
	defer recoverTest(context.Background())
	one()
}
func TestInfo(t *testing.T) {
	Info(context.Background(), "TestInfo Message")
}
func TestError(t *testing.T) {
	Error(context.Background(), errors.New("TestError"))
}
func recoverTest(ctx context.Context) {
	if err := recover(); err != nil {
		FatalError(ctx, errors.Wrap(err, 2))
	}
}
func one() {
	two()
}
func two() {
	panic(errors.New("test"))
}

func TestException(t *testing.T) {
	confg := ConfigSentry{
		DSN:         os.Getenv("TESTING_SENTRY_DSN"),
		Environment: "testing",
	}
	initializeSentry(confg)
	Error(context.Background(), NewException(
		errors.New("test exception"),
		map[string]map[string]string{"test": {"test": "test"}},
	))
	Error(context.Background(), NewException(
		NewException(NewException(
			errors.New("test tree exception"),
			map[string]map[string]string{"test3": {"test3": "test3"}},
		), confg),
		map[string]map[string]string{"test": {"test": "test"}},
	))
}
