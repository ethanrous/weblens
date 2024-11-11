package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/gin-gonic/gin"
)

func getServices(ctx *gin.Context) *models.ServicePack {
	srv, ok := ctx.Get("services")
	if !ok {
		return nil
	}
	return srv.(*models.ServicePack)
}

// readCtxBody reads the body of a gin context and unmarshal it into the given generic type.
// It returns the unmarshalled object or an error if one occurred. It also sets the response status
// in the context and logs the error if an error occurred so it is recommended, upon reading an error from this function,
// return from a http handler immediately.
func readCtxBody[T any](ctx *gin.Context) (obj T, err error) {
	if ctx.Request.Method == "GET" {
		err = errors.New("trying to get body of get request")
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.ShowErr(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read request body"})
		return
	}
	err = json.Unmarshal(jsonData, &obj)
	if err != nil {
		log.ShowErr(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Request body is not in expected JSON format"})
		return
	}

	return
}

func readRespBody[T any](resp *http.Response) (obj T, err error) {
	var bodyB []byte
	if resp.ContentLength == 0 {
		return obj, werror.ErrNoBody
	} else if resp.ContentLength == -1 {
		log.Warning.Println("Reading body with unknown content length")
		bodyB, err = io.ReadAll(resp.Body)
	} else {
		bodyB, err = internal.OracleReader(resp.Body, resp.ContentLength)
	}
	if err != nil {
		return
	}
	err = json.Unmarshal(bodyB, &obj)
	return
}

func readRespBodyRaw(resp *http.Response) (bodyB []byte, err error) {
	if resp.ContentLength == 0 {
		return nil, werror.ErrNoBody
	} else if resp.ContentLength == -1 {
		log.Warning.Println("Reading body with unknown content length")
		bodyB, err = io.ReadAll(resp.Body)
	} else {
		bodyB, err = internal.OracleReader(resp.Body, resp.ContentLength)
	}
	return
}

func getUserFromCtx(ctx *gin.Context) *models.User {
	user, ok := ctx.Get("user")
	if !ok {
		return nil
	}
	return user.(*models.User)
}

func getInstanceFromCtx(ctx *gin.Context) *models.Instance {
	server, ok := ctx.Get("server")
	if !ok {
		return nil
	}
	return server.(*models.Instance)
}

func getShareFromCtx[T models.Share](ctx *gin.Context) (T, error) {
	pack := getServices(ctx)

	shareId := models.ShareId(ctx.Query("shareId"))
	if shareId == "" {
		shareId = models.ShareId(ctx.Param("shareId"))
	}
	var empty T
	if shareId == "" {
		return empty, nil
	}

	sh := pack.ShareService.Get(shareId)
	tsh, ok := sh.(T)
	if sh != nil && ok {
		return tsh, nil
	}

	err := werror.ErrNoShare
	ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	return empty, err
}

type FileStat struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	IsDir   bool      `json:"isDir"`
	ModTime time.Time `json:"modifyTimestamp"`
	Exists  bool      `json:"exists"`
}
