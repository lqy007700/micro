package grpc

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"micro/proto/gen"
	"net"
	"testing"
)

func TestServer(t *testing.T) {
	s := &Server{}
	server := grpc.NewServer()
	gen.RegisterUserServiceServer(server, s)
	listen, err := net.Listen("tcp", ":8083")
	if err != nil {
		return
	}
	err = server.Serve(listen)
	t.Log(err)
}

type Server struct {
	gen.UnimplementedUserServiceServer
}

func (s *Server) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	log.Println(req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Id:   100,
			Name: "Liu",
		},
	}, nil
}
