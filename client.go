package micro

import (
	"fmt"
	"google.golang.org/grpc"
	"micro/registry"
)

type ClientOpt func(c *Client)

type Client struct {
	r        registry.Registry
	insecure bool
}

func NewClient(opts ...ClientOpt) (*Client, error) {
	res := &Client{}

	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func WithClientInsecure() ClientOpt {
	return func(c *Client) {
		c.insecure = true
	}
}

func WithClientRegistry(r registry.Registry) ClientOpt {
	return func(c *Client) {
		c.r = r
	}
}

func (c *Client) Dial(name string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	if c.r != nil {
		rb := NewRegistryBuilder(c.r)
		opts = append(opts, grpc.WithResolvers(rb))
	}
	if c.insecure {
		opts = append(opts, grpc.WithInsecure())
	}

	opts = append(opts, grpc.WithDefaultServiceConfig(`{"LoadBalancingPolicy":"DEMO_ROUND_ROBIN"}`))

	cc, err := grpc.Dial(fmt.Sprintf("registry:///%s", name), opts...)
	return cc, err
}
