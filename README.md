# livekit-recording

## Config

### Using env vars
```
LIVEKIT_URL: (required) base url of recording web page
LIVEKIT_WS_URL: (required) livekit server websocket url
LIVEKIT_TOKEN: (required) recorder room token
LIVEKIT_OUTPUT: output stream url or filename, defaults to recording.mp4
```

### Using json config file

config.json:
```
{
	"Url": (required) base url of recording web page
	"WSUrl": (required) livekit server websocket url
	"Token": (required) recorder room token
	"Input": {
		"Width": defaults to 1920
		"Height": defaults to 1080
		"Depth": defaults to 24
		"Framerate": defaults to 25
	}
	"Output": {
		"Location": output stream url or filename, defaults to recording.mp4
		"Width": optional, scales output
		"Height": optional, scales output
		"AudioBitrate": defaults to 128k
		"AudioFrequency": defaults to 44100
		"VideoBitrate": defaults to 1872k
		"VideoBuffer": defaults to 3744k
	}
}
```
```
LIVEKIT_RECORDING_CONFIG="${jq -Rs '.' config.json}"
```

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
