package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

func init() {
	zapLog, _ := zap.NewDevelopment()
	Log = zapr.NewLogger(zapLog)
}

var (
	Log logr.Logger
)

// Doesn't work properly becuase it doesn't check if the context was manually cancelled
// also blocking is bm and makes it hard to call
func WorkerBlockingSideEffectNo(ctx context.Context) {
	deadline, ok := ctx.Deadline()
	if !ok {
		panic("non-expiring context")
	}

	for time.Now().Before(deadline) {
		fmt.Println(getNulary())
	}
}

// Works ok
// but blocking is bm and makes it hard to call
func WorkerBlockingSideEffect(ctx context.Context) {
	for {
		s := getNulary()

		select {
		case <-ctx.Done():
			Log.Error(ctx.Err(), "stopping")
		default:
			// asycn check
		}

		fmt.Print(s)
	}
}

func WorkerBlockingStream(ctx context.Context, out chan<- int) error {
	// has to take the out stream because it's blocking
	// TODO who closes it? We probably should, otherwise the caller needs to stop itself trying to read from it, which means being async (not range'ing over it)

	for {
		s := getNulary()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// async check
			// Can say `case out <- s` here but hard to read imho
		}

		out <- s
	}
}

func WorkerStream(stopCh <-chan struct{}) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)
		for {
			s := getNulary()

			select {
			case out <- s:
			case <-stopCh:
				Log.Info("Stopped")
				panic("foo")
				return
			}

		}
	}()

	return out
}

func WorkerLump(ctx context.Context) ([]int, error) {
	var result []int

	for i := 0; i < 5; i++ {
		c := getNulary()
		result = append(result, c)

		select {
		case <-ctx.Done():
			return []int{}, ctx.Err()
		default:
		}
	}

	return result, nil
}

// simulated long-lived operation
func getNulary() int {
	time.Sleep(1 * time.Second)
	return 42
}

// simulated long-lived operation
func getUnary(n int) int {
	time.Sleep(1 * time.Second)
	return n * n
}
