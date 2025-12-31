package v1

import "net/http"

// webdavOptions handles OPTIONS requests for WebDAV protocol discovery and capabilities negotiation.
func webdavOptions(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(
		"Allow",
		"OPTIONS, GET, HEAD, POST, PUT, DELETE, TRACE, COPY, MOVE, MKCOL, PROPFIND, PROPPATCH, LOCK, UNLOCK, ORDERPATCH",
	)
	w.Header().Set(
		"DAV",
		"1, 2, ordered-collections",
	)
}
