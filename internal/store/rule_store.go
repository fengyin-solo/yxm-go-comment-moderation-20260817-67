package store

import (
	"commentmoderation/internal/model"
)

func (s *MemoryStore) CreateRule(r *model.Rule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules[r.ID] = r
	return nil
}

func (s *MemoryStore) GetRule(id string) (*model.Rule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.rules[id]
	if !ok {
		return nil, ErrNotFound
	}
	return r, nil
}

func (s *MemoryStore) ListRules() []*model.Rule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Rule, 0, len(s.rules))
	for _, r := range s.rules {
		// 返回全部规则：全局规则（ContentID 为空）与内容定向规则都应可见，
		// 作用域过滤交给 Rule.AppliesTo 判断，避免漏掉全局规则。
		list = append(list, r)
	}
	return list
}

func (s *MemoryStore) UpdateRule(r *model.Rule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.rules[r.ID]; !ok {
		return ErrNotFound
	}
	s.rules[r.ID] = r
	return nil
}

func (s *MemoryStore) DeleteRule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.rules[id]; !ok {
		return ErrNotFound
	}
	delete(s.rules, id)
	return nil
}
