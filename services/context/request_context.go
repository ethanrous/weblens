package context

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/ethanrous/weblens/models/client"
	share_model "github.com/ethanrous/weblens/models/share"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/crypto"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/modules/net"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
)

const BaseContextKey = "context"

var _ context_mod.ContextZ = RequestContext{}

type requestContextKey struct{}

type RequestContext struct {
	AppContext

	Req    *http.Request
	ReqCtx context.Context
	W      http.ResponseWriter

	Requester  *user_model.User
	Remote     tower_model.Instance
	IsLoggedIn bool

	Share *share_model.FileShare

	mongoSession mongo.SessionContext
}

func (c RequestContext) GetMongoSession() mongo.SessionContext {
	return c.mongoSession
}

func (c RequestContext) AppCtx() context_mod.ContextZ {
	return c.AppContext
}

func (c RequestContext) SetValue(key any, value any) {
	// c.Req = c.Req.WithContext(context.WithValue(c.Req.Context(), key, value))
	c.ReqCtx = context.WithValue(c.ReqCtx, key, value)
}

func (c RequestContext) Value(key any) any {
	if key == (requestContextKey{}) {
		return c
	}

	if c.ReqCtx == nil {
		panic("request context is nil")
	}

	if v := c.AppContext.Value(key); v != nil {
		return v
	}

	return c.ReqCtx.Value(key)
}

func (c RequestContext) WithContext(ctx context.Context) context.Context {
	l, ok := log.FromContextOk(ctx)
	if !ok {
		l = log.FromContext(c)
	}

	c.BasicContext = NewBasicContext(ctx, l)

	return c
}

// Path returns the value of a URL parameter, or an empty string if the parameter is not found.
func (c RequestContext) Path(paramName string) string {
	q, err := url.QueryUnescape(chi.URLParam(c.Req, paramName))
	if err != nil {
		c.Log().Error().Err(err).Msgf("Failed to unescape URL parameter '%s'", paramName)

		return ""
	}

	return q
}

// Query returns the value of a query parameter, or an empty string if the parameter is not found.
func (c RequestContext) Query(paramName string) string {
	return c.Req.URL.Query().Get(paramName)
}

func (c RequestContext) QueryBool(paramName string) bool {
	qstr := c.Req.URL.Query().Get(paramName)
	if qstr == "" {
		return false
	}

	q, err := strconv.ParseBool(qstr)
	if err != nil {
		c.Log().Error().Err(err).Msgf("Failed to parse query parameter '%s' as bool", paramName)

		return false
	}

	return q
}

func (c RequestContext) Error(code int, err error) {
	if err == nil {
		err = errors.New("error is nil")
		c.Log().Error().Stack().Err(err).Msg("")
		c.W.WriteHeader(code)

		return
	}

	code, errMsg := errors.AsStatus(err, code)

	var e *zerolog.Event
	if code >= http.StatusInternalServerError {
		e = c.Log().Error().Stack()
	} else {
		e = c.Log().Warn()
	}

	e.CallerSkipFrame(1).Caller().Err(err).Msgf("API Error %d %s", code, http.StatusText(code))

	c.JSON(code, net.Error{Error: errMsg})
}

func (c RequestContext) ExpireCookie() {
	cookie := fmt.Sprintf("%s=;Path=/;Expires=Thu, 01 Jan 1970 00:00:00 GMT;HttpOnly", crypto.SessionTokenCookie)
	c.W.Header().Set("Set-Cookie", cookie)
}

func (c RequestContext) GetCookie(cookieName string) (string, error) {
	// Get the value of a specific cookie from the request.
	// This will return an empty string and non-nil error if the cookie is not present.
	cookie, err := c.Req.Cookie(cookieName)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

func (c RequestContext) Header(headerName string) string {
	// Get the value of a specific header from the request.
	// This will return an empty string if the header is not present.
	headerValue := c.Req.Header.Get(headerName)

	return headerValue
}

func (c RequestContext) SetHeader(headerName, headerValue string) {
	c.W.Header().Set(headerName, headerValue)
}

func (c RequestContext) AddHeader(headerName, headerValue string) {
	c.W.Header().Add(headerName, headerValue)
}

// Set the HTTP status code for the response.
func (c RequestContext) Status(code int) {
	if code >= http.StatusBadRequest {
		c.Log().Trace().CallerSkipFrame(1).Caller().Msgf("Setting response code [%d]", code)
	}

	c.W.WriteHeader(code)
}

var rangeMatchR = regexp.MustCompile("^bytes=[0-9]+-[0-9]+/[0-9]+$")

func (c RequestContext) ContentRange() (start, end, total int, err error) {
	// Get the "Range" header from the request.
	rangeHeader := c.Header("Content-Range")

	// If the Range header is empty or not in the expected format, return an error.
	if rangeHeader == "" {
		err = errors.New("Range header not provided")

		return
	}

	if !rangeMatchR.MatchString(rangeHeader) {
		err = errors.New("Invalid Range header format, must match 'bytes=start-end/total'")

		return
	}

	// Parse the range header to extract start, end, and total values.
	_, err = fmt.Sscanf(rangeHeader, "bytes=%d-%d/%d", &start, &end, &total)
	if err != nil {
		err = errors.WithStack(err)

		return
	}

	return start, end, total, nil
}

func (c RequestContext) JSON(code int, data any) {
	bs, err := json.Marshal(data)
	if err != nil {
		c.Error(http.StatusInternalServerError, errors.WithStack(err))

		return
	}

	c.SetHeader("Content-Type", "application/json")
	c.Bytes(code, bs)
}

func (c RequestContext) Bytes(code int, data []byte) {
	c.Status(code)
	_, err := c.W.Write(data)

	// If the write fails, log the error, but don't send a response.
	if err != nil {
		c.Log().Error().Stack().Err(err).Msg("Failed to write response to http request")
	}
}

func (c RequestContext) Client() *client.WsClient {
	return c.ClientService.GetClientByUsername(c.Requester.Username)
}

func (c RequestContext) AttemptGetUsername() string {
	if c.Requester != nil && c.Requester.Username != "" && c.Requester.Username != user_model.PublicUserName {
		return c.Requester.Username
	}

	usernameCookie, err := c.GetCookie(crypto.UserCrumbCookie)
	if err != nil {
		return ""
	}

	return usernameCookie
}

func ReqFromContext(ctx context.Context) (RequestContext, bool) {
	if ctx == nil {
		return RequestContext{}, false
	}

	reqCtx, ok := ctx.Value(requestContextKey{}).(RequestContext)
	if !ok {
		return RequestContext{}, false
	}

	return reqCtx, true
}

func AppContexter(ctx AppContext) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqContext := RequestContext{
				AppContext: ctx,
				Req:        r,
				ReqCtx:     r.Context(),
				W:          w,
			}

			reqContext.SetValue(requestContextKey{}, reqContext)
			reqContext.SetValue("towerId", ctx.LocalTowerId)
			reqContext.Req = reqContext.Req.WithContext(reqContext)
			next.ServeHTTP(reqContext.W, reqContext.Req)
		})
	}
}

func TimestampFromCtx(ctx RequestContext) (time.Time, bool, error) {
	ts := ctx.Query("timestamp")
	if ts == "" || ts == "0" {
		return time.Time{}, false, nil
	}

	millis, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return time.Time{}, false, errors.WithStack(err)
	}

	return time.UnixMilli(millis), true, nil
}
