package net

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	context_mod "github.com/ethanrous/weblens/modules/context"
)

type Error struct {
	Error string `json:"error"`
}

func ReadError(ctx context.Context, resp *http.Response, respErr error) error {
	defer resp.Body.Close()

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
