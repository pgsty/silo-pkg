module github.com/minio/pkg/v3

// Keep one prior Go release available to library consumers while building and
// testing the maintained branch with the current toolchain.
go 1.26.0

toolchain go1.27.0

// v22.7.0 does not compile on NetBSD because its unix implementation uses
// CLOCK_MONOTONIC, which is unavailable there. Keep the last portable release
// until go-systemd ships the upstream fix.
replace github.com/coreos/go-systemd/v22 => github.com/coreos/go-systemd/v22 v22.6.0

require (
	github.com/cheggaaa/pb v1.0.30
	github.com/coreos/go-oidc/v3 v3.17.0
	github.com/fatih/color v1.19.0
	github.com/fatih/structs v1.1.0
	github.com/go-ldap/ldap/v3 v3.4.14
	github.com/go-openapi/swag/conv v0.28.0
	github.com/golang-jwt/jwt/v4 v4.5.2
	github.com/lestrrat-go/jwx/v3 v3.0.13
	github.com/mattn/go-colorable v0.1.15
	github.com/mattn/go-isatty v0.0.24
	github.com/minio/minio-go/v7 v7.0.99
	github.com/minio/mux v1.9.2
	github.com/rjeczalik/notify v0.9.3
	github.com/tinylib/msgp v1.6.4
	github.com/zeebo/xxh3 v1.1.0
	go.etcd.io/etcd/client/v3 v3.7.1
	go.yaml.in/yaml/v3 v3.0.5
	golang.org/x/crypto v0.55.0
	golang.org/x/oauth2 v0.36.0
	golang.org/x/sys v0.47.0
)

require (
	github.com/clipperhouse/stringish v0.1.1 // indirect
	github.com/clipperhouse/uax29/v2 v2.3.1 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/lestrrat-go/dsig v1.0.0 // indirect
	github.com/lestrrat-go/dsig-secp256k1 v1.0.0 // indirect
	github.com/lestrrat-go/httprc/v3 v3.0.6 // indirect
	github.com/lestrrat-go/option/v2 v2.0.0 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/valyala/fastjson v1.6.10 // indirect
)

require (
	github.com/Azure/go-ntlmssp v0.1.1 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.7.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.1 // indirect
	github.com/dustin/go-humanize v1.0.1
	github.com/go-asn1-ber/asn1-ber v1.5.8 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/lestrrat-go/blackmagic v1.0.4 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.19 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	go.etcd.io/etcd/api/v3 v3.7.1 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.7.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.28.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260414002931-afd174a4e478 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260414002931-afd174a4e478 // indirect
	google.golang.org/grpc v1.82.1 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
