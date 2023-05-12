/**
平滑加权轮询
不推荐，需要动态维护机器的权重值
需要加锁更新
*/

package roundrobin

import (
	"errors"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math"
	"sync"
	"sync/atomic"
)

type WeightBalance struct {
	connections []*conn
}

func (w *WeightBalance) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(w.connections) == 0 {
		return balancer.PickResult{}, errors.New("connection is null")
	}

	var totalWeight uint32
	var res *conn
	for _, c := range w.connections {
		c.mutex.Lock()
		totalWeight += c.efficientWeight
		c.currentWeight += c.efficientWeight
		if res == nil || res.currentWeight < c.currentWeight {
			res = c
		}
		c.mutex.Unlock()
	}

	res.mutex.Lock()
	res.currentWeight -= totalWeight
	res.mutex.Unlock()

	return balancer.PickResult{
		SubConn: res.c,
		Done: func(info balancer.DoneInfo) {
			for {
				weight := atomic.LoadUint32(&res.efficientWeight)

				if info.Err != nil && weight == 0 {
					return
				}

				if info.Err == nil && weight == math.MaxUint32 {
					return
				}

				newWeight := weight
				if info.Err != nil {
					newWeight--
				} else {
					newWeight++
				}

				if atomic.CompareAndSwapUint32(&res.efficientWeight, weight, newWeight) {
					return
				}
			}
		},
	}, nil
}

type WeightBalanceBuilder struct {
}

func (w *WeightBalanceBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	cs := make([]*conn, 0, len(info.ReadySCs))
	for subConn, connInfo := range info.ReadySCs {
		weight, ok := connInfo.Address.Attributes.Value("weight").(uint32)
		if !ok || weight == 0 {
			panic("weight empty")
		}

		cs = append(cs, &conn{
			c:               subConn,
			weight:          weight,
			currentWeight:   weight,
			efficientWeight: weight,
		})
	}
	return &WeightBalance{}
}

type conn struct {
	c               balancer.SubConn
	weight          uint32 // 权重
	currentWeight   uint32 // 当前权重
	efficientWeight uint32 // 有效权重
	mutex           sync.Mutex
}
