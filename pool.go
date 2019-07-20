package tmplts

import (
	"bytes"
	"sync"
)

type templatePool struct {
	sync.Pool
}

func (p *templatePool) get() *bytes.Buffer {
	return p.Pool.Get().(*bytes.Buffer)
}

func (p *templatePool) put(b *bytes.Buffer) {
	b.Reset()
	p.Pool.Put(b)
}

func newPool() *templatePool {
	p := &templatePool{}

	p.Pool.New = func() interface{} {
		return &bytes.Buffer{}
	}

	return p
}
