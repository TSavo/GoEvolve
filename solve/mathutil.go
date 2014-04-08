package solve

import (
	"github.com/seehuhn/mt19937"
	"math"
	"math/rand"
	"sync"
	"time"
)

// Integer limit values.
const (
	MaxInt   = int(^uint(0) >> 1)
	MinInt   = int(-MaxInt - 1)
	MaxInt32 = int32(math.MaxInt32)
	MinInt32 = int32(math.MinInt32)
	MaxInt64 = int64(math.MaxInt64)
	MinInt64 = int64(math.MinInt64)
)

type SafeRNG struct {
	rng *rand.Rand
}

var (
	rng  SafeRNG
	lock sync.Mutex
)

func init() {
	rng = SafeRNG{rand.New(mt19937.New())}
	rng.rng.Seed(time.Now().UnixNano())
	lock = sync.Mutex{}
}

func (rng *SafeRNG) Int() int {
	lock.Lock()
	defer lock.Unlock()
	return rng.rng.Int()
}

func (rng *SafeRNG) Int63() int64 {
	lock.Lock()
	defer lock.Unlock()
	return rng.rng.Int63()
}

func (rng *SafeRNG) Float64() float64 {
	lock.Lock()
	defer lock.Unlock()
	return rng.rng.Float64()
}

func (rng *SafeRNG) SmallInt() int {
	lock.Lock()
	defer lock.Unlock()
	x := rng.rng.Intn(10000)
	if rng.rng.Float64() < 0.5 {
		x *= -1
	}
	return x
}

func Max8(left, right int8) int8 {
	if left > right {
		return left
	} else {
		return right
	}
}

func Min8(left, right int8) int8 {
	if left < right {
		return left
	} else {
		return right
	}
}

func Max(left, right int) int {
	if left > right {
		return left
	} else {
		return right
	}
}

func Min(left, right int) int {
	if left < right {
		return left
	} else {
		return right
	}
}

func Max32(left, right int32) int32 {
	if left > right {
		return left
	} else {
		return right
	}
}

func Min32(left, right int32) int32 {
	if left < right {
		return left
	} else {
		return right
	}
}

func Max64(left, right int64) int64 {
	if left > right {
		return left
	} else {
		return right
	}
}

func Min64(left, right int64) int64 {
	if left < right {
		return left
	} else {
		return right
	}
}

// Abs returns the absolute value of x.
func Abs(x int) int {
	switch {
	case x >= 0:
		return x
	case x > MinInt:
		return -x
	}
	panic("Abs: invalid argument")

}

// Abs32 returns the absolute value of x.
func Abs32(x int32) int32 {
	switch {
	case x >= 0:
		return x
	case x > MinInt32:
		return -x
	}
	panic("Abs32: invalid argument")

}

// Abs64 returns the absolute value of x.
func Abs64(x int64) int64 {
	switch {
	case x >= 0:
		return x
	case x > MinInt64:
		return -x
	}
	panic("Abs64: invalid argument")
}
