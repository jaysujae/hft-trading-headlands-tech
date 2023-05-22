package main

import (
	"fmt"
	"github.com/alexey-ernest/go-hft-orderbook"
)

type Calculator interface {
	Calculate() float64
}

type bookPressureCalculator struct {
	ob *hftorderbook.Orderbook
}

func (b bookPressureCalculator) Calculate() float64 {
	bestOfferPrice := b.ob.GetBestOffer()
	bestOfferVolume := b.ob.GetVolumeAtAskLimit(bestOfferPrice)
	bestBidPrice := b.ob.GetBestBid()
	bestBidVolume := b.ob.GetVolumeAtBidLimit(bestBidPrice)
	return (bestOfferPrice*bestBidVolume + bestBidPrice*bestOfferVolume) / (bestBidVolume + bestOfferVolume)
}

type transaction struct {
	gap    float64
	volume float64
}

type tradeImpulseCalculator struct {
	averageVolume float64
	transaction   transaction
}

func (t tradeImpulseCalculator) Calculate() float64 {
	return t.transaction.gap * t.transaction.volume / t.averageVolume
}

type compositeCalculator struct {
	calculators []Calculator
}

func (c compositeCalculator) Calculate() float64 {
	price := float64(0)
	for _, calculator := range c.calculators {
		price += calculator.Calculate()
	}
	return price
}

type Strategy interface {
	Execute()
}

type CalculatorStrategy struct {
	calculator Calculator
	ob         *hftorderbook.Orderbook
}

func (c CalculatorStrategy) Execute() {
	calculatedPrice := c.calculator.Calculate()
	profit := float64(0)
	if calculatedPrice < c.ob.GetBestBid() {
		profit += (c.ob.GetBestBid() - calculatedPrice) * c.ob.GetVolumeAtBidLimit(c.ob.GetBestBid())
	}
	if calculatedPrice > c.ob.GetBestOffer() {
		profit += (c.ob.GetBestOffer() - calculatedPrice) * c.ob.GetVolumeAtBidLimit(c.ob.GetBestOffer())
	}
	fmt.Printf("calculated price : %f\n", calculatedPrice)
	fmt.Printf("profit : %f\n", profit)
}

func main() {
	ob := hftorderbook.NewOrderbook()
	ask1 := &hftorderbook.Order{
		Id:       1,
		BidOrAsk: false,
		Volume:   5,
	}
	bid1 := &hftorderbook.Order{
		Id:       2,
		BidOrAsk: true,
		Volume:   1,
	}
	ob.Add(99.0, ask1)
	ob.Add(98.75, bid1)

	strategy1 := &CalculatorStrategy{
		calculator: &compositeCalculator{
			calculators: []Calculator{
				&bookPressureCalculator{
					ob: &ob,
				},
				&tradeImpulseCalculator{
					averageVolume: 15,
					transaction: transaction{
						gap:    -0.25,
						volume: 9,
					},
				},
			},
		},
		ob: &ob,
	}
	strategy1.Execute()

	return
}
