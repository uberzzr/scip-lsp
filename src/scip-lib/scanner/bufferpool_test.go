package scanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewBufferPool(t *testing.T) {
	pool := NewBufferPool(1024, 4)
	assert.NotNil(t, pool)
	assert.Equal(t, 4, len(pool.sizes))
	assert.Equal(t, 1024, pool.sizes[0])
	assert.Equal(t, 8192, pool.sizes[len(pool.sizes)-1])
}

func TestPoolGetFixedSize(t *testing.T) {
	pool := NewBufferPool(1024, 4)
	buf := pool.Get(1024)
	assert.Equal(t, 1024, len(buf))
}

func TestPoolGetVariableSize(t *testing.T) {
	pool := NewBufferPool(1024, 4)
	buf := pool.Get(1500)
	assert.Equal(t, 1500, len(buf))
	assert.Equal(t, 2048, cap(buf))
}

func TestPoolPutCapSize(t *testing.T) {
	pool, sp1Mock, _ := NewMockBufferPool(t)
	sp1Mock.EXPECT().Get().Return(make([]byte, 1024))
	sp1Mock.EXPECT().Put(gomock.Any())
	buf := pool.Get(1024)
	pool.Put(buf)
}

func TestPoolPutVariableSize(t *testing.T) {
	pool, _, sp2Mock := NewMockBufferPool(t)
	b := make([]byte, 2048)
	sp2Mock.EXPECT().Get().Return(b[:1500])
	sp2Mock.EXPECT().Put(gomock.Any())
	buf := pool.Get(1500)
	pool.Put(buf)
}

func TestPoolOverCapacityBuf(t *testing.T) {
	pool, _, _ := NewMockBufferPool(t)
	pool.sizes[0] = 20
	pool.sizes[1] = 40
	pool.maxSize = 40
	buf := pool.Get(150)
	pool.Put(buf)
}

func NewMockBufferPool(t *testing.T) (*BufferPool, *MockSyncPool, *MockSyncPool) {
	ctrl := gomock.NewController(t)
	sp1Mock := NewMockSyncPool(ctrl)
	sp2Mock := NewMockSyncPool(ctrl)
	return &BufferPool{
		pools:   []SyncPool{sp1Mock, sp2Mock},
		sizes:   []int{1024, 2048},
		maxSize: 2048,
	}, sp1Mock, sp2Mock
}
