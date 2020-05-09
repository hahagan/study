package sort

import (
	"fmt"
)

func Insert(a []int) {
	fmt.Println("")
	length := len(a)
	for i := 1; i < length; i++ {
		tmp := a[i]

		for j := i - 1; j >= 0; j-- {
			if a[j] > tmp {
				a[j+1] = a[j]
				if j == 0 {
					a[j] = tmp
				}
			} else {
				a[j+1] = tmp
				break
			}
		}
	}
}

func TwoPointInsert(a []int) {
	length := len(a)

	for i := 1; i < length; i++ {
		tmp := a[i]
		h := i
		l := 0
		for l < h {
			var m int
			m = (h + l) / 2
			if a[m] > tmp {
				h = m - 1
			} else {
				l = m + 1
			}
		}
		for j := i; j > l; j-- {
			a[j] = a[j-1]
		}
		a[l] = tmp
	}
}

func Buble(a []int) {
	length := len(a)
	for i := 0; i < length-1; i++ {
		for j := i + 1; j < length; j++ {
			if a[i] > a[j] {
				tmp := a[j]
				a[j] = a[i]
				a[i] = tmp
			}
		}

	}
}
