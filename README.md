# livekit-recording

## Config

### Using templates

We have 3 templates available - grid, gallery, and speaker.  
Just supply your server url, api key, and secret.
```bash
LIVEKIT_TEMPLATE = <grid/gallery/speaker>
LIVEKIT_WS_URL = <livekit server ws url>
LIVEKIT_API_KEY = <livekit server api key>
LIVEKIT_API_SECRET = <livekit server api secret>
```

### Using custom url

You can also use your own custom recording url - your app/server will need to handle room connection.  
To stop the recorder, the page should log a `console.log('END_RECORDING')` message.  
Our templates send this message when the last participant leaves the room. 

```bash
LIVEKIT_URL = <custom recording webpage url>
```

### Using json config file

Either Template or Url required - all other fields optional.
```
LIVEKIT_RECORDING_CONFIG = "${jq -Rs '.' config.json}"
```
config.json:
```yaml
{
    "Template": {
        "Type": grid | gallery | speaker
        "WSUrl": livekit server websocket url
        "ApiKey": livekit server api key
        "ApiSecret": livekit server api secret
    }
    "Url": custom url of recording web page
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

## Examples

### Basic

```bash
docker build -t recorder \
&& docker run \
    -e LIVEKIT_TEMPLATE="grid" \
    -e LIVEKIT_WS_URL="wss://your-domain.com" \
    -e LIVEKIT_API_KEY="<key>" \
    -e LIVEKIT_API_SECRET="<secret>" \
    recorder

// copy file to host after completion
docker cp <container_name>:app/recording.mp4 .
```

### Record on custom webpage and upload to s3

```bash
docker build -t recorder \
&& docker run \
    -e LIVEKIT_URL="https://your-domain.com/record" \
    -e LIVEKIT_OUTPUT="s3://bucket/path" \
    recorder
```

### Stream to twitch

twitch.json
```json
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

```bash
docker build -t recorder \
&& docker run -e LIVEKIT_RECORDING_CONFIG="${jq -Rs '.' twitch.json}" recorder
```
