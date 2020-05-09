package sort

func HeapSort(a []int) {
	for i := (len(a) - 1 - 1) / 2; i >= 0; i-- {
		heapAdjust(a, i, len(a))
	}

	for i := len(a) - 1; i > 0; i-- {
		tmp := a[i]
		a[i] = a[0]
		a[0] = tmp
		heapAdjust(a, 0, i)
	}
}

func heapAdjust(a []int, i int, l int) {
	tmp := a[i]
	for j := 2*i + 1; j < l; j = 2*j + 1 {
		if j+1 < l && a[j] < a[j+1] {
			j++
		}
		if a[j] <= tmp {
			break
		} else {
			a[i] = a[j]
			a[j] = tmp
			i = j
		}
	}
}
