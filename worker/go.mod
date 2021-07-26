module github.com/livekit/livekit-recording/worker

go 1.16

require (
	github.com/go-logr/logr v1.0.0
	github.com/go-logr/zapr v1.0.0
	github.com/go-redis/redis/v8 v8.11.0
	github.com/livekit/livekit-server v1.10.6
	github.com/magefile/mage v1.11.0
	github.com/pion/ion-sfu v1.10.7
	github.com/pion/logging v0.2.2
	github.com/pkg/errors v0.9.1
	github.com/twitchtv/twirp v8.1.0+incompatible
	github.com/urfave/cli/v2 v2.3.0
	go.uber.org/zap v1.18.1
	google.golang.org/protobuf v1.27.1
)

replace github.com/livekit/livekit-server => ../../livekit-server
