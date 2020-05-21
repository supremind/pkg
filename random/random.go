// Package random helps using crypto/rand as simple as math/rand
package random

import (
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"math"
	mrand "math/rand"
)

const (
	Digits   = "0123456789"
	Lowers   = "abcdefghijklmnopqrstuvwxyz"
	Capitals = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Signs    = `+/_=!@#$%^&*()-[]{}\|,.<>?~:;`
	Alphabet = Digits + Lowers + Capitals + Signs
)

const RandomAlphabetLength = len(Alphabet)

// a math/random.Rand or a (wrapped) crypto/rand.Generator
type metaGenerator interface {
	Uint64() uint64
	Uint32() uint32
}

type randomGenerator struct {
	metaGenerator
}

var (
	RandomGenerator       *randomGenerator
	SecureRandomGenerator *randomGenerator
)

func init() {
	SecureRandomGenerator = &randomGenerator{
		metaGenerator: &secureMetaGenerator{},
	}
	RandomGenerator = &randomGenerator{
		metaGenerator: mrand.New(mrand.NewSource(SecureRandomGenerator.Int63())),
	}
}

// sampling n times with replacement from dict
func (rg *randomGenerator) Pick(dict []byte, n int) []byte {
	base := len(dict)
	secLen := bits(uint64(base))
	var mask uint64 = 1<<secLen - 1
	secInI64 := int(64 / secLen)
	nuance := make([]byte, n)

	var u64 uint64
	for i, s := 0, 0; i < n; {
		if s < 1 {
			u64 = rg.Uint64()
			s = secInI64
		}
		num := int(u64 & mask)
		u64 >>= secLen
		s--
		if num < base {
			nuance[i] = dict[num]
			i++
		}
	}

	return nuance
}

func (rg *randomGenerator) String(length, base int) (string, error) {
	if base <= 0 || base > RandomAlphabetLength {
		return "", errors.New("invalid Alphabet base size for generating random string")
	}
	dict := Alphabet[:base]

	out := rg.Pick([]byte(dict), length)
	return string(out), nil
}

func bits(u uint64) uint64 {
	if u == 0 {
		return 0
	}
	b := uint64(math.Ilogb(float64(u)))
	if u > 1<<b-1 {
		b++
	}
	return b
}

// MustGenString generates a random string with given length and Alphabet size
// length is positive, and base is in [1, 91], or it will panic
// it does not return an error, and is given a long name for use with caution.
func (rg *randomGenerator) MustGenString(length, base int) string {
	s, e := rg.String(length, base)
	if e != nil {
		panic(e.Error())
	}
	return s
}

func (rg *randomGenerator) Int63() int64 {
	u := rg.Uint64()
	return int64(u & (1<<63 - 1))
}

func (rg *randomGenerator) Int31() int32 {
	u := rg.Uint32()
	return int32(u & (1<<31 - 1))
}

func (rg *randomGenerator) Int31n(n int) int {
	// var math/rand/rand.go
	if n <= 0 {
		panic("invalid argument to Int31n")
	}
	if n&(n-1) == 0 { // n is power of two, can mask
		return int(rg.Int31()) & (n - 1)
	}
	max := int32((1 << 31) - 1 - (1<<31)%uint32(n))
	v := rg.Int31()
	for v > max {
		v = rg.Int31()
	}
	return int(v) % n
}

func (rg *randomGenerator) Int31Range(min, max int) int {
	return rg.Int31n(max-min+1) + min
}

var _ metaGenerator = &secureMetaGenerator{}

type secureMetaGenerator struct{}

func (sg *secureMetaGenerator) Uint64() uint64 {
	var u uint64
	e := binary.Read(crand.Reader, binary.BigEndian, &u)
	if e != nil {
		panic("read from crypto/rand: " + e.Error())
	}
	return u
}

func (sg *secureMetaGenerator) Uint32() uint32 {
	var u uint32
	e := binary.Read(crand.Reader, binary.BigEndian, &u)
	if e != nil {
		panic("read from crypto/rand: " + e.Error())
	}
	return u
}
