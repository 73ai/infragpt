package identitysvc_test

import (
	"github.com/priyanshujain/infragpt/services/infragpt/identity/identitytest"
	"github.com/priyanshujain/infragpt/services/infragpt/internal/identitysvc/identitysvctest"

	"testing"
)

func TestService(t *testing.T) {
	svc, externalActions := identitysvctest.NewServiceWithExternalActions(t)
	identitytest.Ensure(t, svc, externalActions)
}
