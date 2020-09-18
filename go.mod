module github.com/zehome/sintls

require (
	github.com/Azure/azure-sdk-for-go v36.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest/adal v0.8.0 // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/blang/semver v3.5.1+incompatible
	github.com/cheynewallace/tabby v1.1.0
	github.com/cloudflare/cloudflare-go v0.10.8 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/exoscale/egoscale v1.19.0 // indirect
	github.com/go-acme/lego/v3 v3.9.0
	github.com/go-pg/migrations/v7 v7.1.11
	github.com/go-pg/pg/v9 v9.2.0
	github.com/gophercloud/gophercloud v0.7.0 // indirect
	github.com/hashicorp/go-hclog v0.10.0 // indirect
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/labstack/echo/v4 v4.1.17
	github.com/linode/linodego v0.12.2 // indirect
	github.com/liquidweb/liquidweb-go v1.6.1 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/miekg/dns v1.1.31
	github.com/oracle/oci-go-sdk v13.0.0+incompatible // indirect
	github.com/ovh/go-ovh v1.1.0
	github.com/rhysd/go-github-selfupdate v1.2.2
	github.com/sacloud/libsacloud v1.32.0 // indirect
	github.com/ulikunitz/xz v0.5.6 // indirect
	github.com/urfave/cli v1.22.4
	go.uber.org/ratelimit v0.1.0 // indirect
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	golang.org/x/net v0.0.0-20200904194848-62affa334b73
	gopkg.in/AlecAivazis/survey.v1 v1.8.8
	gopkg.in/ns1/ns1-go.v2 v2.0.0-20191126161805-25b9eac84517 // indirect
	gopkg.in/square/go-jose.v2 v2.4.0 // indirect
)

// broken go.mod inside a package needing this
replace github.com/h2non/gock => gopkg.in/h2non/gock.v1 v1.0.14

// Otherwise, the build will fail in dh_golang
//replace google.golang.org/api => google.golang.org/api v0.3.0

go 1.13
