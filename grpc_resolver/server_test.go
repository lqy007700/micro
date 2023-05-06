package grpc_resolver

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"micro"
	"micro/proto/gen"
	"micro/registry/etcd"
	"testing"
)

func TestServer(t *testing.T) {
	// 启动etcd
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	if err != nil {
		t.Log(err)
		return
	}

	// etcd注册
	registry, err := etcd.NewRegistry(client)
	if err != nil {
		t.Log(err)
		return
	}

	// 创建服务
	server, err := micro.NewServer("user-service", micro.ServerWithRegistry(registry))
	if err != nil {
		t.Log(err)
		return
	}

	// 注册服务
	gen.RegisterUserServiceServer(server, &UserServiceServer{})

	// 启动
	err = server.Start(context.Background(), ":8083")
	t.Log(err)
}

type UserServiceServer struct {
	gen.UnimplementedUserServiceServer
}

func (s *UserServiceServer) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	log.Println(req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Id:   100,
			Name: "Liu",
		},
	}, nil
}
