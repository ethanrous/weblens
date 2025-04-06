package context

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/ethanrous/weblens/models/client"
	share_model "github.com/ethanrous/weblens/models/share"
	tower_model "github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/crypto"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

const BaseContextKey = "context"

type RequestContext struct {
	AppContext

	Req *http.Request
	W   http.ResponseWriter

	Requester  *user_model.User
	Remote     *tower_model.Instance
	IsLoggedIn bool

	Share *share_model.FileShare
}

// Path returns the value of a URL parameter, or an empty string if the parameter is not found.
func (c *RequestContext) Path(paramName string) string {
	return chi.URLParam(c.Req, paramName)
}

// Query returns the value of a query parameter, or an empty string if the parameter is not found.
func (c *RequestContext) Query(paramName string) string {
	return c.Req.URL.Query().Get(paramName)
}

func (c *RequestContext) Error(code int, err error) {
	err = errors.WithStack(err)
	c.Logger.Error().Stack().Err(err).Msg("API Error")

	c.JSON(code, map[string]string{"error": err.Error()})
}

func (c *RequestContext) ExpireCookie() {
	cookie := fmt.Sprintf("%s=;Path=/;Expires=Thu, 01 Jan 1970 00:00:00 GMT;HttpOnly", crypto.SessionTokenCookie)
	c.W.Header().Set("Set-Cookie", cookie)
}

func (c *RequestContext) Header(headerName string) string {
	// Get the value of a specific header from the request.
	// This will return an empty string if the header is not present.
	headerValue := c.Req.Header.Get(headerName)

	// If you want to log or process the header value, you can do it here.
	if headerValue == "" {
		c.Logger.Trace().Msgf("Header '%s' not found", headerName)
	} else {
		c.Logger.Trace().Msgf("Header '%s' found with value: %s", headerName, headerValue)
	}

	return headerValue
}

var rangeMatchR = regexp.MustCompile("^bytes=[0-9]+-[0-9]+/[0-9]+$")

func (c *RequestContext) ContentRange() (start, end, total int, err error) {
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

func (c *RequestContext) JSON(code int, data any) {
	bs, err := json.Marshal(data)
	if err != nil {
		err = errors.WithStack(err)
		c.Logger.Error().Stack().Err(err).Msg("Failed to marshal JSON")

		c.W.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.W.WriteHeader(code)
	c.W.Write(bs)
}

func (c *RequestContext) Client() *client.WsClient {
	return c.ClientService.GetClientByUsername(c.Requester.Username)
}
