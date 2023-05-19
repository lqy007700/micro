package ratelimit

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"sync/atomic"
	"time"
)

type FixWindowLimiter struct {
	timestamp int64 // 窗口起始位置
	interval  int64 // 窗口长度
	rate      int64 // 总请求数
	cnt       int64 // 当前请求数
}

func (f *FixWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		current := time.Now().UnixNano()

		timestamp := atomic.LoadInt64(&f.timestamp)
		cnt := atomic.LoadInt64(&f.cnt)
		if timestamp+f.interval < current {
			if atomic.CompareAndSwapInt64(&f.timestamp, timestamp, current) {
				atomic.CompareAndSwapInt64(&f.cnt, cnt, 0)
			}
		}

		cnt = atomic.AddInt64(&f.cnt, 1)
		defer atomic.AddInt64(&f.cnt, -1)
		if cnt > f.rate {
			err = errors.New("err")
			return
		}

		f.cnt++
		resp, err = handler(ctx, req)
		return
	}
}
