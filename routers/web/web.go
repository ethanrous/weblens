package web

func (s *Server) UseUi() *chi.Mux {
	memFs := &InMemoryFS{routes: make(map[string]*memFileReal, 10), routesMu: &sync.RWMutex{}, Pack: s.services, proxyAddress: env.GetProxyAddress(s.services.Cnf)}
	memFs.loadIndex(s.services.Cnf.UiPath)

	r := chi.NewMux()
	r.Route("/assets", func(r *Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", "public, max-age=3600")
				w.Header().Set("Content-Encoding", "gzip")
				next.ServeHTTP(w, r)
			})
		})
		r.Handle("/*", http.FileServer(memFs))
	})

	r.NotFound(
		func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.RequestURI, "/api") {
				index := memFs.Index(r.RequestURI)
				_, err := w.Write(index.realFile.data)
				SafeErrorAndExit(err, w)
			} else {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		},
	)

	return r
}
