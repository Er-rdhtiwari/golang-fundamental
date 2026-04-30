package main

import "testing"

func TestSumN(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want int
	}{
		{
			name: "sum of 5",
			n:    5,
			want: 15,
		},
		{
			name: "sum of 1",
			n:    1,
			want: 1,
		},
		{
			name: "sum of 0",
			n:    0,
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SumN(tt.n)
			if got != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}
