package cloud

/*
import (
	"fmt"

	"github.com/company/infragpt"
	"github.com/company/infragpt/internal/infragptsvc/domain"
	"github.com/company/infragpt/internal/infragptsvc/supporting/cloud/gcp"
)

// Factory implements domain.CloudFactory
type Factory struct {
	gcpCredentialsFile string
	// Add other provider credentials as needed
}

// NewFactory creates a new cloud factory
func NewFactory(gcpCredentialsFile string) *Factory {
	return &Factory{
		gcpCredentialsFile: gcpCredentialsFile,
	}
}

// GetService returns a cloud service for the specified provider
func (f *Factory) GetService(provider infragpt.CloudProvider) (domain.CloudService, error) {
	switch provider {
	case infragpt.ProviderGCP:
		return gcp.NewGCPService(f.gcpCredentialsFile)
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", provider)
	}
}

*/
