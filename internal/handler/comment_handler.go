package handler

import (
	"net/http"

	"commentmoderation/internal/model"
	"commentmoderation/pkg/httpx"
)

func (s *Server) registerCommentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/comments", s.postComment)
	mux.HandleFunc("GET /api/comments", s.listComments)
	mux.HandleFunc("GET /api/comments/{id}", s.getComment)
	mux.HandleFunc("POST /api/comments/{id}/approve", s.approveComment)
	mux.HandleFunc("POST /api/comments/{id}/reject", s.rejectComment)
	mux.HandleFunc("DELETE /api/comments/{id}", s.deleteComment)
	mux.HandleFunc("POST /api/contents/{id}/batch-approve", s.batchApprovePending)
}

type postCommentRequest struct {
	ContentID string `json:"content_id"`
	UserID    string `json:"user_id"`
	Body      string `json:"body"`
}

func (s *Server) postComment(w http.ResponseWriter, r *http.Request) {
	var req postCommentRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	c, err := s.svc.PostComment(req.ContentID, req.UserID, req.Body)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, c)
}

func (s *Server) listComments(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.CommentFilter{
		ContentID: r.URL.Query().Get("content_id"),
		UserID:    r.URL.Query().Get("user_id"),
		Status:    r.URL.Query().Get("status"),
	}
	items, total, err := s.svc.ListComments(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getComment(w http.ResponseWriter, r *http.Request) {
	c, err := s.svc.GetComment(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, c)
}

type moderateCommentRequest struct {
	Moderator string `json:"moderator"`
	Reason    string `json:"reason"`
}

func (s *Server) approveComment(w http.ResponseWriter, r *http.Request) {
	var req moderateCommentRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	c, err := s.svc.ApproveComment(r.PathValue("id"), req.Moderator, req.Reason)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, c)
}

func (s *Server) rejectComment(w http.ResponseWriter, r *http.Request) {
	var req moderateCommentRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	c, err := s.svc.RejectComment(r.PathValue("id"), req.Moderator, req.Reason)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, c)
}

func (s *Server) deleteComment(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteComment(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

func (s *Server) batchApprovePending(w http.ResponseWriter, r *http.Request) {
	var req moderateCommentRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	count, err := s.svc.BatchApprovePending(r.PathValue("id"), req.Moderator, req.Reason)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, map[string]int{"approved": count})
}
