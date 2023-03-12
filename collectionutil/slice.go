package collectionutil

func Filter[T any](arr []T, test func(T) bool) (ret []T) {
	for _, s := range arr {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return ret
}

func HasPrefix[T byte | int | uint | string | uint32 | uint64 | uint16 | int64 | int32 | int16 | bool](arr []T, prefix []T) bool {
	arrLen := len(arr)
	prefixLen := len(prefix)
	if prefixLen > arrLen {
		return false
	}
	for i := 0; i < prefixLen; i++ {
		if prefix[i] != arr[i] {
			return false
		}
	}
	return true
}

func HasSuffix[T byte | int | uint | string | uint32 | uint64 | uint16 | int64 | int32 | int16 | bool](arr []T, suffix []T) bool {
	arrLen := len(arr)
	suffixLen := len(suffix)
	if suffixLen > arrLen {
		return false
	}
	diff := arrLen - suffixLen
	for i := diff; i < arrLen; i++ {
		if suffix[i-diff] != arr[i] {
			return false
		}
	}
	return true
}

func LastIndexFunc[T any](arr []T, f func(val T, idx int) bool) int {
	arrLen := len(arr)
	for i := arrLen - 1; i >= 0; i-- {
		if f(arr[i], i) {
			return i
		}
	}
	return -1
}

func IndexFunc[T any](arr []T, f func(val T, idx int) bool) int {
	for i, v := range arr {
		if f(v, i) {
			return i
		}
	}
	return -1
}

func Fill[T any](arr []T, val T) []T {
	for i := 0; i < len(arr); i++ {
		arr[i] = val
	}
	return arr
}

func Equal[T byte | int | uint | string | uint32 | uint64 | uint16 | int64 | int32 | int16 | bool](a []T, b []T) bool {
	l := len(a)
	if l != len(b) {
		return false
	}
	for i := 0; i < l; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
