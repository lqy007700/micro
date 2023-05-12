package micro

import (
	"context"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"micro/registry"
)

type grpcResolverBuilder struct {
	r registry.Registry
}

func NewRegistryBuilder(r registry.Registry) resolver.Builder {
	return &grpcResolverBuilder{
		r: r,
	}
}

func (b *grpcResolverBuilder) Build(target resolver.Target,
	cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &grpcResolver{
		cc: cc,
		r:  b.r,
		t:  target,
	}

	r.resolve()
	go r.watch()
	return r, nil
}

func (b *grpcResolverBuilder) Scheme() string {
	return "registry"
}

type grpcResolver struct {
	t     resolver.Target
	cc    resolver.ClientConn
	r     registry.Registry
	close chan struct{}
}

func (g *grpcResolver) ResolveNow(options resolver.ResolveNowOptions) {
	g.resolve()
}

func (g *grpcResolver) watch() {
	events, err := g.r.Subscribe(g.t.Endpoint())
	if err != nil {
		g.cc.ReportError(err)
		return
	}
	for {
		select {
		case <-events:
			g.resolve()
		case <-g.close:
			return
		}
	}
}

func (g *grpcResolver) resolve() {
	ctx := context.Background()
	services, err := g.r.ListServices(ctx, g.t.Endpoint())
	if err != nil {
		g.cc.ReportError(err)
		return
	}
	addr := make([]resolver.Address, 0, len(services))
	for _, service := range services {
		addr = append(addr, resolver.Address{
			Addr:       service.Address,
			Attributes: attributes.New("weight", service.Weight),
		})
	}
	err = g.cc.UpdateState(resolver.State{Addresses: addr})
	if err != nil {
		g.cc.ReportError(err)
		return
	}
}

func (g *grpcResolver) Close() {
	close(g.close)
}
