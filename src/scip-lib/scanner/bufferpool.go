package scanner

import (
	"sync"
)

// SyncPool is an interface compatible with sync.Pool.
type SyncPool interface {
	Get() any
	Put(any)
}

// BufferPool is a wrapper for sync.Pool that returns variable size buffers.
// It uses a set of sync.Pools under the hood to bucket the buffers (defined
// by the startSize and sizeCount).
// Buffer requests beyond the underlying sync pool will be one-off allocations.
// Ideally a consumer picks a set of pools/sizes that minimizes the amount of one-off
// allocations.
type BufferPool struct {
	pools   []SyncPool
	sizes   []int
	maxSize int
	mu      sync.Mutex
}

// NewBufferPool creates a new BufferPool. startSize indicates the initial
// bucket's size, which gets doubled for the subsequent buckets, up to the
// number defined in sizeCnt.
func NewBufferPool(startSize, sizeCnt int) *BufferPool {
	sizes := make([]int, sizeCnt)
	pools := make([]SyncPool, sizeCnt)
	for i := 0; i < sizeCnt; i++ {
		sizes[i] = startSize << i
		size := sizes[i]
		pools[i] = &sync.Pool{
			New: func() interface{} {
				return make([]byte, size)
			},
		}
	}

	return &BufferPool{
		pools:   pools,
		sizes:   sizes,
		maxSize: sizes[sizeCnt-1],
	}
}

// Get returns a buffer of a specific size.
func (bp *BufferPool) Get(size int) []byte {
	for i, bucketSize := range bp.sizes {
		if size <= bucketSize {
			return bp.pools[i].Get().([]byte)[:size]
		}
	}
	// If size is larger than any predefined bucket, create a new buffer
	return make([]byte, size)
}

// Put discards a specific size bufer back into the pools.
func (bp *BufferPool) Put(buf []byte) {
	// Only need to return buffers that are within the bucket range
	for i, bucketSize := range bp.sizes {
		if cap(buf) == bucketSize {
			bp.pools[i].Put(buf[:cap(buf)])
			return
		}
	}
}
