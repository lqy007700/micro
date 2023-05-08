package roundrobin

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"log"
	"micro"
	"micro/proto/gen"
	"micro/registry/etcd"
	"testing"
	"time"
)

func TestBalancer_Pick(t *testing.T) {
	// 启动etcd
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 3,
	})

	if err != nil {
		t.Log(err)
		return
	}

	r, err := etcd.NewRegistry(client)
	if err != nil {
		t.Log(err)
		return
	}

	go func() {
		// 创建服务
		server, err := micro.NewServer("user-service", micro.ServerWithRegistry(r))
		if err != nil {
			t.Error(err)
			return
		}

		// 注册服务
		gen.RegisterUserServiceServer(server, &UserServiceServer{})

		// 启动
		err = server.Start(context.Background(), ":8083")
		t.Error(err)
	}()

	time.Sleep(time.Second * 3)

	balancer.Register(base.NewBalancerBuilder("DEMO_ROUND_ROBIN", &Builder{}, base.Config{HealthCheck: true}))

	cc, _ := micro.NewClient(micro.WithClientInsecure(), micro.WithClientRegistry(r))
	grpcCC, err := cc.Dial("user-service")
	if err != nil {
		t.Error(err)
		return
	}

	userServiceClient := gen.NewUserServiceClient(grpcCC)
	res, err := userServiceClient.GetById(context.Background(), &gen.GetByIdReq{
		Id: 123,
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(res)
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

func TestBalancer_Pick1(t *testing.T) {
	tests := []struct {
		name    string
		b       *Balancer
		want    SubConn
		wantIdx int32
		wantErr bool
	}{
		{
			name: "start",
			b: &Balancer{
				idx: 0,
				connections: []balancer.SubConn{
					SubConn{name: "8001"},
					SubConn{name: "8002"},
				},
				length: 2,
			},
			want:    SubConn{name: "8001"},
			wantIdx: 1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pick, err := tt.b.Pick(balancer.PickInfo{})
			if err != nil {
				return
			}
			t.Log(pick.SubConn.(SubConn).name)
		})
	}
}

type SubConn struct {
	name string
	balancer.SubConn
}
