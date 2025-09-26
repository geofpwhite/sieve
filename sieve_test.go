package main

import "testing"

func Test_getPrimesAlternating(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		max int
	}{
		// TODO: Add test cases.
		{"", 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getPrimesAlternating(tt.max)
		})
	}
}

func Test_simpleSieve(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		max int
	}{
		{"", 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			simpleSieve(tt.max)
		})
	}
}
