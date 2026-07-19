package orders

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestCooker_Cook(t *testing.T) {
	t.Parallel()

	type fields struct {
		recipe          func(ctx context.Context, plate Plate) error
		orderRepository func() *orderRepositoryMock
	}

	type args struct {
		order Order
	}

	tests := []struct {
		name       string
		fields     fields
		args       args
		setupMocks func(orderRepository *orderRepositoryMock)
		want       Order
		check      func(orderRepository *orderRepositoryMock) bool
		wantErr    bool
	}{
		{
			name: "should return error on invalid order plate",
			fields: fields{
				recipe: func(ctx context.Context, plate Plate) error {
					return nil
				},
				orderRepository: func() *orderRepositoryMock {
					return nil
				},
			},
			args: args{
				order: Order{},
			},
			wantErr: true,
		},
		{
			name: "should cook the order",
			args: args{
				order: Order{Plate: Burger},
			},
			fields: fields{
				recipe: func(ctx context.Context, plate Plate) error {
					return nil
				},
				orderRepository: func() *orderRepositoryMock {
					m := &orderRepositoryMock{}
					m.SaveFunc = func(o Order) (Order, error) {
						return o, nil
					}
					return m
				},
			},
			want: Order{Plate: Burger, Status: Cooked},
			check: func(orderRepository *orderRepositoryMock) bool {
				// check the order was first being marked as cooking
				call := orderRepository.SaveCalls[0]
				if call.Status != Cooking {
					return false
				}

				// then cooked
				call = orderRepository.SaveCalls[1]
				if call.Status != Cooked {
					return false
				}
				return true
			},
		},
		{
			name: "should return error on plate recipe problem",
			args: args{
				order: Order{Plate: Burger},
			},
			fields: fields{
				recipe: func(ctx context.Context, plate Plate) error {
					return errors.New("oops")
				},
				orderRepository: func() *orderRepositoryMock {
					m := &orderRepositoryMock{}
					m.SaveFunc = func(o Order) (Order, error) {
						return o, nil
					}
					return m
				},
			},
			wantErr: true,
		},
		{
			name: "should return error on order repository failure",
			args: args{
				order: Order{Plate: Burger},
			},
			fields: fields{
				recipe: func(ctx context.Context, plate Plate) error {
					return nil
				},
				orderRepository: func() *orderRepositoryMock {
					m := &orderRepositoryMock{}
					m.SaveFunc = func(o Order) (Order, error) {
						return o, errors.New("oops")
					}
					return m
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			orderRepository := tt.fields.orderRepository()
			c := &Cooker{
				recipe:          tt.fields.recipe,
				orderRepository: orderRepository,
			}

			got, err := c.Cook(t.Context(), tt.args.order)
			if (err != nil) != tt.wantErr {
				t.Errorf("Cook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cook() got = %v, want %v", got, tt.want)
			}
			if tt.check != nil && !tt.check(orderRepository) {
				t.Errorf("Calls not expected")
			}
		})
	}
}

func TestCooker_Cook_persistsCancelledOnRecipeFailure(t *testing.T) {
	t.Parallel()

	// A failed recipe must leave a terminal status behind: otherwise the order
	// keeps the Cooking status written just before the recipe ran, forever.
	repo := NewInMemoryOrderRepository()
	c := &Cooker{
		recipe:          func(context.Context, Plate) error { return ErrPlateNotInMenu },
		orderRepository: repo,
	}

	if _, err := c.Cook(t.Context(), Order{Plate: Burger, Table: 3}); err == nil {
		t.Fatal("Cook() error = nil, want an error")
	}

	got, err := repo.FindByTable(3)
	if err != nil {
		t.Fatalf("FindByTable() unexpected error = %v", err)
	}
	if got.Status != Cancelled {
		t.Errorf("order status = %v, want Cancelled (%v)", got.Status, Cancelled)
	}
}
