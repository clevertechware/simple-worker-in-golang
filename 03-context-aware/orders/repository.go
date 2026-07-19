package orders

import (
	"sync"
	"sync/atomic"
)

type OrderRepository interface {
	FindByTable(table Table) (Order, error)
	Save(o Order) (Order, error)
}

type inMemoryOrderRepository struct {
	mu     sync.RWMutex
	orders map[uint32]Order

	orderIdx *atomic.Uint32
}

func NewInMemoryOrderRepository() OrderRepository {
	return &inMemoryOrderRepository{orders: make(map[uint32]Order), orderIdx: &atomic.Uint32{}}
}

// FindByTable returns the most recent order for the table. A table can hold
// several orders over time, and map iteration order is randomised, so picking
// the highest id keeps the result stable instead of arbitrary.
func (r *inMemoryOrderRepository) FindByTable(table Table) (Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	found := Order{}
	for _, o := range r.orders {
		if table == o.Table && o.Id >= found.Id {
			found = o
		}
	}
	if found.Id == 0 {
		return Order{}, errNotFound
	}
	return found, nil
}

func (r *inMemoryOrderRepository) Save(o Order) (Order, error) {
	if o.Id == 0 {
		o.Id = r.orderIdx.Add(1)
	}

	r.mu.Lock()
	r.orders[o.Id] = o
	r.mu.Unlock()

	return o, nil
}
