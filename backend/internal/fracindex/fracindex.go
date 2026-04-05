package fracindex

import "fmt"

const (
	minByte = byte(0x21) // '!'
	maxByte = byte(0x7E) // '~'
)

func GenerateKeyBetween(a, b string) (string, error) {
	if a != "" && b != "" && a >= b {
		return "", fmt.Errorf("invalid bounds: %q >= %q", a, b)
	}
	switch {
	case a == "" && b == "":
		return string(midByte(minByte, maxByte)), nil
	case a == "":
		return before(b), nil
	case b == "":
		return after(a), nil
	default:
		return between(a, b), nil
	}
}

func between(a, b string) string {
	n := min(len(a), len(b))
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			if b[i]-a[i] > 1 {
				return a[:i] + string(midByte(a[i], b[i]))
			}
			return a[:i+1] + after(a[i+1:])
		}
	}
	return a + before(b[len(a):])
}

func before(b string) string {
	for i := 0; i < len(b); i++ {
		if b[i] > minByte {
			mid := midByte(minByte, b[i])
			if mid > minByte {
				return b[:i] + string(mid)
			}
			return b[:i] + string(minByte) + string(midByte(minByte, maxByte))
		}
	}
	return b + string(midByte(minByte, maxByte))
}

func after(a string) string {
	for i := len(a) - 1; i >= 0; i-- {
		if a[i] < maxByte {
			return a[:i] + string(midByte(a[i], maxByte))
		}
	}
	return a + string(midByte(minByte, maxByte))
}

func midByte(lo, hi byte) byte {
	return lo + (hi-lo)/2
}
