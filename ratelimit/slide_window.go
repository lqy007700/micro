package ratelimit

import (
	"container/list"
	"context"
	"errors"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type SlideWindowLimiter struct {
	queue    *list.List // 前面的请求
	interval int64      // 窗口长度
	rate     int64      // 最大请求
	mu       sync.Mutex
}

func (s *SlideWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		now := time.Now().UnixNano()
		boundary := now - s.interval

		s.mu.Lock()
		// 快路径
		length := s.queue.Len()
		if int64(length) < s.rate {
			resp, err = handler(ctx, req)
			s.queue.PushBack(now)
			s.mu.Unlock()
			return
		}

		// 窗口满了
		timestamp := s.queue.Front()
		for timestamp != nil && timestamp.Value.(int64) < boundary {
			s.queue.Remove(timestamp)
			timestamp = s.queue.Front()
		}

		length = s.queue.Len()
		s.mu.Unlock()
		if int64(length) > s.rate {
			err = errors.New("到达瓶颈")
			return
		}

		resp, err = handler(ctx, req)
		s.queue.PushBack(now)
		return
	}
}
