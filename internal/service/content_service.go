package service

import (
	"sort"
	"time"

	"commentmoderation/internal/model"
	"commentmoderation/internal/store"
	"commentmoderation/pkg/idgen"
)

// CreateContent 创建内容对象。
func (s *Service) CreateContent(input model.Content) (*model.Content, error) {
	input.ID = ""
	input.CommentCount = 0
	if err := input.Validate(); err != nil {
		return nil, err
	}
	now := time.Now()
	input.ID = idgen.Hex()
	input.CreatedAt = now
	input.UpdatedAt = now
	if err := s.store.CreateContent(&input); err != nil {
		return nil, err
	}
	s.log.Infof("创建内容 %s", input.Title)
	return &input, nil
}

// GetContent 查询内容详情。
func (s *Service) GetContent(id string) (*model.Content, error) {
	return s.store.GetContent(id)
}

// ListContents 分页查询内容列表。
func (s *Service) ListContents(filter model.ContentFilter, page, size int) ([]*model.Content, int, error) {
	all := s.store.ListContents()
	matched := make([]*model.Content, 0, len(all))
	for _, c := range all {
		if filter.Match(c) {
			matched = append(matched, c)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Content{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UpdateContent 更新内容字段。
func (s *Service) UpdateContent(id, title, category, status string) (*model.Content, error) {
	c, err := s.store.GetContent(id)
	if err != nil {
		return nil, err
	}
	if title != "" {
		c.Title = title
	}
	if category != "" {
		c.Category = category
	}
	if status != "" {
		c.Status = status
	}
	c.UpdatedAt = time.Now()
	if err := c.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.UpdateContent(c); err != nil {
		return nil, err
	}
	return c, nil
}

// DeleteContent 删除内容；若仍有未删除评论则拒绝删除。
func (s *Service) DeleteContent(id string) error {
	if _, err := s.store.GetContent(id); err != nil {
		return err
	}
	for _, c := range s.store.ListComments() {
		if c.ContentID == id && c.Status != model.CommentDeleted {
			return store.ErrConflict
		}
	}
	return s.store.DeleteContent(id)
}
