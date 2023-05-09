package roundrobin

import (
	"errors"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync/atomic"
)

// Balancer 轮询
type Balancer struct {
	idx         int32
	connections []balancer.SubConn
	length      int32
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if b.length == 0 {
		return balancer.PickResult{}, errors.New("no SubConn")
	}

	idx := atomic.AddInt32(&b.idx, 1)
	c := b.connections[idx%b.length]

	return balancer.PickResult{
		SubConn: c,
		Done:    func(info balancer.DoneInfo) {},
	}, nil
}

type Builder struct {
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]balancer.SubConn, 0, len(info.ReadySCs))

	for c := range info.ReadySCs {
		connections = append(connections, c)
	}

	return &Balancer{
		connections: connections,
		idx:         -1,
		length:      int32(len(connections)),
	}
}
