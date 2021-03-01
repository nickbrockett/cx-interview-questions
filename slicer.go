package slicer

import (
	"sort"

	"github.com/shopspring/decimal"
)

type GetPrice func(product string) decimal.Decimal

// Basket represents basic Basket interface.
type Basket interface {
	GetProducts() []string    // returns slice of all products in the basket
	GetQuantity(s string) int // returns quantity of given product in the basket
}

type Slicer struct {
	getPrice GetPrice
	offers   []Offer
}

// Offer returns calculated discount for given Basket.
type Offer func(Basket, GetPrice) (discount decimal.Decimal)

// Total contains Basket's totals.
type Total struct {
	SubTotal, Discount, Total decimal.Decimal
}

// NewSlicer creates new Slicer with given pricing (GetPrice) and list of offers.
func NewSlicer(getPrice GetPrice, offers []Offer) Slicer {
	return Slicer{getPrice: getPrice, offers: offers}
}

// Calc returns Basket's Total.
func (s Slicer) Calc(basket Basket) Total {

	basketTotal := Total{SubTotal: decimal.Zero, Discount: decimal.Zero, Total: decimal.Zero}

	for _, product := range basket.GetProducts() {
		productTotal := decimal.Zero
		quantity := decimal.NewFromInt(int64(basket.GetQuantity(product)))
		productTotal = s.getPrice(product).Mul(quantity)
		basketTotal.SubTotal = basketTotal.SubTotal.Add(productTotal)
		basketTotal.Total = basketTotal.Total.Add(productTotal)
	}

	for _, offer := range s.offers {
		basketTotal.Discount = basketTotal.Discount.Add(offer(basket, s.getPrice))
	}

	basketTotal.Total = basketTotal.Total.Sub(basketTotal.Discount)

	return basketTotal
}

// PrecentageDiscount returns Offer which calculates percentage discount for given Product slice.
func PercentageDiscount(products []string, percentage int) Offer {

	// sort products to use binary search
	sort.Strings(products)

	return func(basket Basket, getPrice GetPrice) decimal.Decimal {

		discountTotal := decimal.Zero
		discountPercentage := decimal.NewFromInt(int64(percentage))
		if discountPercentage.IsZero() {
			return discountTotal
		}

		for _, product := range basket.GetProducts() {
			i := sort.SearchStrings(products, product)
			if i < len(products) && products[i] == product {
				quantity := decimal.NewFromInt(int64(basket.GetQuantity(product)))
				if quantity.IsZero() {
					continue
				}
				cost := getPrice(product).Mul(quantity)
				discount := cost.Mul(discountPercentage).Div(decimal.New(100, 0)).Round(2)
				discountTotal = discountTotal.Add(discount)
			}
		}
		return discountTotal
	}
}

// BuyNGetCheapestFree returns Offer that calculates discount for "buy N, get the cheapest free" (also "buy N, get one free").
func BuyNGetCheapestFree(products []string, n int) Offer {

	// sort products to use binary search
	sort.Strings(products)

	return func(basket Basket, getPrice GetPrice) decimal.Decimal {

		var qtyInSet, qtyToDiscount, remainder int

		var matched []struct {
			product  string
			price    decimal.Decimal
			quantity int
		}

		for _, basketItem := range basket.GetProducts() {
			i := sort.SearchStrings(products, basketItem)
			if i < len(products) && products[i] == basketItem {
				matched = append(matched, struct {
					product  string
					price    decimal.Decimal
					quantity int
				}{basketItem, getPrice(basketItem), basket.GetQuantity(basketItem)})
				qtyInSet += basket.GetQuantity(basketItem)
			}
		}

		// sort matched products by price - less expensive first, and use every n item as discount
		sort.Slice(matched, func(i, j int) bool { return matched[i].price.Cmp(matched[j].price) < 0 })

		remainder = (qtyInSet) % n
		qtyToDiscount = (qtyInSet - remainder) / n
		discountTotal := decimal.Zero

		if qtyToDiscount > 0 {
			for _, match := range matched {

				if qtyToDiscount > match.quantity {
					discountTotal = discountTotal.Add(match.price.Mul(decimal.New(int64(match.quantity), 0)))
					qtyToDiscount -= match.quantity
				} else {
					discountTotal = discountTotal.Add(match.price.Mul(decimal.New(int64(qtyToDiscount), 0)))
					break
				}

			}
		}

		return discountTotal
	}
}
