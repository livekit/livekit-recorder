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

Input: Either Url or Template required.  
Output: Either File, RTMP, or S3 required.  
All other fields optional.

```
LIVEKIT_RECORDING_CONFIG = "$(jq -Rs '.' config.json)"
```
config.json:
```yaml
{   
    "Input": {
        "Url": custom url of recording web page
        "Template": {
            "Type": grid | gallery | speaker
            "WSUrl": livekit server websocket url
            "ApiKey": livekit server api key
            "ApiSecret": livekit server api secret
        }
        "Width": defaults to 1920
        "Height": defaults to 1080
        "Depth": defaults to 24
        "Framerate": defaults to 25
    }
    "Output": {
        "File": filename
        "RTMP": rtmp url
        "S3": {
            "AccessID": aws access id
            "Secret": aws secret
            "Bucket": s3 bucket
            "Key": filename
        }
        "Width": optional, scale output width
        "Height": optional, scale output height
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
docker build -t livekit-recorder . \
&& docker run \
    -e LIVEKIT_TEMPLATE="grid" \
    -e LIVEKIT_WS_URL="wss://your-domain.com" \
    -e LIVEKIT_API_KEY="<key>" \
    -e LIVEKIT_API_SECRET="<secret>" \
    livekit-recorder

// copy file to host after completion
docker cp <container_name>:app/recording.mp4 .
```

### Record on custom webpage and upload to s3

s3.json
```json
{
    "Input": {
        "Url": "https://your-recording-domain.com"
    },
    "Output": {
        "S3": {
            "AccessID": "<aws-access-id>",
            "Secret": "<aws-secret>",
            "Bucket": "bucket-name",
            "Key": "recording.mp4"
        }
    }
}
```

```bash
docker build -t livekit-recorder . \
&& docker run -e LIVEKIT_RECORDING_CONFIG="$(jq -Rs '.' s3.json)" livekit-recorder
```

### Streaming to twitch, scaled to 720p

twitch.json
```json
{
    "Input": {
        "Template": {
            "Type": "speaker",
            "WSUrl": "wss://your-domain.com",
            "ApiKey": "<api-key>",
            "ApiSecret": "<api-secret>"
        },
        "Width": 1920,
        "Height": 1080
    },
    "Output": {
        "RTMP": "rtmp://live.twitch.tv/app/<stream key>",
        "Width": 1280,
        "Height": 720
    }
}
```

```bash
docker build -t livekit-recorder . \
&& docker run -e LIVEKIT_RECORDING_CONFIG="$(jq -Rs '.' twitch.json)" livekit-recorder
```
