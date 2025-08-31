package routine

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/platform/logger"
	"github.com/tucanbit/platform/wait"
	"go.uber.org/zap"
)

// Routine contains the information of a routine.
type Routine struct {
	// Name is the name of the routine.
	Name string
	// Operation is the operation to be executed.
	Operation func(ctx context.Context, log logger.Logger)
	// NoWait indicates to not add the routine to the wait group.
	NoWait bool
	// Timeout is the timeout of the routine.
	//
	// The default value is 5 minute.
	Timeout time.Duration
	// OnCancelled is the function to be executed when the routine is cancelled.
	OnCancelled func(ctx context.Context, log logger.Logger)
	// RandomDelayMax can be set to add a random delay
	// in the range of 0 to RandomDelayMax before executing the routine.
	//
	// No delay is added if RandomDelayMax is 0.
	RandomDelayMax time.Duration
}

// ExecuteRoutine executes a routine and recovers from panic.
// It also adds the routine to the wait group if NoWait is false.
// The routine is cancelled if it exceeds the timeout specified.
func ExecuteRoutine(ctx context.Context, routine Routine, log logger.Logger) {
	go GetExecutableRoutine(ctx, routine, log)()
}

// GetExecutableRoutine returns a routine that can be executed with panic safety and graceful shutdown.
func GetExecutableRoutine(ctx context.Context, routine Routine, log logger.Logger) func() {
	return func() {
		func(wtGroup *sync.WaitGroup) {
			if !routine.NoWait {
				wait.RoutineWaitGroup.Add(1)
			}

			defer func() {
				if r := recover(); r != nil {
					log.Error(ctx, "routine parent panicked",
						zap.String("routine-name", routine.Name),
						zap.Any("panic", r))
				}
			}()

			log := log.Named("routine")
			if routine.Timeout == 0 {
				routine.Timeout = 5 * time.Minute
			}
			ctxWithTimeout, cancel := context.WithTimeout(context.Background(), routine.Timeout)

			newCtx := context.WithValue(context.WithValue(
				context.WithValue(
					context.WithValue(
						ctxWithTimeout, constant.ContextKey("request-start-time"), time.Now(),
					), constant.ContextKey("x-ws-request-id"), ctx.Value("x-ws-request-id"),
				), constant.ContextKey("x-user"), ctx.Value("x-user"),
			), constant.ContextKey("x-request-id"), ctx.Value("x-request-id"))

			defer func() {
				cancel()
				if !routine.NoWait {
					wtGroup.Done()
				}
			}()

			c := make(chan int)

			go func(ctx context.Context, log logger.Logger) {
				defer func() {
					if r := recover(); r != nil {
						log.Error(ctx, "routine panicked",
							zap.String("routine-name", routine.Name),
							zap.Any("panic", r))
						c <- 1
					}
				}()

				if routine.RandomDelayMax > 0 {
					delay := time.Duration(
						rand.Int63n(int64(routine.RandomDelayMax))) //nolint:gosec // no need for secure random
					time.Sleep(delay)
					log.Info(ctx, "routine delayed",
						zap.String("routine-name", routine.Name),
						zap.Any("random-delay", delay),
						zap.Any("random-delay-max", routine.RandomDelayMax))
				}

				routine.Operation(ctx, log)
				c <- 1
			}(newCtx, log)

			if !routine.NoWait {
				select {
				case <-newCtx.Done():
					if routine.OnCancelled != nil {
						routine.OnCancelled(context.WithValue(context.WithValue(
							context.WithValue(
								context.WithValue(
									context.Background(), constant.ContextKey("request-start-time"), time.Now(),
								), constant.ContextKey("x-ws-request-id"), ctx.Value("x-ws-request-id"),
							), constant.ContextKey("x-user"), ctx.Value("x-user"),
						), constant.ContextKey("x-request-id"), ctx.Value("x-request-id")), log)
					}
					log.Error(newCtx, "routine cancelled", zap.String("routine-name", routine.Name))
				case <-c:
				}
			} else {
				<-c
			}
		}(wait.RoutineWaitGroup)
	}
}
