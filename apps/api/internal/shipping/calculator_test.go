package shipping

import "testing"

func TestCalculateShipping(t *testing.T) {
	calc := NewCalculator(Config{
		Mode:       "TABLE",
		FeePercent: 3.0,
		FXUSDJPY:   150.0,
	})

	tests := []struct {
		name           string
		priceCents     int
		expectedMin    int
		expectedMax    int
	}{
		{
			name:        "Low price (< $20)",
			priceCents:  1999, // $19.99
			expectedMin: 1050, // $9.99 base + $0.60 fee (3% of $19.99) = $10.59 = 1059 cents
			expectedMax: 1060,
		},
		{
			name:        "Mid price ($20-$50)",
			priceCents:  3999, // $39.99
			expectedMin: 1600, // ~$14.99 base + ~$1.20 fee = ~$16.19
			expectedMax: 1700,
		},
		{
			name:        "High price (> $50)",
			priceCents:  9999, // $99.99
			expectedMin: 2200, // ~$19.99 base + ~$3.00 fee = ~$22.99
			expectedMax: 2400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateShipping(tt.priceCents)
			if result < tt.expectedMin || result > tt.expectedMax {
				t.Errorf("CalculateShipping(%d) = %d, want between %d and %d",
					tt.priceCents, result, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestCalculateTotal(t *testing.T) {
	calc := NewCalculator(Config{
		Mode:       "TABLE",
		FeePercent: 3.0,
		FXUSDJPY:   150.0,
	})

	priceCents := 4999 // $49.99
	shipping := calc.CalculateShipping(priceCents)
	total := calc.CalculateTotal(priceCents)

	expectedTotal := priceCents + shipping
	if total != expectedTotal {
		t.Errorf("CalculateTotal(%d) = %d, want %d", priceCents, total, expectedTotal)
	}
}

func TestConvertToJPY(t *testing.T) {
	calc := NewCalculator(Config{
		Mode:       "TABLE",
		FeePercent: 3.0,
		FXUSDJPY:   150.0,
	})

	tests := []struct {
		name        string
		usdCents    int
		expectedJPY int
	}{
		{
			name:        "$1.00 = ¥150",
			usdCents:    100, // $1.00 = 100 cents
			expectedJPY: 150, // $1.00 * 150 = ¥150
		},
		{
			name:        "$10.00 = ¥1500",
			usdCents:    1000, // $10.00 = 1000 cents
			expectedJPY: 1500, // $10.00 * 150 = ¥1500
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.ConvertToJPY(tt.usdCents)
			if result != tt.expectedJPY {
				t.Errorf("ConvertToJPY(%d) = %d, want %d",
					tt.usdCents, result, tt.expectedJPY)
			}
		})
	}
}

