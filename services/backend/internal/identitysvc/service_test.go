package identitysvc_test

import (
	"testing"

	"github.com/priyanshujain/infragpt/services/backend/identitytest"
	"github.com/priyanshujain/infragpt/services/backend/internal/identitysvc/identitysvctest"
)

func TestService(t *testing.T) {
	identitytest.Ensure(t, identitysvctest.NewConfig().Fixture())
}
