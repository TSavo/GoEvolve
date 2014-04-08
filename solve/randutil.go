package solve

import (
	"github.com/seehuhn/mt19937"
	"math/rand"
	"time"
)

var (
	rng = rand.New(mt19937.New())
)

func init() {
	rng.Seed(time.Now().UnixNano())
}
