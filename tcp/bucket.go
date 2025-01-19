package tcp

import (
	"sync"
)

// TCPConnBucket manages and stores TCPConn connections.
type TCPConnBucket struct {
	m  map[string]*TCPConn
	mu sync.RWMutex
}

// NewTCPConnBucket initializes a new TCPConnBucket.
func NewTCPConnBucket() *TCPConnBucket {
	return &TCPConnBucket{
		m: make(map[string]*TCPConn),
	}
}

// Put adds or replaces a connection in the bucket.
// If a connection with the same ID exists, it is closed before being replaced.
func (b *TCPConnBucket) Put(id string, c *TCPConn) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if conn, exists := b.m[id]; exists {
		conn.Close()
	}
	b.m[id] = c
}

// Get retrieves a connection by its ID.
func (b *TCPConnBucket) Get(id string) *TCPConn {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.m[id]
}

// Delete removes a connection from the bucket by its ID.
func (b *TCPConnBucket) Delete(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.m, id)
}

// GetAll returns a copy of all connections in the bucket.
func (b *TCPConnBucket) GetAll() map[string]*TCPConn {
	b.mu.RLock()
	defer b.mu.RUnlock()
	allConns := make(map[string]*TCPConn, len(b.m))
	for k, v := range b.m {
		allConns[k] = v
	}
	return allConns
}

// removeClosedTCPConn removes all closed connections from the bucket.
func (b *TCPConnBucket) removeClosedTCPConn() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for key, conn := range b.m {
		if conn.IsClosed() {
			delete(b.m, key)
		}
	}
}
