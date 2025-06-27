module github.com/priyanshujain/infragpt/services/infragpt

go 1.24.0

replace github.com/priyanshujain/infragpt/services/agent/src/client/go => ../agent/src/client/go

require (
	github.com/clerk/clerk-sdk-go/v2 v2.3.1
	github.com/golang-jwt/jwt/v4 v4.5.2
	github.com/google/go-github/v69 v69.2.0
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	github.com/mitchellh/mapstructure v1.5.0
	github.com/priyanshujain/infragpt/services/agent/src/client/go v0.0.0-00010101000000-000000000000
	github.com/slack-go/slack v0.16.0
	github.com/sqlc-dev/pqtype v0.3.0
	github.com/svix/svix-webhooks v1.67.0
	golang.org/x/sync v0.14.0
	google.golang.org/grpc v1.72.2
	google.golang.org/protobuf v1.36.6
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/go-jose/go-jose/v3 v3.0.4 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)
