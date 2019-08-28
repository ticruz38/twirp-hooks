package logging

import (
	"context"
	"log"

	"time"

	"io/ioutil"

	"github.com/google/logger"
	"github.com/twitchtv/twirp"
)

var timeKey = new(int)
var logs *logger.Logger

// Logging logs request information to the console
func Logging(verbose *bool) *twirp.ServerHooks {
	if logs == nil {
		logs = logger.Init("logs", *verbose, true, ioutil.Discard)
		logger.SetFlags(log.LstdFlags)
	}
	hooks := &twirp.ServerHooks{}
	hooks.RequestReceived = func(ctx context.Context) (context.Context, error) {
		ctx = context.WithValue(ctx, timeKey, time.Now())
		return ctx, nil
	}
	hooks.Error = func(ctx context.Context, err twirp.Error) context.Context {
		svc, _ := twirp.ServiceName(ctx)
		method, _ := twirp.MethodName(ctx)
		duration := time.Since(ctx.Value(timeKey).(time.Time))
		logs.Errorf("Service: %v, Method: %v, Duration: %v ms, Code: %v, Message: %v", svc, method, duration/1000, err.Code(), err.Msg())
		ctx = context.WithValue(ctx, timeKey, nil)
		return ctx
	}
	hooks.ResponseSent = func(ctx context.Context) {
		if ctx.Value(timeKey) == nil {
			return
		}
		svc, _ := twirp.ServiceName(ctx)
		method, _ := twirp.MethodName(ctx)
		duration := time.Since(ctx.Value(timeKey).(time.Time))
		logs.Infof("Service: %v, Method: %v, Duration: %v ms", svc, method, duration)
	}
	return hooks
}
