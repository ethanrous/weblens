package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func writeJson(w http.ResponseWriter, status int, obj any) {
	bs, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	w.Header().Set("content-type", "application/json")
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

// SafeErrorAndExit reads the given error, and if not nil will write to the
// response writer the correct http code and json error response. It returns
// true if there is an error and the http request should be terminated, and
// false if the error is nil
func SafeErrorAndExit(err error, w http.ResponseWriter, log ...*zerolog.Logger) (shouldExit bool) {
	if err == nil {
		return false
	}

	safe, code := werror.GetSafeErr(err)
	writeError(w, code, safe)
	if len(log) > 0 && log[0] != nil {
		log[0].Error().CallerSkipFrame(1).Stack().Err(err).Msg("")
	}
	return true
}

// readCtxBody reads the body of a gin context and unmarshal it into the given generic type.
// It returns the unmarshalled object or an error if one occurred. It also sets the response status
// in the context and logs the error if an error occurred so it is recommended, upon reading an error from this function,
// return from a http handler immediately.
func readCtxBody[T any](w http.ResponseWriter, r *http.Request) (obj T, err error) {
	if r.Method == "GET" {
		err = errors.New("trying to get body of get request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	jsonData, err := io.ReadAll(r.Body)
	if err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]any{"error": "Could not read request body"})
		return
	}
	err = json.Unmarshal(jsonData, &obj)
	if err != nil {
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
		log.Warn().Msg("Reading body with unknown content length")
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
		log.Warn().Msg("Reading body with unknown content length")
		bodyB, err = io.ReadAll(resp.Body)
	} else {
		bodyB, err = internal.OracleReader(resp.Body, resp.ContentLength)
	}
	return
}

func getUserFromCtx(r *http.Request, allowPublic bool) (*models.User, error) {
	userI := r.Context().Value(UserKey)
	if userI == nil {
		return nil, werror.ErrCtxMissingUser
	}

	u, _ := userI.(*models.User)

	if u.IsPublic() && (!allowPublic && (r.Context().Value(AllowPublicKey) == nil && r.Context().Value(ServerKey) == nil)) {
		return nil, werror.WithStack(werror.ErrNoPublicUser)
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
		pack.Log.Debug().Func(func(e *zerolog.Event) { e.Msgf("Got share [%s]", tsh.ID()) })
		return tsh, nil
	}

	err := werror.ErrNoShare
	writeError(w, http.StatusNotFound, err)
	return empty, err
}

type FileStat struct {
	ModTime time.Time `json:"modifyTimestamp"`
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	IsDir   bool      `json:"isDir"`
	Exists  bool      `json:"exists"`
}
