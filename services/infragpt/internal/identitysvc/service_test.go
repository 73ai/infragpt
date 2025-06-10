package identitysvc_test

import (
	"testing"

	"github.com/priyanshujain/infragpt/services/infragpt/identitytest"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/identitysvctest"
)

func TestService(t *testing.T) {
	identitytest.Ensure(t, identitysvctest.NewConfig().Fixture())
}
