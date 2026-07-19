package orders

import (
	"context"
	"errors"
	"simple-worker-in-golang/03-context-aware/fp"
	"testing"
	"time"
)

func TestKitchen_Start(t *testing.T) {
	type args struct {
		order *Order
	}

	type fields struct {
		cooker *Cooker
	}

	tests := []struct {
		name string

		fields fields
		args   args
		cancel bool // cancel ctx instead of / after sending order

		wantErr      error
		wantCooked   bool
		wantedStatus Status
	}{
		{
			// Cancellation is how the kitchen is meant to stop, so it must not
			// look like a failure to the caller.
			name: "should return no error when context is cancelled",
			fields: fields{
				cooker: NewCooker(NewInMemoryOrderRepository()),
			},
			cancel:  true,
			wantErr: nil,
		},
		{
			name: "should cook order received on channel",
			fields: fields{
				cooker: &Cooker{
					recipe:          func(context.Context, Plate) error { return nil }, // skip defaultRecipe's real sleep
					orderRepository: NewInMemoryOrderRepository(),
				},
			},
			args: args{
				order: &Order{Id: 1, Plate: Burger, Table: 1},
			},
			wantCooked:   true,
			wantedStatus: Cooked,
		},
		{
			name: "should return error when order could not be cooked",
			fields: fields{
				cooker: NewCooker(NewInMemoryOrderRepository()),
			},
			args: args{
				order: &Order{Id: 1, Table: 1},
			},
			cancel:     false,
			wantCooked: false,
			wantErr:    ErrPlateNotInMenu,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderCh := make(chan Order)
			cookedCh := make(chan fp.Result[Order])
			k := NewKitchen(tt.fields.cooker, orderCh, cookedCh)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			errCh := make(chan error, 1)
			go func() { errCh <- k.Start(ctx) }()

			if tt.args.order != nil {
				orderCh <- *tt.args.order
			}
			if tt.cancel {
				cancel()
			}

			select {
			case err := <-errCh:
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Start() error = %v, want %v", err, tt.wantErr)
				}
			case result := <-cookedCh:
				if !tt.wantCooked {
					if tt.wantErr == nil {
						t.Fatalf("unexpected cooked result: %+v", result)
					}

					if !result.IsError() {
						t.Errorf("Start() error = %v, want %v", !result.IsError(), tt.wantErr)
					}

					if !errors.Is(result.Error(), tt.wantErr) {
						t.Errorf("Start() error = %v, want %v", result.Error(), tt.wantErr)
					}
				} else {

					if result.IsError() {
						t.Fatalf("unexpected error: %v", result.Error())
					}
					if got := result.Get(); got.Status != tt.wantedStatus {
						t.Errorf("order status = %v, want %v", got.Status, tt.wantedStatus)
					}

				}
			case <-time.After(time.Second):
				t.Fatal("Start() produced no result in time")
			}
		})
	}
}

// TestKitchen_Start_CooksOrdersConcurrently guards against a regression back
// to serial processing: if onOrder is ever called inline again instead of
// fanned out into a goroutine, this test times out.
func TestKitchen_Start_CooksOrdersConcurrently(t *testing.T) {
	const cookTime = 200 * time.Millisecond

	cooker := &Cooker{
		recipe: func(ctx context.Context, _ Plate) error {
			select {
			case <-time.After(cookTime):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
		orderRepository: NewInMemoryOrderRepository(),
	}

	orderCh := make(chan Order)
	cookedCh := make(chan fp.Result[Order])
	k := NewKitchen(cooker, orderCh, cookedCh)

	go k.Start(t.Context())

	start := time.Now()
	orderCh <- Order{Id: 1, Plate: Burger, Table: 1}
	orderCh <- Order{Id: 2, Plate: Salad, Table: 2}

	for range 2 {
		select {
		case result := <-cookedCh:
			if result.IsError() {
				t.Fatalf("unexpected error: %v", result.Error())
			}
		case <-time.After(2 * cookTime):
			t.Fatal("timed out waiting for cooked orders")
		}
	}

	if elapsed := time.Since(start); elapsed >= 2*cookTime {
		t.Errorf("orders were not cooked concurrently: took %v, want < %v", elapsed, 2*cookTime)
	}
}
