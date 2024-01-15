package qsort

import (
	"fmt"
	"testing"
)

func TestQSort(t *testing.T) {
	values := []int64{42, 9, 101, 95, 27, 25}

	Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})

	fmt.Println(values)
}
