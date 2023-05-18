package ratelimit

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"time"
)

type TokenBucketLimit struct {
	tokens chan struct{}
	close  chan struct{}
}

func NewTokenBucketLimit(capacity int, interval time.Duration) *TokenBucketLimit {
	res := &TokenBucketLimit{
		tokens: make(chan struct{}, capacity),
		close:  make(chan struct{}),
	}

	// 放令牌
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				select {
				case res.tokens <- struct{}{}:
				default: // 防止桶满放令牌阻塞 无法执行其它分之
				}
			case <-res.close:
				return
			}
		}
	}()

	return res
}

func (t *TokenBucketLimit) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 拿令牌
		select {
		case <-t.close:
			resp, err = handler(ctx, req)
		case <-t.tokens:
			resp, err = handler(ctx, req)
		default:
			err = errors.New("限流了")
		}
		return
	}
}

func (t *TokenBucketLimit) Close() error {
	//t.close <- struct{}{} // 因为有多个地方监听t.close 信号，所以不能采用这种方式
	close(t.close)
	return nil
}

//func (t *TokenBucketLimit) BuildClientInterceptor() grpc.UnaryClientInterceptor {
//}
