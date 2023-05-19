package ratelimit

import (
	"context"
	"google.golang.org/grpc"
	"time"
)

type LeakyBucketLimit struct {
	producer *time.Ticker
}

func NewLeakyBucketLimit(interval time.Duration) *LeakyBucketLimit {
	res := &LeakyBucketLimit{
		producer: time.NewTicker(interval),
	}
	return res
}

func (l *LeakyBucketLimit) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		select {
		case <-l.producer.C:
			resp, err = handler(ctx, req)
		}
		return
	}
}
