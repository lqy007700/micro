package grpc_resolver

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"micro"
	"micro/proto/gen"
	"micro/registry/etcd"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	// 启动etcd
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 3,
	})
	t.Log(client)
	if err != nil {
		t.Log(err)
		return
	}

	// 初始化注册中心实例
	r, err := etcd.NewRegistry(client)
	if err != nil {
		t.Log(err)
		return
	}
	cc, _ := micro.NewClient(micro.WithClientInsecure(), micro.WithClientRegistry(r))
	grpcCC, err := cc.Dial("user-service")
	if err != nil {
		return
	}

	userServiceClient := gen.NewUserServiceClient(grpcCC)
	res, err := userServiceClient.GetById(context.Background(), &gen.GetByIdReq{
		Id: 123,
	})
	if err != nil {
		return
	}
	t.Log(res)
}
