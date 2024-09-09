package cache

import (
	"sync"
)

type ParsedFileCache struct {
	cache map[string][]byte
	mutex sync.RWMutex
}

func NewParsedFileCache() *ParsedFileCache {
	return &ParsedFileCache{
		cache: make(map[string][]byte),
	}
}

func (p *ParsedFileCache) Get(path string) ([]byte, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	file, ok := p.cache[path]
	return file, ok
}

func (p *ParsedFileCache) Set(path string, file []byte) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.cache[path] = file
}
