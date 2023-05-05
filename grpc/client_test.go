package grpc

import (
	"context"
	"google.golang.org/grpc"
	"micro/proto/gen"
	"testing"
)

func TestClient(t *testing.T) {
	cc, _ := grpc.Dial("127.0.0.1:8083", grpc.WithInsecure())

	client := gen.NewUserServiceClient(cc)
	res, err := client.GetById(context.Background(), &gen.GetByIdReq{
		Id: 123,
	})
	if err != nil {
		return
	}
	t.Log(res)
}
