// Package datastore — найнижчий шар: "сире" сховище записів за числовим id.
// Репозиторій ховає від сервісу, що саме за ним стоїть (map, файл, SQL, MongoDB...).
package datastore

import (
	"sort"
	"sync"
)

// MemoryStore — потокобезпечне in-memory сховище записів типу T.
type MemoryStore[T any] struct {
	mu     sync.RWMutex
	items  map[int]T
	lastID int
}

func NewMemoryStore[T any]() *MemoryStore[T] {
	return &MemoryStore[T]{items: make(map[int]T)}
}

// NextID видає наступний унікальний id (аналог AUTO_INCREMENT).
func (s *MemoryStore[T]) NextID() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastID++
	return s.lastID
}

func (s *MemoryStore[T]) Get(id int) (T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.items[id]
	return v, ok
}

func (s *MemoryStore[T]) Put(id int, v T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[id] = v
	if id > s.lastID {
		s.lastID = id
	}
}

func (s *MemoryStore[T]) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	return true
}

// List повертає всі записи, впорядковані за id.
func (s *MemoryStore[T]) List() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]int, 0, len(s.items))
	for id := range s.items {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	out := make([]T, 0, len(ids))
	for _, id := range ids {
		out = append(out, s.items[id])
	}
	return out
}
