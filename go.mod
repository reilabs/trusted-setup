module github.com/reilabs/trusted-setup

go 1.24.0

toolchain go1.24.4

require (
	github.com/aws/aws-sdk-go-v2 v1.38.2
	github.com/aws/aws-sdk-go-v2/config v1.31.4
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.19.2
	github.com/aws/aws-sdk-go-v2/service/s3 v1.87.2
	github.com/consensys/gnark v0.14.0
	github.com/consensys/gnark-crypto v0.19.0
	github.com/drand/go-clients v0.2.3
	github.com/johannesboyne/gofakes3 v0.0.0-20250825084532-6555d310c473
	github.com/rs/zerolog v1.34.0
	github.com/stretchr/testify v1.10.0
	github.com/urfave/cli/v3 v3.3.8
	github.com/worldcoin/ptau-deserializer v0.2.0
	google.golang.org/grpc v1.73.0
	google.golang.org/protobuf v1.36.6
)

replace github.com/worldcoin/ptau-deserializer => github.com/reilabs/ptau-deserializer v0.0.0-20250630133456-6f3242b72b0a

require (
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.18.8 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.8.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.28.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.34.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.38.1 // indirect
	github.com/aws/smithy-go v1.23.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.24.0 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/drand/drand/v2 v2.1.3 // indirect
	github.com/drand/kyber v1.3.1 // indirect
	github.com/drand/kyber-bls12381 v0.3.3 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/google/pprof v0.0.0-20250820193118-f64d9cf942d6 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/ingonyama-zk/icicle-gnark/v3 v3.2.2 // indirect
	github.com/kilic/bls12-381 v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nikkolasg/hexjson v0.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.22.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.65.0 // indirect
	github.com/prometheus/procfs v0.17.0 // indirect
	github.com/ronanh/intcomp v1.1.1 // indirect
	github.com/ryszard/goskiplist v0.0.0-20150312221310-2dfbae5fcf46 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.dedis.ch/fixbuf v1.0.3 // indirect
	go.shabbyrobe.org/gocovmerge v0.0.0-20230507111327-fa4f82cfbf4d // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	golang.org/x/tools v0.35.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250414145226-207652e42e2e // indirect

	gopkg.in/yaml.v3 v3.0.1 // indirect
)
