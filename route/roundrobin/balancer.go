package roundrobin

import (
	"errors"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"micro/route"
	"sync/atomic"
)

// Balancer 轮询
type Balancer struct {
	idx         int32
	connections []subConn
	length      int32
	filter      route.Filter
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	candidates := make([]balancer.SubConn, 0, len(b.connections))
	for _, c := range b.connections {
		if b.filter(info, c.info.Address) {
			candidates = append(candidates, c.c)
		}
	}

	if len(candidates) == 0 {
		return balancer.PickResult{}, errors.New("no SubConn")
	}

	idx := atomic.AddInt32(&b.idx, 1)
	c := candidates[idx%b.length]

	return balancer.PickResult{
		SubConn: c,
		Done:    func(info balancer.DoneInfo) {},
	}, nil
}

type Builder struct {
	filter route.Filter
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]subConn, 0, len(info.ReadySCs))

	for c, i := range info.ReadySCs {
		connections = append(connections, subConn{
			c:    c,
			info: i,
		})
	}

	var filter route.Filter = func(info balancer.PickInfo, addr resolver.Address) bool {
		return true
	}

	if b.filter != nil {
		filter = b.filter
	}

	return &Balancer{
		connections: connections,
		idx:         -1,
		length:      int32(len(connections)),
		filter:      filter,
	}
}

type subConn struct {
	c    balancer.SubConn
	info base.SubConnInfo
}
