package handler

import (
	"net/http"

	"commentmoderation/internal/model"
	"commentmoderation/pkg/httpx"
)

func (s *Server) registerReportRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/reports", s.createReport)
	mux.HandleFunc("GET /api/reports", s.listReports)
	mux.HandleFunc("GET /api/reports/{id}", s.getReport)
	mux.HandleFunc("POST /api/reports/{id}/resolve", s.resolveReport)
	mux.HandleFunc("POST /api/reports/{id}/dismiss", s.dismissReport)
}

type createReportRequest struct {
	CommentID  string `json:"comment_id"`
	ReporterID string `json:"reporter_id"`
	Reason     string `json:"reason"`
}

func (s *Server) createReport(w http.ResponseWriter, r *http.Request) {
	var req createReportRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	report, err := s.svc.CreateReport(req.CommentID, req.ReporterID, req.Reason)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, report)
}

func (s *Server) listReports(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.ReportFilter{
		CommentID: r.URL.Query().Get("comment_id"),
		Status:    r.URL.Query().Get("status"),
	}
	items, total, err := s.svc.ListReports(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getReport(w http.ResponseWriter, r *http.Request) {
	report, err := s.svc.GetReport(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, report)
}

type handleReportRequest struct {
	Handler string `json:"handler"`
}

func (s *Server) resolveReport(w http.ResponseWriter, r *http.Request) {
	var req handleReportRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	report, err := s.svc.ResolveReport(r.PathValue("id"), req.Handler)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, report)
}

func (s *Server) dismissReport(w http.ResponseWriter, r *http.Request) {
	var req handleReportRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	report, err := s.svc.DismissReport(r.PathValue("id"), req.Handler)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, report)
}
