package cache

import (
	"sync"
)

type ParsedFileCache struct {
	cache sync.Map
}

func NewParsedFileCache() *ParsedFileCache {
	return &ParsedFileCache{}
}

func (p *ParsedFileCache) Get(path string) ([]byte, bool) {
	value, ok := p.cache.Load(path)
	if ok {
		return value.([]byte), true
	}
	return nil, false
}

func (p *ParsedFileCache) Set(path string, file []byte) {
	p.cache.Store(path, file)
}
