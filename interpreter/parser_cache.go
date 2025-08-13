package interpreter

import (
	"sync"
)

// TokenCache provides caching for frequently used tokens to reduce parsing overhead
type TokenCache struct {
	cache map[string]Token
	mutex sync.RWMutex
	maxSize int
}

// NewTokenCache creates a new token cache with specified maximum size
func NewTokenCache(maxSize int) *TokenCache {
	return &TokenCache{
		cache: make(map[string]Token),
		maxSize: maxSize,
	}
}

// Global token cache instance
var globalTokenCache = NewTokenCache(1000)

// GetCachedToken retrieves a token from cache or computes and caches it
func (tc *TokenCache) GetCachedToken(tokenStr string) (Token, bool) {
	tc.mutex.RLock()
	token, exists := tc.cache[tokenStr]
	tc.mutex.RUnlock()
	
	if exists {
		return token, true
	}
	
	return 0, false
}

// SetCachedToken stores a token in the cache
func (tc *TokenCache) SetCachedToken(tokenStr string, token Token) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	
	// Check cache size limit
	if len(tc.cache) >= tc.maxSize {
		// Simple eviction: clear half the cache
		for k := range tc.cache {
			delete(tc.cache, k)
			if len(tc.cache) <= tc.maxSize/2 {
				break
			}
		}
	}
	
	tc.cache[tokenStr] = token
}

// ClearCache clears all cached tokens
func (tc *TokenCache) ClearCache() {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()
	tc.cache = make(map[string]Token)
}

// GetCacheSize returns the current cache size
func (tc *TokenCache) GetCacheSize() int {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()
	return len(tc.cache)
}

// GetGlobalTokenCache returns the global token cache instance
func GetGlobalTokenCache() *TokenCache {
	return globalTokenCache
}

// AST Node Pool for reducing allocation overhead
type ASTNodePool struct {
	binaryPool    sync.Pool
	unaryPool     sync.Pool
	literalPool   sync.Pool
	variablePool  sync.Pool
	callPool      sync.Pool
	listPool      sync.Pool
	mapPool       sync.Pool
	subscriptPool sync.Pool
}

// NewASTNodePool creates a new AST node pool
func NewASTNodePool() *ASTNodePool {
	return &ASTNodePool{
		binaryPool: sync.Pool{
			New: func() any {
				return &Binary{}
			},
		},
		unaryPool: sync.Pool{
			New: func() any {
				return &Unary{}
			},
		},
		literalPool: sync.Pool{
			New: func() any {
				return &Literal{}
			},
		},
		variablePool: sync.Pool{
			New: func() any {
				return &Variable{}
			},
		},
		callPool: sync.Pool{
			New: func() any {
				return &Call{Arguments: make([]Expression, 0, 4)}
			},
		},
		listPool: sync.Pool{
			New: func() any {
				return &List{Values: make([]Expression, 0, 8)}
			},
		},
		mapPool: sync.Pool{
			New: func() any {
				return &Map{Items: make([]MapItem, 0, 4)}
			},
		},
		subscriptPool: sync.Pool{
			New: func() any {
				return &Subscript{}
			},
		},
	}
}

// Global AST node pool
var globalASTNodePool = NewASTNodePool()

// GetBinary gets a Binary node from the pool
func (pool *ASTNodePool) GetBinary() *Binary {
	return pool.binaryPool.Get().(*Binary)
}

// PutBinary returns a Binary node to the pool
func (pool *ASTNodePool) PutBinary(node *Binary) {
	// Reset node state
	node.pos = Position{}
	node.Left = nil
	node.Right = nil
	node.Operator = 0
	pool.binaryPool.Put(node)
}

// GetUnary gets a Unary node from the pool
func (pool *ASTNodePool) GetUnary() *Unary {
	return pool.unaryPool.Get().(*Unary)
}

// PutUnary returns a Unary node to the pool
func (pool *ASTNodePool) PutUnary(node *Unary) {
	node.pos = Position{}
	node.Operand = nil
	node.Operator = 0
	pool.unaryPool.Put(node)
}

// GetLiteral gets a Literal node from the pool
func (pool *ASTNodePool) GetLiteral() *Literal {
	return pool.literalPool.Get().(*Literal)
}

// PutLiteral returns a Literal node to the pool
func (pool *ASTNodePool) PutLiteral(node *Literal) {
	node.pos = Position{}
	node.Value = nil
	pool.literalPool.Put(node)
}

// GetVariable gets a Variable node from the pool
func (pool *ASTNodePool) GetVariable() *Variable {
	return pool.variablePool.Get().(*Variable)
}

// PutVariable returns a Variable node to the pool
func (pool *ASTNodePool) PutVariable(node *Variable) {
	node.pos = Position{}
	node.Name = ""
	pool.variablePool.Put(node)
}

// GetCall gets a Call node from the pool
func (pool *ASTNodePool) GetCall() *Call {
	node := pool.callPool.Get().(*Call)
	node.Arguments = node.Arguments[:0] // Reset slice length
	return node
}

// PutCall returns a Call node to the pool
func (pool *ASTNodePool) PutCall(node *Call) {
	node.pos = Position{}
	node.Function = nil
	node.Arguments = node.Arguments[:0]
	node.Ellipsis = false
	pool.callPool.Put(node)
}

// GetList gets a List node from the pool
func (pool *ASTNodePool) GetList() *List {
	node := pool.listPool.Get().(*List)
	node.Values = node.Values[:0] // Reset slice length
	return node
}

// PutList returns a List node to the pool
func (pool *ASTNodePool) PutList(node *List) {
	node.pos = Position{}
	node.Values = node.Values[:0]
	pool.listPool.Put(node)
}

// GetMap gets a Map node from the pool
func (pool *ASTNodePool) GetMap() *Map {
	node := pool.mapPool.Get().(*Map)
	node.Items = node.Items[:0] // Reset slice length
	return node
}

// PutMap returns a Map node to the pool
func (pool *ASTNodePool) PutMap(node *Map) {
	node.pos = Position{}
	node.Items = node.Items[:0]
	pool.mapPool.Put(node)
}

// GetSubscript gets a Subscript node from the pool
func (pool *ASTNodePool) GetSubscript() *Subscript {
	return pool.subscriptPool.Get().(*Subscript)
}

// PutSubscript returns a Subscript node to the pool
func (pool *ASTNodePool) PutSubscript(node *Subscript) {
	node.pos = Position{}
	node.Container = nil
	node.Subscript = nil
	pool.subscriptPool.Put(node)
}

// GetGlobalASTNodePool returns the global AST node pool
func GetGlobalASTNodePool() *ASTNodePool {
	return globalASTNodePool
}