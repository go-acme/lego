package internal

import (
	"math/rand"
	"strings"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const (
	letterIndexBits = 6                      // 6 bits to represent a letter index
	letterIndexMask = 1<<letterIndexBits - 1 // All 1-bits, as many as letterIndexBits
	letterIndexMax  = 63 / letterIndexBits   // # of letter indices fitting in 63 bits
)

type Nonce struct {
	src rand.Source
}

func NewNonce() *Nonce {
	return &Nonce{
		src: rand.NewSource(time.Now().UnixNano()),
	}
}

func (c Nonce) Generate(length int) string {
	sb := strings.Builder{}
	sb.Grow(length)

	for i, cache, remain := length-1, c.src.Int63(), letterIndexMax; i >= 0; {
		if remain == 0 {
			cache, remain = c.src.Int63(), letterIndexMax
		}

		if idx := int(cache & letterIndexMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])

			i--
		}

		cache >>= letterIndexBits
		remain--
	}

	return sb.String()
}
