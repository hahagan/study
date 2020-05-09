package sort

func MergeSort(a []int) []int {
	l := len(a)
	if l < 2 {
		return a
	}

	i := l / 2
	left := MergeSort(a[:i])
	right := MergeSort(a[i:])
	return merge(left, right)

}

func merge(a []int, b []int) []int {
	tmp := make([]int, 0)
	i := 0
	j := 0
	for i < len(a) && j < len(b) {
		if a[i] <= b[j] {
			tmp = append(tmp, a[i])
			i++
		} else {
			tmp = append(tmp, b[j])
			j++
		}
	}

	for i < len(a) {
		tmp = append(tmp, a[i])

		i++
	}

	for j < len(b) {
		tmp = append(tmp, b[j])

		j++
	}
	return tmp
}
