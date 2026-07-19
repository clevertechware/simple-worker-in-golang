package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"simple-worker-in-golang/04-advanced/fp"
	"simple-worker-in-golang/04-advanced/orders"
	"sync"
	"syscall"
	"time"
)

func main() {
	orderRepository := orders.NewInMemoryOrderRepository()
	cooker := orders.NewCooker(orderRepository)
	orderChan := make(chan orders.Order)
	cookedChan := make(chan fp.Result[orders.Order])
	kitchen := orders.NewKitchen(cooker, orderChan, cookedChan)
	waiter := orders.NewWaiter(orderRepository, orderChan, cookedChan)

	// Timeout is sized above the slowest recipe (Burger, 10s) so the single
	// demo order has time to be cooked and served before shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Create channel to receive interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// Goroutine to handle signals. Cancelling the context unwinds the kitchen,
	// the waiter and any customer still waiting to place an order, so main
	// falls through to wg.Wait() and returns on its own.
	go func() {
		sig := <-sigChan
		log.Printf("received %s, shutting down", sig)
		cancel()
	}()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := kitchen.Start(ctx)
		if err != nil {
			log.Printf("kitchen error: %v", err)
			cancel()
		}
	}()
	go func() {
		defer wg.Done()
		waiter.Serve(ctx)
	}()

	customers := []struct {
		plate orders.Plate
		table int
	}{
		{orders.Burger, 1},
		{orders.Salad, 2},
		{orders.Burger, 3},
	}

	var customersWg sync.WaitGroup
	customersWg.Add(len(customers))
	for _, c := range customers {
		go func(plate orders.Plate, table int) {
			defer customersWg.Done()
			if err := waiter.TakeOrder(ctx, string(plate), table); err != nil {
				log.Printf("could not take order: %v", err)
				cancel()
			}
		}(c.plate, c.table)
	}
	customersWg.Wait()

	wg.Wait()
}
