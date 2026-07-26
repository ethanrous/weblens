package wlcontext_test

import (
	"strings"
	"testing"

	"github.com/ethanrous/weblens/modules/wlcontext"
	"github.com/stretchr/testify/assert"
)

func TestCanceledError_DoesNotClaimDBError(t *testing.T) {
	err := wlcontext.NewCanceledError("some operation")

	assert.NotContains(t, err.Error(), "DB error")
	// db.WrapError classifies errors as database cancellations by this suffix;
	// CanceledError living outside the db package must not collide with it.
	assert.False(t, strings.HasSuffix(err.Error(), "context canceled"))
}
