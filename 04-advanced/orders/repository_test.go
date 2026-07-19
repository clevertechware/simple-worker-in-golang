package orders

import (
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
)

// =====================================================================
// ======== MOCK FOR OTHER ORDER REPOSITORY CONSUMER INTERFACE =========
// =====================================================================
type orderRepositoryMock struct {
	FindByTableFunc func(table Table) (Order, error)
	SaveFunc        func(o Order) (Order, error)

	// optionnel : pour vérifier les appels
	FindByTableCalls []Table
	SaveCalls        []Order
}

func (m *orderRepositoryMock) FindByTable(table Table) (Order, error) {
	m.FindByTableCalls = append(m.FindByTableCalls, table)
	return m.FindByTableFunc(table)
}

func (m *orderRepositoryMock) Save(o Order) (Order, error) {
	m.SaveCalls = append(m.SaveCalls, o)
	return m.SaveFunc(o)
}

// =====================================================================
// =====================================================================
// =====================================================================

// newTestOrderIdx builds an *atomic.Uint32 preset to v, since atomic.Uint32
// can't be built with a composite literal (unexported fields).
func newTestOrderIdx(v uint32) *atomic.Uint32 {
	var idx atomic.Uint32
	idx.Store(v)
	return &idx
}

func Test_inMemoryOrderRepository_FindByTable(t *testing.T) {
	t.Parallel()

	type fields struct {
		orders   map[uint32]Order
		orderIdx *atomic.Uint32
	}
	type args struct {
		table Table
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Order
		wantErr bool
	}{
		{
			name: "should return the order sitting at the requested table",
			fields: fields{
				orders: map[uint32]Order{
					1: {Id: 1, Plate: Burger, Table: 5},
					2: {Id: 2, Plate: Salad, Table: 7},
				},
				orderIdx: newTestOrderIdx(2),
			},
			args: args{table: 7},
			want: Order{Id: 2, Plate: Salad, Table: 7},
		},
		{
			name: "should return the most recent order when a table has several",
			fields: fields{
				orders: map[uint32]Order{
					1: {Id: 1, Plate: Burger, Table: 5},
					2: {Id: 2, Plate: Salad, Table: 5},
				},
				orderIdx: newTestOrderIdx(2),
			},
			args: args{table: 5},
			want: Order{Id: 2, Plate: Salad, Table: 5},
		},
		{
			name: "should return errNotFound when no order matches the table",
			fields: fields{
				orders: map[uint32]Order{
					1: {Id: 1, Plate: Burger, Table: 5},
				},
				orderIdx: newTestOrderIdx(1),
			},
			args:    args{table: 42},
			want:    Order{},
			wantErr: true,
		},
		{
			name: "should return errNotFound on an empty repository",
			fields: fields{
				orders:   map[uint32]Order{},
				orderIdx: newTestOrderIdx(0),
			},
			args:    args{table: 1},
			want:    Order{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &inMemoryOrderRepository{
				orders:   tt.fields.orders,
				orderIdx: tt.fields.orderIdx,
			}
			got, err := r.FindByTable(tt.args.table)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByTable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !errors.Is(err, errNotFound) {
				t.Errorf("FindByTable() error = %v, want %v", err, errNotFound)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindByTable() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_inMemoryOrderRepository_Save(t *testing.T) {
	t.Parallel()

	type fields struct {
		orders   map[uint32]Order
		orderIdx *atomic.Uint32
	}
	type args struct {
		o Order
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       Order
		wantOrders map[uint32]Order
		wantErr    bool
	}{
		{
			name: "should assign an auto-incremented id when the order has none",
			fields: fields{
				orders:   map[uint32]Order{},
				orderIdx: newTestOrderIdx(0),
			},
			args: args{o: Order{Plate: Burger, Table: 3}},
			want: Order{Id: 1, Plate: Burger, Table: 3},
			wantOrders: map[uint32]Order{
				1: {Id: 1, Plate: Burger, Table: 3},
			},
		},
		{
			name: "should keep the existing id when the order already has one",
			fields: fields{
				orders:   map[uint32]Order{},
				orderIdx: newTestOrderIdx(0),
			},
			args: args{o: Order{Id: 5, Plate: Salad, Table: 2}},
			want: Order{Id: 5, Plate: Salad, Table: 2},
			wantOrders: map[uint32]Order{
				5: {Id: 5, Plate: Salad, Table: 2},
			},
		},
		{
			name: "should continue incrementing from the initial index value",
			fields: fields{
				orders:   map[uint32]Order{},
				orderIdx: newTestOrderIdx(3),
			},
			args: args{o: Order{Plate: Burger}},
			want: Order{Id: 4, Plate: Burger},
			wantOrders: map[uint32]Order{
				4: {Id: 4, Plate: Burger},
			},
		},
		{
			name: "should overwrite an existing order stored under the same id",
			fields: fields{
				orders: map[uint32]Order{
					2: {Id: 2, Plate: Burger, Status: Cooking},
				},
				orderIdx: newTestOrderIdx(2),
			},
			args: args{o: Order{Id: 2, Plate: Salad, Status: Cooked}},
			want: Order{Id: 2, Plate: Salad, Status: Cooked},
			wantOrders: map[uint32]Order{
				2: {Id: 2, Plate: Salad, Status: Cooked},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &inMemoryOrderRepository{
				orders:   tt.fields.orders,
				orderIdx: tt.fields.orderIdx,
			}
			got, err := r.Save(tt.args.o)
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Save() got = %v, want %v", got, tt.want)
			}
			if tt.wantOrders != nil && !reflect.DeepEqual(r.orders, tt.wantOrders) {
				t.Errorf("Save() stored = %v, want %v", r.orders, tt.wantOrders)
			}
		})
	}
}

// Test_inMemoryOrderRepository_Save_Concurrent guards against a regression
// of the map-write race: run with `go test -race` to catch it.
func Test_inMemoryOrderRepository_Save_Concurrent(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryOrderRepository()

	const writers = 20
	var wg sync.WaitGroup
	wg.Add(writers)
	for i := range writers {
		go func(table int) {
			defer wg.Done()
			_, _ = repo.Save(Order{Plate: Burger, Table: Table(table)})
			_, _ = repo.FindByTable(Table(table))
		}(i)
	}
	wg.Wait()
}

func Test_newInMemoryOrderRepository(t *testing.T) {
	t.Parallel()

	// The repository owns its storage: no caller holds a reference to the map,
	// so nothing can mutate it outside the mutex.
	want := &inMemoryOrderRepository{orders: map[uint32]Order{}, orderIdx: newTestOrderIdx(0)}
	if got := NewInMemoryOrderRepository(); !reflect.DeepEqual(got, want) {
		t.Errorf("NewInMemoryOrderRepository() = %v, want %v", got, want)
	}

	// Ids must start at 1: FindByTable treats a zero id as "no match".
	saved, err := NewInMemoryOrderRepository().Save(Order{Plate: Burger, Table: 1})
	if err != nil {
		t.Fatalf("Save() unexpected error = %v", err)
	}
	if saved.Id != 1 {
		t.Errorf("Save() first id = %d, want 1", saved.Id)
	}
}
