package cache

import (
	"go/ast"
	"sync"
)

type ParsedFileCache struct {
	cache map[string]*ast.File
	mutex sync.Mutex
}

func NewParsedFileCache() *ParsedFileCache {
	return &ParsedFileCache{
		cache: make(map[string]*ast.File),
	}
}

func (p *ParsedFileCache) Get(path string) (*ast.File, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	file, ok := p.cache[path]
	return file, ok
}

func (p *ParsedFileCache) Set(path string, file *ast.File) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.cache[path] = file
}
