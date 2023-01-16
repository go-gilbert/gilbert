package log

import (
	"bytes"
	"io"
	"os"
	"sync"
)

type sourceFilesPool struct {
	cache *sync.Map
	pool  *sync.Pool
}

func newSourceFilesPool() sourceFilesPool {
	return sourceFilesPool{
		cache: new(sync.Map),
		pool: &sync.Pool{New: func() any {
			return new(bytes.Buffer)
		}},
	}
}

func (p sourceFilesPool) get(path string) ([]byte, error) {
	if data, ok := p.cache.Load(path); ok {
		return data.(*bytes.Buffer).Bytes(), nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buff := p.pool.Get().(*bytes.Buffer)
	if _, err := io.Copy(buff, f); err != nil {
		p.pool.Put(buff)
	}

	p.cache.Store(path, buff)
	return buff.Bytes(), nil
}

func (p sourceFilesPool) dispose() {
	p.cache.Range(func(key, value any) bool {
		p.cache.Delete(key)
		p.pool.Put(value)
		return true
	})
}
