package v1

import "net/http"

func webdavOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(
		"Allow",
		"OPTIONS, GET, HEAD, POST, PUT, DELETE, TRACE, COPY, MOVE, MKCOL, PROPFIND, PROPPATCH, LOCK, UNLOCK, ORDERPATCH",
	)
	w.Header().Set(
		"DAV",
		"1, 2, ordered-collections",
	)
}
