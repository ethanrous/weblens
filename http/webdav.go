package http

import "github.com/gin-gonic/gin"

func webdavOptions(ctx *gin.Context) {
	ctx.Header(
		"Allow",
		"OPTIONS, GET, HEAD, POST, PUT, DELETE, TRACE, COPY, MOVE, MKCOL, PROPFIND, PROPPATCH, LOCK, UNLOCK, ORDERPATCH",
	)
	ctx.Header(
		"DAV",
		"1, 2, ordered-collections",
	)
}
