package store

import (
	"sync"

	"commentmoderation/internal/model"
)

// MemoryStore 基于内存 map 的 Store 实现，所有操作线程安全。
type MemoryStore struct {
	mu         sync.RWMutex
	contents   map[string]*model.Content
	comments   map[string]*model.Comment
	moderations map[string]*model.ModerationRecord
	rules      map[string]*model.Rule
	reports    map[string]*model.Report
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		contents:    make(map[string]*model.Content),
		comments:    make(map[string]*model.Comment),
		moderations: make(map[string]*model.ModerationRecord),
		rules:       make(map[string]*model.Rule),
		reports:     make(map[string]*model.Report),
	}
}

var _ Store = (*MemoryStore)(nil)
