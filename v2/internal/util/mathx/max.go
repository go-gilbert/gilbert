package mathx

import "golang.org/x/exp/constraints"

type Number interface {
	constraints.Integer | constraints.Float
}

// Max returns the largest of number sequence.
func Max[T Number](a, b T, other ...T) T {
	max := a
	if b > a {
		max = b
	}

	if len(other) == 0 {
		return max
	}

	for _, val := range other {
		if val > max {
			max = val
		}
	}

	return max
}

// Min returns the smallest of number sequence.
func Min[T Number](a, b T, other ...T) T {
	min := a
	if b < a {
		min = b
	}

	if len(other) == 0 {
		return min
	}

	for _, val := range other {
		if val < min {
			min = val
		}
	}

	return min
}
