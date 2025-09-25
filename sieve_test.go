package main

import "testing"

func Test_getPrimesAlternating(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		primesByCurMultiple map[int]int
	}{
		// TODO: Add test cases.
		{"", make(map[int]int)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getPrimesAlternating(tt.primesByCurMultiple)
		})
	}
}
