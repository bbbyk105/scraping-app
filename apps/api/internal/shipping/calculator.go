package shipping

import (
	"math"
)

type Calculator struct {
	config Config
}

type Config struct {
	Mode       string
	FeePercent float64
	FXUSDJPY   float64
}

func NewCalculator(config Config) *Calculator {
	return &Calculator{config: config}
}

// CalculateShipping calculates shipping cost to US based on price amount (in cents)
func (c *Calculator) CalculateShipping(priceAmountCents int) int {
	priceUSD := float64(priceAmountCents) / 100.0

	var shippingUSD float64
	switch c.config.Mode {
	case "TABLE":
		shippingUSD = c.calculateByTable(priceUSD)
	default:
		// Default flat rate
		shippingUSD = 14.99
	}

	// Add fee percentage
	feeAmount := priceUSD * (c.config.FeePercent / 100.0)
	totalShipping := shippingUSD + feeAmount

	// Convert back to cents
	return int(math.Round(totalShipping * 100))
}

func (c *Calculator) calculateByTable(priceUSD float64) float64 {
	if priceUSD < 20.0 {
		return 9.99
	} else if priceUSD < 50.0 {
		return 14.99
	} else {
		return 19.99
	}
}

// CalculateTotal calculates total amount (price + shipping) in cents
func (c *Calculator) CalculateTotal(priceAmountCents int) int {
	shipping := c.CalculateShipping(priceAmountCents)
	return priceAmountCents + shipping
}

// ConvertToJPY converts USD cents to JPY (for display purposes)
func (c *Calculator) ConvertToJPY(usdCents int) int {
	usdAmount := float64(usdCents) / 100.0
	jpyAmount := usdAmount * c.config.FXUSDJPY
	return int(math.Round(jpyAmount))
}

