package orders

import (
	"errors"
	"reflect"
	"simple-worker-in-golang/03-context-aware/fp"
	"testing"
	"time"
)

func TestWaiter_TakeOrder(t *testing.T) {
	t.Parallel()

	type fields struct {
		orderRepository OrderRepository
	}

	type args struct {
		plate string
		table int
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Order
		wantErr bool
	}{
		{
			name: "should take order",
			fields: fields{
				orderRepository: func() OrderRepository {
					return &orderRepositoryMock{
						SaveFunc: func(o Order) (Order, error) {
							return o, nil
						},
					}
				}(),
			},
			args: args{
				plate: "Burger",
				table: 1,
			},
			want: Order{
				Plate: "Burger",
				Table: 1,
			},
		},
		{
			name: "should return err when plate is not on menu",
			fields: fields{
				orderRepository: func() OrderRepository {
					return &orderRepositoryMock{}
				}(),
			},
			args:    args{plate: "Rice"},
			wantErr: true,
		},
		{
			name: "should return err when order cannot be saved",
			fields: fields{
				orderRepository: func() OrderRepository {
					return &orderRepositoryMock{
						SaveFunc: func(o Order) (Order, error) {
							return Order{}, errors.New("technical error")
						},
					}
				}(),
			},
			args:    args{plate: "Burger"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			order := make(chan Order, 1) // Leave the order channel unbuffered to avoid blocking.
			cooked := make(chan fp.Result[Order])
			w := &Waiter{
				orderRepository: tt.fields.orderRepository,
				order:           order,
				cooked:          cooked,
			}

			if err := w.TakeOrder(t.Context(), tt.args.plate, tt.args.table); (err != nil) != tt.wantErr {
				t.Errorf("TakeOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			select {
			case orderReceived := <-order:
				if !reflect.DeepEqual(orderReceived, tt.want) {
					t.Errorf("TakeOrder() orderReceived = %v, want %v", orderReceived, tt.want)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("timed out")
			}
		})
	}
}
