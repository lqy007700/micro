package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"micro/registry"
	"sync"
)

type Registry struct {
	c       *clientv3.Client
	session *concurrency.Session
	cancels []func()
	mutex   sync.Mutex
}

func NewRegistry(c *clientv3.Client) (*Registry, error) {
	session, err := concurrency.NewSession(c)
	if err != nil {
		return nil, err
	}
	return &Registry{
		c:       c,
		session: session,
	}, nil
}

func (r *Registry) Register(ctx context.Context, si registry.ServiceInstance) error {
	val, err := json.Marshal(si)
	if err != nil {
		return err
	}

	_, err = r.c.Put(ctx, r.instanceKey(si), string(val), clientv3.WithLease(r.session.Lease()))
	return err
}

func (r *Registry) UnRegister(ctx context.Context, si registry.ServiceInstance) error {
	_, err := r.c.Delete(ctx, r.instanceKey(si))
	return err
}

func (r *Registry) ListServices(ctx context.Context, name string) ([]registry.ServiceInstance, error) {
	resp, err := r.c.Get(context.Background(), r.serviceKey(name), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	res := make([]registry.ServiceInstance, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var si registry.ServiceInstance
		err = json.Unmarshal(kv.Value, &si)
		if err != nil {
			return nil, err
		}
		res = append(res, si)
	}
	return res, nil
}

func (r *Registry) Subscribe(name string) (<-chan registry.Event, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	ctx = clientv3.WithRequireLeader(ctx)
	r.mutex.Lock()
	r.cancels = append(r.cancels, cancelFunc)
	r.mutex.Unlock()

	watchEvents := r.c.Watch(ctx, r.serviceKey(name), clientv3.WithPrefix())
	res := make(chan registry.Event)
	go func() {
		for {
			select {
			case watch := <-watchEvents:
				if watch.Err() != nil {
					continue
				}
				if watch.Canceled {
					return
				}
				res <- registry.Event{}
			case <-ctx.Done():
				return
			}
		}
	}()
	return res, nil
}

func (r *Registry) Close() error {
	r.mutex.Lock()
	cancels := r.cancels
	r.cancels = nil
	r.mutex.Unlock()

	err := r.session.Close()
	for _, cancel := range cancels {
		cancel()
	}
	return err
}

func (r *Registry) instanceKey(si registry.ServiceInstance) string {
	return fmt.Sprintf("/micro/%s/%s", si.Name, si.Address)
}

func (r *Registry) serviceKey(name string) string {
	return fmt.Sprintf("/micro/%s", name)
}
