package netwrk

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ethanrous/weblens/modules/wlerrors"
)

// ReadRequestBody reads the body of a http request and unmarshal it into the given generic type.
// It returns the unmarshalled object or an error if one occurred.
func ReadRequestBody[T any](r *http.Request) (obj T, err error) {
	if r.Method == http.MethodGet {
		err = wlerrors.New("trying to get body of get request")

		return
	}

	jsonData, err := io.ReadAll(r.Body)
	if err != nil {
		return obj, wlerrors.Wrap(err, "could not read request body")
	}

	err = json.Unmarshal(jsonData, &obj)
	if err != nil {
		return obj, wlerrors.Wrap(err, "could not unmarshal request body")
	}

	return
}
