package context

import (
	"encoding/json"
	"fmt"
	"net/http"

	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/crypto"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

const BaseContextKey = "context"

type RequestContext struct {
	BasicContext
	Req        *http.Request
	W          http.ResponseWriter
	DB         *mongo.Database
	Requester  *user_model.User
	IsLoggedIn bool

	// TowerId is the id of the tower that the request is being handled on
	TowerId string
}

func GetFromHTTP(r *http.Request) RequestContext {
	ctx, _ := r.Context().(RequestContext)
	return ctx
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
	c.Log.Error().Stack().Err(err).Msg("API Error")

	c.JSON(code, map[string]string{"error": err.Error()})
}

func (c *RequestContext) ExpireCookie() {
	cookie := fmt.Sprintf("%s=;Path=/;Expires=Thu, 01 Jan 1970 00:00:00 GMT;HttpOnly", crypto.SessionTokenCookie)
	c.W.Header().Set("Set-Cookie", cookie)
}

func (c *RequestContext) JSON(code int, data any) {
	bs, err := json.Marshal(data)
	if err != nil {
		err = errors.WithStack(err)
		c.Log.Error().Stack().Err(err).Msg("Failed to marshal JSON")

		c.W.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.W.WriteHeader(code)
	c.W.Write(bs)
}
