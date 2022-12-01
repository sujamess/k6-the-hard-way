package uniquer

import (
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

const upperCase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const numeric = "0123456789"
const upperCaseWithNumeric = upperCase + numeric
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func OrderNumber() string {
	const n = 8
	sb, src := strings.Builder{}, rand.NewSource(time.Now().UnixNano())
	sb.Grow(n)

	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(upperCaseWithNumeric) {
			sb.WriteByte(upperCaseWithNumeric[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return time.Now().Format("20060102") + sb.String()
}

func UUID() string {
	return uuid.New().String()
}
