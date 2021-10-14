package logs

import (
	"context"
	"os"
	"testing"

	"golang.org/x/xerrors"
)

func TestException(t *testing.T) {
	confg := ConfigSentry{
		DSN:         os.Getenv("TESTING_SENTRY_DSN"),
		Environment: "testing",
	}
	initializeSentry(confg)
	Error(context.Background(), NewException(
		xerrors.New("test exception"),
		map[string]map[string]string{"test": {"test": "test"}},
	))
	Error(context.Background(), NewException(
		NewException(NewException(
			xerrors.New("test tree exception"),
			map[string]map[string]string{"test3": {"test3": "test3"}},
		), confg),
		map[string]map[string]string{"test": {"test": "test"}},
	))
}
