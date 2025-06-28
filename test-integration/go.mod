module test-integration

go 1.24.3

replace github.com/priyanshujain/infragpt/services/infragpt => ../services/infragpt

require github.com/priyanshujain/infragpt/services/infragpt v0.0.0-00010101000000-000000000000

require (
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
)
