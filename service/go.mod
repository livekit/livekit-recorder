module github.com/livekit/livekit-recorder/service

go 1.16

require (
	github.com/eapache/channels v1.1.0
	github.com/eapache/queue v1.1.0 // indirect
	github.com/go-logr/logr v1.0.0
	github.com/go-logr/zapr v1.0.0
	github.com/go-redis/redis/v8 v8.11.0
	github.com/livekit/protocol v0.7.8
	github.com/magefile/mage v1.11.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli/v2 v2.3.0
	go.uber.org/zap v1.18.1
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/livekit/protocol => ../../protocol
