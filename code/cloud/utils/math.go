package utils

import (
	"errors"
)

// Max returns the maximum element and its index in an integer slice.
// An error is raised if the length of the slice is 0.
func MaxInt(nums []int) (int, int, error) {
	if len(nums) == 0 {
		return 0, 0, errors.New("Empty slice.")
	}
	best := 0
	bestIDx := 0
	for i, num := range nums {
		if best < num {
			best = num
			bestIDx = i
		}
	}
	return best, bestIDx, nil
}
