package sort

import (
	"testing"
)

func TestInsertSort(t *testing.T) {
	a := []int{2, 3, 4, 1, 8, 6, 1}
	Insert(a)
	for i := 1; i < len(a); i += 2 {
		if a[i-1] > a[i] {
			t.Errorf("Insert sort error, %v", a)
		}
	}
}

func TestTwoPointInsertSort(t *testing.T) {
	a := []int{2, 3, 4, 1, 8, 6, 1}
	TwoPointInsert(a)
	for i := 1; i < len(a); i += 2 {
		if a[i-1] > a[i] {
			t.Errorf("Insert sort error, %v", a)
		}
	}
}
