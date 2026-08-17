package store

import (
	"commentmoderation/internal/model"
)

func (s *MemoryStore) CreateContent(c *model.Content) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.contents[c.ID] = c
	return nil
}

func (s *MemoryStore) GetContent(id string) (*model.Content, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.contents[id]
	if !ok {
		return nil, ErrNotFound
	}
	return c, nil
}

func (s *MemoryStore) ListContents() []*model.Content {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Content, 0, len(s.contents))
	for _, c := range s.contents {
		list = append(list, c)
	}
	return list
}

func (s *MemoryStore) UpdateContent(c *model.Content) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.contents[c.ID]; !ok {
		return ErrNotFound
	}
	s.contents[c.ID] = c
	return nil
}

func (s *MemoryStore) DeleteContent(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.contents[id]; !ok {
		return ErrNotFound
	}
	delete(s.contents, id)
	return nil
}
