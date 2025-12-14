package util

// Helper function to get last N characters
func GetLastN(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
