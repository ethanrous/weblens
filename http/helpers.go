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
	"github.com/ethanrous/weblens/models/rest"
	"github.com/go-chi/chi/v5"
)

func writeJson(w http.ResponseWriter, status int, obj interface{}) {
	bs, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	w.WriteHeader(status)
	_, err = w.Write(bs)
	if err != nil {
		panic(err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJson(w, status, rest.WeblensErrorInfo{Error: err.Error()})
}

func getServices(r *http.Request) *models.ServicePack {
	srv := r.Context().Value(ServicesKey)
	if srv == nil {
		panic(werror.Errorf("Could not assert services from context"))
	}
	return srv.(*models.ServicePack)
}

func SafeErrorAndExit(err error, w http.ResponseWriter) (shouldExit bool) {
	if err == nil {
		return false
	}
	safe, code := werror.TrySafeErr(err)
	writeError(w, code, safe)
	return true
}

// readCtxBody reads the body of a gin context and unmarshal it into the given generic type.
// It returns the unmarshalled object or an error if one occurred. It also sets the response status
// in the context and logs the error if an error occurred so it is recommended, upon reading an error from this function,
// return from a http handler immediately.
func readCtxBody[T any](w http.ResponseWriter, r *http.Request) (obj T, err error) {
	if r.Method == "GET" {
		err = errors.New("trying to get body of get request")
		log.ShowErr(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	jsonData, err := io.ReadAll(r.Body)
	if err != nil {
		log.ShowErr(err)
		writeJson(w, http.StatusInternalServerError, map[string]any{"error": "Could not read request body"})
		return
	}
	err = json.Unmarshal(jsonData, &obj)
	if err != nil {
		log.ShowErr(err)
		writeJson(w, http.StatusBadRequest, map[string]any{"error": "Request body is not in expected JSON format"})
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

func getUserFromCtx(w http.ResponseWriter, r *http.Request) (*models.User, error) {
	userI := r.Context().Value(UserKey)
	if userI == nil {
		return nil, werror.ErrCtxMissingUser
	}

	u, _ := userI.(*models.User)

	if u.IsPublic() && r.Context().Value(AllowPublicKey) == nil && r.Context().Value(ServerKey) == nil {
		return nil, werror.ErrNoPublicUser
	}
	return u, nil
}

func getInstanceFromCtx(r *http.Request) *models.Instance {
	serverI := r.Context().Value(ServerKey)
	if serverI == nil {
		return nil
	}

	srv, _ := serverI.(*models.Instance)
	return srv
}

func getShareFromCtx[T models.Share](w http.ResponseWriter, r *http.Request) (T, error) {
	pack := getServices(r)

	shareId := models.ShareId(r.URL.Query().Get("shareId"))
	if shareId == "" {
		shareId = models.ShareId(chi.URLParam(r, "shareId"))
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
	writeJson(w, http.StatusNotFound, map[string]any{"error": err.Error()})
	return empty, err
}

type FileStat struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	IsDir   bool      `json:"isDir"`
	ModTime time.Time `json:"modifyTimestamp"`
	Exists  bool      `json:"exists"`
}
