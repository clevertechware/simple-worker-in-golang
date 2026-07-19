package orders

import (
	"context"
	"fmt"
	"log"
	"simple-worker-in-golang/04-advanced/fp"
)

type Kitchen struct {
	cooker *Cooker

	order  <-chan Order
	cooked chan<- fp.Result[Order]
}

func NewKitchen(cooker *Cooker, order <-chan Order, cooked chan<- fp.Result[Order]) *Kitchen {
	return &Kitchen{
		cooker: cooker,
		order:  order,
		cooked: cooked,
	}
}

func (k *Kitchen) Start(ctx context.Context) error {
	for {
		select {
		case order := <-k.order:
			go k.onOrder(ctx, order)
		case <-ctx.Done():
			// Expected shutdown, not a failure: the caller distinguishes a
			// broken kitchen from a cancelled one by this being nil.
			return nil
		}
	}
}

func (k *Kitchen) onOrder(ctx context.Context, order Order) {
	orderCooked, err := k.cooker.Cook(ctx, order)

	result := fp.NewValue(orderCooked)
	if err != nil {
		result = fp.NewError[Order](fmt.Errorf("order %d failed: %w", order.Id, err))
	}

	select {
	case k.cooked <- result:
	case <-ctx.Done():
		log.Printf("order %d cooked but not delivered: %v", order.Id, ctx.Err())
	}
}
