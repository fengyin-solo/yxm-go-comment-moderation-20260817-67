package service

import (
	"sort"

	"commentmoderation/internal/model"
)

// GetModerationRecord 查询审核记录详情。
func (s *Service) GetModerationRecord(id string) (*model.ModerationRecord, error) {
	return s.store.GetModerationRecord(id)
}

// ListModerationRecords 分页查询审核记录，按时间倒序。
func (s *Service) ListModerationRecords(filter model.ModerationFilter, page, size int) ([]*model.ModerationRecord, int, error) {
	all := s.store.ListModerationRecords()
	matched := make([]*model.ModerationRecord, 0, len(all))
	for _, m := range all {
		if filter.Match(m) {
			matched = append(matched, m)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.ModerationRecord{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}
