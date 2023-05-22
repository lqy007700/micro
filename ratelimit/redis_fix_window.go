package ratelimit

import (
	"context"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

//go::embed lua/fix_window.lua
var luaFixWindow string

type RedisFixWindowLimiter struct {
	client redis.Cmdable
}

func (f *RedisFixWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		limit, err := f.limit(ctx, "123")
		if err != nil || !limit {
			return
		}

		resp, err = handler(ctx, req)
		return
	}
}

func (f *RedisFixWindowLimiter) limit(ctx context.Context, service string) (bool, error) {
	return f.client.Eval(ctx, luaFixWindow, []string{service}).Bool()
}
