package logs

import (
	"context"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/blake2b"
	"os"
	"testing"

	"github.com/go-errors/errors"
)

func TestFatalError(t *testing.T) {
	defer recoverTest(context.Background())
	one()
}
func TestInfo(t *testing.T) {
	inputString1 := "/test/test2"
	inputString2 := "/test/test3"

	// Get the BLAKE2b hash of the input strings
	hash1 := blake2b.Sum256([]byte(inputString1))
	hash2 := blake2b.Sum256([]byte(inputString2))

	// Encode the hash to base64
	hashBase641 := base64.StdEncoding.EncodeToString(hash1[:])
	hashBase642 := base64.StdEncoding.EncodeToString(hash2[:])

	fmt.Println(hashBase641)
	fmt.Println(hashBase642)

	// Compare the beginning of the base64-encoded hashes
	if hashBase641[:8] == hashBase642[:8] {
		fmt.Println("The hashes have the same beginning!")
	} else {
		fmt.Println("The hashes are different.")
	}
	//Info(context.Background(), "TestInfo Message")
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
	//Error(context.Background(), NewException(
	//	errors.New("test exception"),
	//	map[string]map[string]string{"test": {"test": "test"}},
	//))
	Error(context.Background(), NewExceptionWithMeta(
		NewExceptionWithMeta(NewExceptionWithMeta(
			errors.New("test tree exception"),
			"test3", map[string]string{"test3": "test3"},
		), "config", confg),
		"test", map[string]string{"test": "test"},
	))
	Error(context.Background(), NewExceptionWithMeta(errors.New("exception meta nil")))
}
