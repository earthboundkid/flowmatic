// Package flowpool has pools for reusing buffers.
package flowpool

import (
	"bytes"
	"sync"
)

// Buffer is a pool of *bytes.Buffers.
type Buffer struct {
	pool sync.Pool
}

// Get returns a *bytes.Buffer
// and a function to put it back in the pool.
func (bp *Buffer) Get() (buf *bytes.Buffer) {
	if v := bp.pool.Get(); v != nil {
		return v.(*bytes.Buffer)
	}
	return new(bytes.Buffer)
}

func (bp *Buffer) Put(buf *bytes.Buffer) {
	buf.Reset()
	bp.pool.Put(buf)
}

var defaultPool Buffer

// GetBuffer returns a *bytes.Buffer
// and a function to put it back in the package default pool.
func GetBuffer() (buf *bytes.Buffer) {
	return defaultPool.Get()
}

func PutBuffer(buf *bytes.Buffer) {
	defaultPool.Put(buf)
}
