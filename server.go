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
	server   *grpc.Server
	listen   net.Listener
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

	err = s.server.Serve(listen)
	return err
}

func (s *Server) Close() error {
	if s.registry != nil {
		err := s.registry.Close()
		if err != nil {
			return err
		}
	}

	s.server.GracefulStop()
	return nil
}
