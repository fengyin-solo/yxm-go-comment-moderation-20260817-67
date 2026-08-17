package handler

import (
	"net/http"
	"strconv"

	"commentmoderation/pkg/httpx"
)

func (s *Server) registerStatsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/stats/overview", s.statsOverview)
	mux.HandleFunc("GET /api/stats/by-content", s.statsByContent)
	mux.HandleFunc("GET /api/stats/top-moderators", s.statsTopModerators)
}

func (s *Server) statsOverview(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.Stats())
}

func (s *Server) statsByContent(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.StatsByContent())
}

func (s *Server) statsTopModerators(w http.ResponseWriter, r *http.Request) {
	n, _ := strconv.Atoi(r.URL.Query().Get("n"))
	httpx.OK(w, s.svc.TopModerators(n))
}
