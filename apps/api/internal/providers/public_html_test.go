package providers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePrice(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Dollar sign",
			input:    "$79.99",
			expected: 7999,
		},
		{
			name:     "Plain number",
			input:    "149.99",
			expected: 14999,
		},
		{
			name:     "With USD text",
			input:    "USD 89.99",
			expected: 8999,
		},
		{
			name:     "With comma",
			input:    "$1,299.99",
			expected: 129999,
		},
		{
			name:     "Invalid",
			input:    "invalid",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePrice(tt.input)
			if result != tt.expected {
				t.Errorf("parsePrice(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseHTMLFile(t *testing.T) {
	// Create a temporary HTML file
	tmpDir := t.TempDir()
	htmlFile := filepath.Join(tmpDir, "test.html")
	htmlContent := `
	<!DOCTYPE html>
	<html>
	<body>
		<div class="product">
			<h2 class="title">Test Product</h2>
			<img src="https://example.com/image.jpg" alt="Test">
			<div class="offer">
				<span class="seller">Test Seller</span>
				<span class="price">$99.99</span>
			</div>
		</div>
	</body>
	</html>
	`
	if err := os.WriteFile(htmlFile, []byte(htmlContent), 0644); err != nil {
		t.Fatalf("Failed to create test HTML file: %v", err)
	}

	provider := NewPublicHTMLProvider("test-agent")
	provider.samplesDir = tmpDir

	products, err := provider.parseHTMLFile(htmlFile)
	if err != nil {
		t.Fatalf("parseHTMLFile failed: %v", err)
	}

	if len(products) == 0 {
		t.Error("Expected at least one product, got 0")
	}

	if products[0].Title != "Test Product" {
		t.Errorf("Expected title 'Test Product', got %q", products[0].Title)
	}
}

func TestExtractBrand(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected *string
	}{
		{
			name:     "Brand at start",
			title:    "Sony WH-1000XM4 Headphones",
			expected: stringPtr("Sony"),
		},
		{
			name:     "Single word",
			title:    "Headphones",
			expected: nil, // Too short or single word
		},
		{
			name:     "Long first word",
			title:    "SuperLongBrandNameThatIsTooLong Product",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractBrand(tt.title)
			if (result == nil) != (tt.expected == nil) {
				t.Errorf("extractBrand(%q) = %v, want %v", tt.title, result, tt.expected)
			} else if result != nil && tt.expected != nil && *result != *tt.expected {
				t.Errorf("extractBrand(%q) = %q, want %q", tt.title, *result, *tt.expected)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

