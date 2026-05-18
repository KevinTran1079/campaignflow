package server

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
}