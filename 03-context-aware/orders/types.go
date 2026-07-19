package orders

import (
	"fmt"
	"slices"
)

const (
	Burger Plate = "Burger"
	Salad  Plate = "Salad"
)

var (
	ErrPlateNotInMenu = fmt.Errorf("plate not in menu")

	Menu = []Plate{Burger, Salad}
)

type Table int

type Status int

const (
	Ordered Status = iota
	Cooking
	Cooked
	Served
	Cancelled
)

type Order struct {
	Id     uint32
	Plate  Plate
	Status Status
	Table  Table
}

func (o *Order) MarkAsCooking() {
	o.Status = Cooking
}

func (o *Order) MarkAsCooked() {
	o.Status = Cooked
}

func (o *Order) MarkAsServed() {
	o.Status = Served
}

func (o *Order) MarkAsCancelled() {
	o.Status = Cancelled
}

type Plate string

func (p Plate) Validate() error {
	if !slices.Contains(Menu, p) {
		return ErrPlateNotInMenu
	}
	return nil
}

func (p Plate) String() string {
	return string(p)
}
