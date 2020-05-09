package sort

import (
	// "fmt"
	"testing"
)

func TestInsertSort(t *testing.T) {
	b := []int{1, 1, 1, 1, 2, 2, 2, 3, 4, 6, 6, 6, 8}
	a := []int{2, 2, 2, 3, 4, 1, 8, 6, 6, 6, 1, 1, 1}
	Insert(a)
	for i := 0; i < len(b); i++ {
		if a[i] != b[i] || len(a) != len(b) {
			t.Errorf("Insert sort error, %v", a)
			break
		}
	}
}

func TestTwoPointInsertSort(t *testing.T) {
	b := []int{1, 1, 1, 1, 2, 2, 2, 3, 4, 6, 6, 6, 8}
	a := []int{2, 2, 2, 3, 4, 1, 8, 6, 6, 6, 1, 1, 1}
	TwoPointInsert(a)
	for i := 0; i < len(b); i++ {
		if a[i] != b[i] || len(a) != len(b) {
			t.Errorf("Insert sort error, %v", a)
			break
		}
	}
}

func TestBubleInsertSort(t *testing.T) {
	b := []int{1, 1, 1, 1, 2, 2, 2, 3, 4, 6, 6, 6, 8}
	a := []int{2, 2, 2, 3, 4, 1, 8, 6, 6, 6, 1, 1, 1}
	Buble(a)
	for i := 0; i < len(b); i++ {
		if a[i] != b[i] || len(a) != len(b) {
			t.Errorf("Insert sort error, %v", a)
			break
		}
	}
}

func TestFasetSort(t *testing.T) {
	b := []int{1, 1, 1, 1, 2, 2, 2, 3, 4, 6, 6, 6, 8}
	a := []int{2, 2, 2, 3, 4, 1, 8, 6, 6, 6, 1, 1, 1}
	FastSort(a, 0, len(a)-1)
	for i := 0; i < len(b); i++ {
		if a[i] != b[i] || len(a) != len(b) {
			t.Errorf("Insert sort error, %v", a)
			break
		}
	}
}

func TestMergerSort(t *testing.T) {
	b := []int{1, 1, 1, 1, 2, 2, 2, 3, 4, 6, 6, 6, 8}
	a := []int{2, 2, 2, 3, 4, 1, 8, 6, 6, 6, 1, 1, 1}
	a = MergeSort(a)
	for i := 0; i < len(b); i++ {
		if a[i] != b[i] || len(a) != len(b) {
			t.Errorf("Insert sort error, %v", a)
			break
		}
	}
}

func TestHeapSort(t *testing.T) {
	b := []int{1, 1, 1, 1, 2, 2, 2, 3, 4, 6, 6, 6, 8}
	a := []int{2, 2, 2, 3, 4, 1, 8, 6, 6, 6, 1, 1, 1}
	HeapSort(a)
	for i := 0; i < len(b); i++ {
		if a[i] != b[i] || len(a) != len(b) {
			t.Errorf("Insert sort error, %v", a)
			break
		}
	}
}
