// Package builderpool is a small wrapper of sync.Pool to hold bytes.Buffer.
//
// Derived from https://github.com/nasa9084/go-builderpool 0ff03b3
package builderpool

import (
	"bytes"
	"sync"
)

// ByteBuilderPool is wrapper struct of sync.Pool for bytes.Buffer objects.
// A comparison against strings.Builder showed that bytes,Buffer is much
// more effective to pool. Using ByteBuilderPool is about 10% faster than
// plain strings.Builder or bytes.Buffer.
type ByteBuilderPool struct {
	pool sync.Pool
}

// NewB returns a new ByteBuilderPool instance.
func NewB() *ByteBuilderPool {
	bp := ByteBuilderPool{}
	bp.pool.New = allocBuilderB
	return &bp
}

func allocBuilderB() interface{} {
	return &bytes.Buffer{}
}

// Get returns a bytes.Buffer from the pool.
func (bp *ByteBuilderPool) Get() *bytes.Buffer {
	return bp.pool.Get().(*bytes.Buffer)
}

// Release puts the given strings.Builder back into the pool
// after resetting the builder.
func (bp *ByteBuilderPool) Release(b *bytes.Buffer) {
	b.Reset()
	bp.pool.Put(b)
}
