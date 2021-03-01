package slicer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shopspring/decimal"
)

type Cheeses map[string]decimal.Decimal

func (c Cheeses) GetPrice(product string) decimal.Decimal { return c[product] }

var catalogue = Cheeses{
	"Pilgrims Choice": decimal.RequireFromString("5.60"),
	"Nantwich":        decimal.RequireFromString("4.99"),
	"Cheddar":         decimal.RequireFromString("1.00"),
	"Camembert":       decimal.RequireFromString("7.20"),
	"Edam":            decimal.RequireFromString("4.89"),
	"Port Salut":      decimal.RequireFromString("2.99"),
	"Brie (Small)":    decimal.RequireFromString("1.30"),
	"Brie (Medium)":   decimal.RequireFromString("2.89"),
	"Brie (Large)":    decimal.RequireFromString("3.50"),
}

type SimpleBasket struct {
	Basket

	items map[string]int
}

func (c *SimpleBasket) GetProducts() []string {
	products := make([]string, 0, len(c.items))

	for product := range c.items {
		products = append(products, product)
	}

	return products
}

func (c *SimpleBasket) GetQuantity(cheese string) int {
	return c.items[cheese]
}

func TestSlicer_Calc(t *testing.T) {

	slicer := NewSlicer(catalogue.GetPrice, []Offer{
		PercentageDiscount([]string{"Edam"}, 25),
		BuyNGetCheapestFree([]string{"Cheddar"}, 3),
		BuyNGetCheapestFree([]string{"Brie (Large)", "Brie (Medium)", "Brie (Small)"}, 3),
	})

	for _, test := range []struct {
		name      string
		basket    Basket
		expected  Total
		sayCheese bool
	}{
		{
			name:     "empty basket",
			basket:   &SimpleBasket{items: map[string]int{}},
			expected: Total{SubTotal: decimal.Zero, Discount: decimal.Zero, Total: decimal.Zero},
		},
		{
			name: "no promotions",
			basket: &SimpleBasket{items: map[string]int{
				"Camembert":  2,
				"Port Salut": 1}},
			expected: Total{SubTotal: decimal.RequireFromString("17.39"), Discount: decimal.Zero, Total: decimal.RequireFromString("17.39")},
		},
		{
			name: "buy_two_get_one_free(1 bought)",
			basket: &SimpleBasket{items: map[string]int{
				"Cheddar": 1}},
			expected: Total{SubTotal: decimal.RequireFromString("1.00"), Discount: decimal.Zero, Total: decimal.RequireFromString("1.00")},
		},
		{
			name: "buy_two_get_one_free(2 bought)",
			basket: &SimpleBasket{items: map[string]int{
				"Cheddar": 2}},
			expected: Total{SubTotal: decimal.RequireFromString("2.00"), Discount: decimal.Zero, Total: decimal.RequireFromString("2.00")},
		},
		{
			name: "buy_two_get_one_free(3 bought)",
			basket: &SimpleBasket{items: map[string]int{
				"Cheddar": 3}},
			expected: Total{SubTotal: decimal.RequireFromString("3.00"), Discount: decimal.RequireFromString("1.00"), Total: decimal.RequireFromString("2.00")},
		},
		{
			name: "buy_two_get_one_free(7 bought)",
			basket: &SimpleBasket{items: map[string]int{
				"Cheddar": 7}},
			expected: Total{SubTotal: decimal.RequireFromString("7.00"), Discount: decimal.RequireFromString("2.00"), Total: decimal.RequireFromString("5.00")},
		},
		{
			name: "25% off Edam",
			basket: &SimpleBasket{items: map[string]int{
				"Pilgrims Choice": 1,
				"Edam":            1}},
			expected: Total{SubTotal: decimal.RequireFromString("10.49"), Discount: decimal.RequireFromString("1.22"), Total: decimal.RequireFromString("9.27")},
		},
		{
			name: "buy_three_of_set_get_cheapest_free (negative test)",
			basket: &SimpleBasket{items: map[string]int{
				"Pilgrims Choice": 1,
				"Edam":            1}},
			expected: Total{SubTotal: decimal.RequireFromString("10.49"), Discount: decimal.RequireFromString("1.22"), Total: decimal.RequireFromString("9.27")},
		},
		{
			name: "buy_three_of_set_get_cheapest_free (1 free)",
			basket: &SimpleBasket{items: map[string]int{
				"Brie (Large)":  1,
				"Brie (Medium)": 1,
				"Brie (Small)":  1}},
			expected:  Total{SubTotal: decimal.RequireFromString("7.69"), Discount: decimal.RequireFromString("1.30"), Total: decimal.RequireFromString("6.39")},
			sayCheese: true,
		},
		{
			name: "buy_three_of_set_get_cheapest_free (2 free)",
			basket: &SimpleBasket{items: map[string]int{
				"Brie (Large)":  2,
				"Brie (Medium)": 2,
				"Brie (Small)":  2}},
			expected:  Total{SubTotal: decimal.RequireFromString("15.38"), Discount: decimal.RequireFromString("2.60"), Total: decimal.RequireFromString("12.78")},
			sayCheese: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			basketTotal := slicer.Calc(test.basket)
			assert.Equal(t, test.expected.SubTotal.StringFixed(2), basketTotal.SubTotal.StringFixed(2))
			assert.Equal(t, test.expected.Discount.StringFixed(2), basketTotal.Discount.StringFixed(2))
			assert.Equal(t, test.expected.Total.StringFixed(2), basketTotal.Total.StringFixed(2))

			if test.sayCheese {
				fmt.Printf("basket contents....\n")
				for _, v := range test.basket.GetProducts() {
					fmt.Printf("%d pack(s) of %s at £ %s each \n", test.basket.GetQuantity(v), v, catalogue.GetPrice(v).StringFixed(2))
				}
				fmt.Printf("discount of £ %s was applied. \n", basketTotal.Discount.StringFixed(2))

			}
		})
	}

}
