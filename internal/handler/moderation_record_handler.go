package handler

import (
	"net/http"

	"commentmoderation/internal/model"
	"commentmoderation/pkg/httpx"
)

func (s *Server) registerModerationRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/moderations", s.listModerations)
	mux.HandleFunc("GET /api/moderations/{id}", s.getModeration)
}

func (s *Server) listModerations(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.ModerationFilter{
		CommentID: r.URL.Query().Get("comment_id"),
		Action:    r.URL.Query().Get("action"),
		Source:    r.URL.Query().Get("source"),
	}
	items, total, err := s.svc.ListModerationRecords(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getModeration(w http.ResponseWriter, r *http.Request) {
	m, err := s.svc.GetModerationRecord(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, m)
}
