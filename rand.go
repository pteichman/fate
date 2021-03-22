package fate

import "sync/atomic"

type intn interface {
	Intn(n int) int
}

type prng struct {
	uint64 uint64
}

func (r *prng) Next() uint64 {
	for {
		c := atomic.LoadUint64(&r.uint64)
		x := c ^ (c >> 12)
		x ^= x << 25
		x ^= x >> 27
		if x == 0 {
			x = 0x4030eab5124e7c33
		}
		if atomic.CompareAndSwapUint64(&r.uint64, c, x) {
			return x * 2685821657736338717
		}
	}
}

func (r *prng) Intn(n int) int {
	// Clamp to uint32 to support 32-bit CPUs: this will silently
	// fail to choose new things if there are ever 2^32 tokens in
	// a tokset.
	return int(uint32(r.Next()>>1) % uint32(n))
}
