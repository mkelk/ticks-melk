package calculator

import "testing"

func TestAdd(t *testing.T) {
	cases := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"positive numbers", 2, 3, 5},
		{"zeros", 0, 0, 0},
		{"negative and positive", -1, 1, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Add(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("Add(%d, %d) = %d, want %d", tc.a, tc.b, result, tc.expected)
			}
		})
	}
}

func TestSubtract(t *testing.T) {
	cases := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"positive result", 5, 3, 2},
		{"zeros", 0, 0, 0},
		{"negative result", 1, 5, -4},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Subtract(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("Subtract(%d, %d) = %d, want %d", tc.a, tc.b, result, tc.expected)
			}
		})
	}
}

func TestMultiply(t *testing.T) {
	cases := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"positive numbers", 2, 3, 6},
		{"multiply by zero", 0, 5, 0},
		{"negative and positive", -2, 3, -6},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Multiply(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("Multiply(%d, %d) = %d, want %d", tc.a, tc.b, result, tc.expected)
			}
		})
	}
}
