package store

import (
	"commentmoderation/internal/model"
)

func (s *MemoryStore) CreateModerationRecord(m *model.ModerationRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.moderations[m.ID] = m
	return nil
}

func (s *MemoryStore) GetModerationRecord(id string) (*model.ModerationRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.moderations[id]
	if !ok {
		return nil, ErrNotFound
	}
	return m, nil
}

func (s *MemoryStore) ListModerationRecords() []*model.ModerationRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.ModerationRecord, 0, len(s.moderations))
	for _, m := range s.moderations {
		list = append(list, m)
	}
	return list
}
