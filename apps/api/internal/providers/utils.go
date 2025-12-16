package providers

// stringPtr returns a pointer to the given string, or nil if empty
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// intPtr returns a pointer to the given int
func intPtr(i int) *int {
	return &i
}

