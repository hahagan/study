package sort

func FastSort(a []int, l int, h int) {
	if l < h {
		p := partition(a, l, h)
		FastSort(a, l, p-1)
		FastSort(a, p+1, h)
	}

}

func partition(a []int, l int, h int) int {
	tmp := a[l]
	for l < h {
		for h > l && a[h] >= tmp {
			h--
		}
		a[l] = a[h]
		for h > l && a[l] <= tmp {
			l++
		}
		a[h] = a[l]
	}
	a[l] = tmp
	return l
}
