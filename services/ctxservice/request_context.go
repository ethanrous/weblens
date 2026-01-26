package ctxservice

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
	"github.com/ethanrous/weblens/modules/cryptography"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/modules/netwrk"
	context_mod "github.com/ethanrous/weblens/modules/wlcontext"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
)

// BaseContextKey is the context key for accessing the base context.
const BaseContextKey = "context"

var _ context_mod.Z = RequestContext{}

type requestContextKey struct{}

// RequestContext represents an HTTP request context with user authentication and request-specific data.
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

// GetMongoSession retrieves the MongoDB session context for this request.
func (c RequestContext) GetMongoSession() mongo.SessionContext {
	return c.mongoSession
}

// AppCtx returns the underlying AppContext from the RequestContext.
func (c RequestContext) AppCtx() context_mod.Z {
	return c.AppContext
}

// SetValue adds a key-value pair to the request context.
func (c RequestContext) SetValue(key any, value any) {
	// c.Req = c.Req.WithContext(context.WithValue(c.Req.Context(), key, value))
	c.ReqCtx = context.WithValue(c.ReqCtx, key, value)
	c.AppContext = c.AppContext.WithValue(key, value)
}

// Value retrieves the value associated with the given key from the RequestContext.
func (c RequestContext) Value(key any) any {
	if key == (requestContextKey{}) {
		return c
	}

	if c.ReqCtx == nil {
		panic("request context is nil")
	}

	if key == context_mod.RequestDoerKey {
		return c.Doer()
	}

	if v := c.AppContext.Value(key); v != nil {
		return v
	}

	return c.ReqCtx.Value(key)
}

// WithContext creates a new context by combining the RequestContext with the provided context.
func (c RequestContext) WithContext(ctx context.Context) context.Context {
	l, ok := log.FromContextOk(ctx)
	if !ok {
		l = log.FromContext(c)
	}

	c.BasicContext = NewBasicContext(ctx, l)

	return c
}

// WithValue returns a copy of RequestContext with the specified key-value pair added.
func (c RequestContext) WithValue(key, value any) RequestContext {
	c.AppContext = c.AppContext.WithValue(key, value)

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

// QueryBool retrieves a query parameter and parses it as a boolean value.
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

// QueryInt retrieves a query parameter and parses it as an integer value.
func (c RequestContext) QueryInt(paramName string) (int64, error) {
	queryStr := c.Query(paramName)
	if queryStr == "" {
		return 0, nil
	}

	num, err := strconv.ParseInt(queryStr, 10, 32)
	if err != nil {
		return 0, wlerrors.New("Invalid query int")
	}

	return num, nil
}

// QueryIntDefault retrieves a query parameter as an integer or returns the default value if not provided.
func (c RequestContext) QueryIntDefault(paramName string, defaultValue int64) (int64, error) {
	queryStr := c.Query(paramName)
	if queryStr == "" {
		return defaultValue, nil
	}

	return c.QueryInt(paramName)
}

func (c RequestContext) Error(code int, err error) {
	if err == nil {
		err = wlerrors.New("error is nil")
		c.Log().Error().Stack().Err(err).Msg("")
		c.W.WriteHeader(code)

		return
	}

	code, errMsg := wlerrors.AsStatus(err, code)

	var e *zerolog.Event
	if code >= http.StatusInternalServerError {
		e = c.Log().Error().Stack()
	} else {
		e = c.Log().Warn()

		if c.Log().GetLevel() <= zerolog.DebugLevel {
			e = e.Stack()
		}
	}

	e.CallerSkipFrame(1).Caller().Err(err).Msgf("API Error %d %s", code, http.StatusText(code))

	c.JSON(code, netwrk.Error{Error: errMsg})
}

// ExpireCookie sets a cookie header that expires the session cookie.
func (c RequestContext) ExpireCookie() {
	cookie := fmt.Sprintf("%s=;Path=/;Expires=Thu, 01 Jan 1970 00:00:00 GMT;HttpOnly", cryptography.SessionTokenCookie)
	c.W.Header().Set("Set-Cookie", cookie)
}

// GetCookie returns the value of a cookie by name.
func (c RequestContext) GetCookie(cookieName string) (string, error) {
	// Get the value of a specific cookie from the request.
	// This will return an empty string and non-nil error if the cookie is not present.
	cookie, err := c.Req.Cookie(cookieName)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

// Header returns the value of a request header.
func (c RequestContext) Header(headerName string) string {
	// Get the value of a specific header from the request.
	// This will return an empty string if the header is not present.
	headerValue := c.Req.Header.Get(headerName)

	return headerValue
}

// SetHeader sets a response header.
func (c RequestContext) SetHeader(headerName, headerValue string) {
	c.W.Header().Set(headerName, headerValue)
}

// SetContentType sets the Content-Type header for the response.
func (c RequestContext) SetContentType(contentType string) {
	c.SetHeader("Content-Type", contentType)
}

// SetLastModified sets the Last-Modified header for the response.
func (c RequestContext) SetLastModified(modified time.Time) {
	c.SetHeader("Last-Modified", modified.UTC().Format(http.TimeFormat))
}

// IfModifiedSince checks if the request's If-Modified-Since header is set and returns the time if it is
func (c RequestContext) IfModifiedSince() (time.Time, bool) {
	ifModifiedSince := c.Req.Header.Get("If-Modified-Since")
	if ifModifiedSince == "" {
		return time.Time{}, false
	}

	modifiedTime, err := time.Parse(http.TimeFormat, ifModifiedSince)
	if err != nil {
		c.Log().Error().Stack().Err(err).Msgf("Failed to parse If-Modified-Since header: %s", ifModifiedSince)

		return time.Time{}, false
	}

	return modifiedTime, true
}

// AddHeader adds a value to a response header.
func (c RequestContext) AddHeader(headerName, headerValue string) {
	c.W.Header().Add(headerName, headerValue)
}

// Status sets the HTTP status code for the response.
func (c RequestContext) Status(code int) {
	if code >= http.StatusBadRequest {
		c.Log().Trace().CallerSkipFrame(1).Caller().Msgf("Setting response code [%d]", code)
	}

	c.W.WriteHeader(code)
}

var rangeMatchR = regexp.MustCompile("^bytes=[0-9]+-[0-9]+/[0-9]+$")

// ContentRange parses the Content-Range header and returns start, end, and total values.
func (c RequestContext) ContentRange() (start, end, total int, err error) {
	// Get the "Range" header from the request.
	rangeHeader := c.Header("Content-Range")

	// If the Range header is empty or not in the expected format, return an error.
	if rangeHeader == "" {
		err = wlerrors.New("Range header not provided")

		return
	}

	if !rangeMatchR.MatchString(rangeHeader) {
		err = wlerrors.New("Invalid Range header format, must match 'bytes=start-end/total'")

		return
	}

	// Parse the range header to extract start, end, and total values.
	_, err = fmt.Sscanf(rangeHeader, "bytes=%d-%d/%d", &start, &end, &total)
	if err != nil {
		err = wlerrors.WithStack(err)

		return
	}

	return start, end, total, nil
}

// JSON writes a JSON response with the given status code.
func (c RequestContext) JSON(code int, data any) {
	bs, err := json.Marshal(data)
	if err != nil {
		c.Error(http.StatusInternalServerError, wlerrors.WithStack(err))

		return
	}

	c.SetHeader("Content-Type", "application/json")
	c.Bytes(code, bs)
}

// Bytes writes a byte slice response with the given status code.
func (c RequestContext) Bytes(code int, data []byte) {
	c.Status(code)
	_, err := c.W.Write(data)

	// If the write fails, log the error, but don't send a response.
	if err != nil {
		c.Log().Error().Stack().Err(err).Msg("Failed to write response to http request")
	}
}

// Client returns the websocket client for the current user.
func (c RequestContext) Client() *client.WsClient {
	return c.ClientService.GetClientByUsername(c.Requester.Username)
}

func (c RequestContext) Write(b []byte) (int, error) {
	return c.W.Write(b)
}

// AttemptGetUsername attempts to retrieve the username from the requester or cookies.
func (c RequestContext) AttemptGetUsername() string {
	if c.Requester != nil && c.Requester.Username != "" && c.Requester.Username != user_model.PublicUserName {
		return c.Requester.Username
	}

	usernameCookie, err := c.GetCookie(cryptography.UserCrumbCookie)
	if err != nil {
		return ""
	}

	return usernameCookie
}

// ReqFromContext extracts a RequestContext from a context.Context.
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

// AppContexter creates a middleware that wraps handlers with an AppContext.
func AppContexter(ctx AppContext) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Make copy of the global logger and attach to a LOCAL copy of ctx.
			// Using a local variable avoids racing with other requests that share
			// the captured ctx parameter.
			logger := *log.GlobalLogger()
			localCtx := ctx.ReplaceLogger(&logger)

			reqContext := RequestContext{
				AppContext: localCtx,
				Req:        r,
				ReqCtx:     r.Context(),
				W:          w,
			}

			reqContext.SetValue(requestContextKey{}, reqContext)
			reqContext.SetValue("towerID", localCtx.LocalTowerID)
			reqContext.Req = reqContext.Req.WithContext(reqContext)
			next.ServeHTTP(reqContext.W, reqContext.Req)
		})
	}
}

// TimestampFromCtx extracts a timestamp from the request context query parameters.
func TimestampFromCtx(ctx RequestContext) (time.Time, bool, error) {
	ts := ctx.Query("timestamp")
	if ts == "" || ts == "0" {
		return time.Time{}, false, nil
	}

	millis, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return time.Time{}, false, wlerrors.WithStack(err)
	}

	return time.UnixMilli(millis), true, nil
}

// WithRequester sets the requesting user in the RequestContext.
func (c RequestContext) WithRequester(u *user_model.User) RequestContext {
	c.Requester = u
	c.WithValue(context_mod.RequestDoerKey, u.GetUsername())

	if u != nil && u.Username != "" && u.Username != user_model.PublicUserName {
		c.IsLoggedIn = true
	}

	return c
}

// Doer returns the username of the requester, or a default unknown user if not set.
func (c RequestContext) Doer() *user_model.User {
	if c.Requester != nil {
		return c.Requester
	}

	return user_model.GetUnknownUser()
}
