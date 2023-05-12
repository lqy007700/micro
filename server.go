package micro

import (
	"context"
	"google.golang.org/grpc"
	"micro/registry"
	"net"
)

type Server struct {
	name     string
	registry registry.Registry
	*grpc.Server
	listen net.Listener

	weight uint32
}

type ServerOpt func(server *Server)

func NewServer(name string, opts ...ServerOpt) (*Server, error) {
	s := &Server{
		name:   name,
		Server: grpc.NewServer(),
	}

	for _, opt := range opts {
		opt(s)
	}
	return s, nil
}

func ServerWithRegistry(r registry.Registry) ServerOpt {
	return func(server *Server) {
		server.registry = r
	}
}

func ServerWithWeight(w uint32) ServerOpt {
	return func(server *Server) {
		server.weight = w
	}
}

func (s *Server) Start(ctx context.Context, addr string) error {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// 使用注册中心
	if s.registry != nil {
		err = s.registry.Register(ctx, registry.ServiceInstance{
			Name:    s.name,
			Address: listen.Addr().String(),
		})
		if err != nil {
			return err
		}
		//defer func() {
		//	_ = s.registry.UnRegister(ctx, registry.ServiceInstance{})
		//}()
	}

	err = s.Serve(listen)
	return err
}

func (s *Server) Close() error {
	if s.registry != nil {
		err := s.registry.Close()
		if err != nil {
			return err
		}
	}

	s.GracefulStop()
	return nil
}
