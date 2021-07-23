# livekit-recording

## Examples

### Basic

```
docker build -t recorder \
&& docker run \
    -e LIVEKIT_URL="https://record.livekit.io/grid" \
    -e LIVEKIT_WS_URL="wss://your-domain.com" \
    -e LIVEKIT_TOKEN="<token>" \
    recorder

docker cp <container_name>:app/recording.mp4 .
```

### Upload to s3 (TODO)

```
docker build -t recorder \
&& docker run \
    -e LIVEKIT_URL="https://record.livekit.io/gallery" \
    -e LIVEKIT_WS_URL="wss://your-domain.com" \
    -e LIVEKIT_TOKEN="<token>" \
    -e LIVEKIT_OUTPUT="s3://bucket/path" \
    recorder
```

### Stream to twitch

conf.json
```
{
	"Url": "https://record.livekit.io/speaker",
	"WSUrl": "wss://your-domain.com",
	"Token": "<token>",
	"Input": {
		"Width": 1920,
		"Height": 1080,
		"Depth": 24,
		"Framerate": 25,
	},
	"Output": {
		"Location": "rtmp://live.twitch.tv/app/<stream key>",
		"Width": 1280,
		"Height": 720
	}
}
```

```
docker build -t recorder \
&& docker run -e LIVEKIT_RECORDING_CONFIG="${jq -Rs '.' conf.json}" recorder
```
