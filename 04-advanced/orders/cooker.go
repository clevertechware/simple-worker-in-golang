package orders

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Cooker struct {
	recipe func(ctx context.Context, plate Plate) error

	orderRepository OrderRepository
}

func NewCooker(orderRepository OrderRepository) *Cooker {
	return &Cooker{recipe: defaultRecipe, orderRepository: orderRepository}
}

func (c *Cooker) Cook(ctx context.Context, order Order) (Order, error) {
	log.Printf("received order %d", order.Id)
	if err := order.Plate.Validate(); err != nil {
		return Order{}, fmt.Errorf("invalid order: %w", err)
	}

	order.MarkAsCooking()
	order, err := c.orderRepository.Save(order)
	if err != nil {
		return Order{}, fmt.Errorf("could not save order: %w", err)
	}

	plate := order.Plate
	err = c.recipe(ctx, plate)
	if err != nil {
		// Without this the order stays Cooking forever, since this is the last
		// thing written for it.
		order.MarkAsCancelled()
		if _, saveErr := c.orderRepository.Save(order); saveErr != nil {
			log.Printf("order %d: could not save cancelled order: %v", order.Id, saveErr)
		}
		return Order{}, fmt.Errorf("could apply recipe on order plate: %w", err)
	}

	order.MarkAsCooked()
	order, err = c.orderRepository.Save(order)
	if err != nil {
		return Order{}, fmt.Errorf("could not save order: %w", err)
	}
	log.Printf("order %d cooked: %s", order.Id, plate)
	return order, nil
}

func defaultRecipe(ctx context.Context, plate Plate) error {
	var cookTime time.Duration
	switch plate {
	// simulating long cooking process
	case Burger:
		cookTime = 10 * time.Second
	case Salad:
		cookTime = 5 * time.Second
	default:
		return ErrPlateNotInMenu
	}

	select {
	case <-time.After(cookTime):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
