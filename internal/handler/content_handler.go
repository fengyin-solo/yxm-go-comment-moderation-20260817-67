package handler

import (
	"net/http"

	"commentmoderation/internal/model"
	"commentmoderation/pkg/httpx"
)

func (s *Server) registerContentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/contents", s.createContent)
	mux.HandleFunc("GET /api/contents", s.listContents)
	mux.HandleFunc("GET /api/contents/{id}", s.getContent)
	mux.HandleFunc("PUT /api/contents/{id}", s.updateContent)
	mux.HandleFunc("DELETE /api/contents/{id}", s.deleteContent)
}

type createContentRequest struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	Category string `json:"category"`
}

func (s *Server) createContent(w http.ResponseWriter, r *http.Request) {
	var req createContentRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	c, err := s.svc.CreateContent(model.Content{
		Title:    req.Title,
		Author:   req.Author,
		Category: req.Category,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, c)
}

func (s *Server) listContents(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.ContentFilter{
		Category: r.URL.Query().Get("category"),
		Status:   r.URL.Query().Get("status"),
		Keyword:  r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListContents(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getContent(w http.ResponseWriter, r *http.Request) {
	c, err := s.svc.GetContent(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, c)
}

type updateContentRequest struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	Status   string `json:"status"`
}

func (s *Server) updateContent(w http.ResponseWriter, r *http.Request) {
	var req updateContentRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	c, err := s.svc.UpdateContent(r.PathValue("id"), req.Title, req.Category, req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, c)
}

func (s *Server) deleteContent(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteContent(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
