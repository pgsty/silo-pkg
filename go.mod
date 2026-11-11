module github.com/pgsty/silo-pkg/v3

// Library consumers and maintained builds use the same Go release.
go 1.27.4

// v22.7.0 does not compile on NetBSD because its unix implementation uses
// CLOCK_MONOTONIC, which is unavailable there. Keep the last portable release
// until go-systemd ships the upstream fix.
replace github.com/coreos/go-systemd/v22 => github.com/coreos/go-systemd/v22 v22.6.0

require (
	github.com/cheggaaa/pb v1.0.30
	github.com/coreos/go-oidc/v3 v3.21.0
	github.com/dustin/go-humanize v1.1.0
	github.com/fatih/color v1.19.0
	github.com/fatih/structs v1.1.0
	github.com/go-ldap/ldap/v3 v3.4.15
	github.com/go-openapi/swag/conv v0.29.2
	github.com/golang-jwt/jwt/v4 v4.5.2
	github.com/lestrrat-go/jwx/v3 v3.4.0
	github.com/mattn/go-colorable v0.1.16
	github.com/mattn/go-isatty v0.0.24
	github.com/minio/minio-go/v7 v7.4.2-0.20261025165926-320af050c445
	github.com/minio/mux v1.10.1
	github.com/rjeczalik/notify v0.9.3
	github.com/tinylib/msgp v1.6.5
	github.com/zeebo/xxh3 v1.1.0
	go.etcd.io/etcd/client/v3 v3.7.3
	go.yaml.in/yaml/v3 v3.0.6
	golang.org/x/crypto v0.58.1-0.20261009181718-2f1f834330b2
	golang.org/x/oauth2 v0.38.0
	golang.org/x/sys v0.50.0
)

require (
	github.com/Azure/go-ntlmssp v0.1.1 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.7.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.1 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.8 // indirect
	github.com/go-jose/go-jose/v4 v4.1.9 // indirect
	github.com/goccy/go-json v0.11.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.33.0 // indirect
	github.com/klauspost/cpuid/v2 v2.4.2 // indirect
	github.com/lestrrat-go/blackmagic v1.0.4 // indirect
	github.com/lestrrat-go/dsig v1.4.0 // indirect
	github.com/lestrrat-go/dsig-secp256k1 v1.0.0 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc/v3 v3.0.6 // indirect
	github.com/lestrrat-go/option/v2 v2.0.0 // indirect
	// v0.0.31 changes existing CJK Latin and grapheme widths in CLI output.
	github.com/mattn/go-runewidth v0.0.29 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/philhofer/fwd v1.2.1 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/valyala/fastjson v1.6.10 // indirect
	go.etcd.io/etcd/api/v3 v3.7.3 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.7.3 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.28.0 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/term v0.48.0 // indirect
	golang.org/x/text v0.44.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20261026165753-2ffab32bce7e // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261019180522-ce908d29299b // indirect
	google.golang.org/grpc v1.85.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/ini.v1 v1.67.3 // indirect
)
