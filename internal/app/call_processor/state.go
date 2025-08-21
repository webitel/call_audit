package processor

import "sync"

type UUIDState struct {
	mu    sync.Mutex
	inUse map[string]struct{}
}

func NewUUIDState() *UUIDState {
	return &UUIDState{
		inUse: make(map[string]struct{}),
	}
}

func (s *UUIDState) TryAdd(uuid string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.inUse[uuid]; exists {
		return false
	}
	s.inUse[uuid] = struct{}{}
	return true
}

func (s *UUIDState) Remove(uuid string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.inUse, uuid)
}
