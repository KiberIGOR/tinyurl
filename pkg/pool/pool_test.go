package pool_test

import (
	"testing"

	"github.com/KiberIGOR/tinyurl/pkg/pool"
)

type counter struct {
	N int
	S []int
}

func (c *counter) Reset() {
	c.N = 0
	c.S = c.S[:0]
}

func TestPoolGetPut(t *testing.T) {
	p := pool.New[counter, *counter]()

	x := p.Get()
	x.N = 42
	x.S = append(x.S, 1, 2)

	p.Put(x)

	y := p.Get()
	if y.N != 0 || len(y.S) != 0 {
		t.Fatalf("object was not reset: %+v", y)
	}
}
