package orders

import (
	"context"
	"fmt"
	"log"
	"simple-worker-in-golang/03-context-aware/fp"
)

type Waiter struct {
	orderRepository OrderRepository

	order  chan<- Order
	cooked <-chan fp.Result[Order]
}

func NewWaiter(orderRepository OrderRepository, order chan<- Order, cooked <-chan fp.Result[Order]) *Waiter {
	return &Waiter{orderRepository: orderRepository, order: order, cooked: cooked}
}

func (w *Waiter) TakeOrder(ctx context.Context, plate string, table int) error {
	p := Plate(plate)
	if err := p.Validate(); err != nil {
		return fmt.Errorf("invalid order: %w", err)
	}
	order, err := w.orderRepository.Save(Order{Plate: p, Table: Table(table), Status: Ordered})
	if err != nil {
		return fmt.Errorf("could not take order: %w", err)
	}
	select {
	case w.order <- order:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("could not take order %d: %w", order.Id, ctx.Err())
	}
}

func (w *Waiter) Serve(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case cookedResult := <-w.cooked:
			if cookedResult.IsError() {
				log.Printf("waiter: could not serve order: %v", cookedResult.Error())
				continue
			}

			order := cookedResult.Get()
			order.MarkAsServed()
			if _, err := w.orderRepository.Save(order); err != nil {
				log.Printf("waiter: could not save served order %d: %v", order.Id, err)
				continue
			}
			log.Printf("waiter: delivering %s to table %d", order.Plate, order.Table)
		}
	}
}
