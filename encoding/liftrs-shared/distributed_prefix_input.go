package liftrs

import "fmt"

// rankPrefixInputWords selects an immutable source view: coefficients before
// count are used;the rest are logically zero irrespective of the source tail.
// The GPU caller must zero its destination before uploading this intersection.
// This is distinct from dense input,whose complete padding must validate as zero.
func rankPrefixInputWords(domain, first, words, count uint64) (uint64, error) {
	if domain == 0 || count == 0 || count > domain || first > domain || words > domain-first {
		return 0, fmt.Errorf("prefix input extent")
	}
	return min(words, count-min(count, first)), nil
}

func newtonInputMode(value string) (string, error) {
	if value == "" {
		return "dense", nil
	}
	if value != "dense" && value != "prefix" {
		return "", fmt.Errorf("Newton input mode")
	}
	return value, nil
}
