package route

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

// Filter true 留下，false 丢掉
type Filter func(info balancer.PickInfo, addr resolver.Address) bool

type GroupFilterBuilder struct {
	Group string
}

func (g *GroupFilterBuilder) Build() Filter {
	return func(info balancer.PickInfo, addr resolver.Address) bool {
		val := addr.Attributes.Value("group")
		return val == g.Group
	}
}
