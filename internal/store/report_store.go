package store

import (
	"commentmoderation/internal/model"
)

func (s *MemoryStore) CreateReport(r *model.Report) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reports[r.ID] = r
	return nil
}

func (s *MemoryStore) GetReport(id string) (*model.Report, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.reports[id]
	if !ok {
		return nil, ErrNotFound
	}
	return r, nil
}

func (s *MemoryStore) ListReports() []*model.Report {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Report, 0, len(s.reports))
	for _, r := range s.reports {
		list = append(list, r)
	}
	return list
}

func (s *MemoryStore) UpdateReport(r *model.Report) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.reports[r.ID]; !ok {
		return ErrNotFound
	}
	if r.Status == model.ReportDismissed {
		r.HandledBy = ""
	}
	s.reports[r.ID] = r
	return nil
}
