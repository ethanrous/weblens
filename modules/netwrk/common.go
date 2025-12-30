// Package netwrk provides HTTP networking utilities for making requests and handling errors.
package netwrk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	context_mod "github.com/ethanrous/weblens/modules/wlcontext"
)

// Error represents an error response from an HTTP request.
type Error struct {
	Error string `json:"error"`
}

// ReadError reads and parses an error from an HTTP response body.
func ReadError(ctx context.Context, resp *http.Response, respErr error) error {
	defer resp.Body.Close() //nolint:errcheck

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		context_mod.ToZ(ctx).Log().Error().Err(err).Msg("Failed to read response body error")

		return respErr
	}

	target := &Error{}

	err = json.Unmarshal(bs, &target)
	if err != nil {
		context_mod.ToZ(ctx).Log().Error().Err(err).Msg("Failed to read response body error")

		return respErr
	}

	return fmt.Errorf("%w: %s", respErr, target.Error)
}
