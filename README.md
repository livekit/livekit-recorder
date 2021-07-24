# livekit-recording

## What it does

The recorder grabs audio from pulse and video from a virtual frame buffer, and feeds them into ffmpeg.  
You can write the output as mp4 to a file or upload it to s3, or forward the output to a rtmp stream.  
If you don't supply any output options, it will write to `/app/recording.mp4`

A simple example:
```bash
docker build -t livekit-recorder . \
&& docker run \
    -e LIVEKIT_TEMPLATE="gallery" \
    -e LIVEKIT_WS_URL="wss://your-livekit-address.com" \
    -e LIVEKIT_API_KEY="<key>" \
    -e LIVEKIT_API_SECRET="<secret>" \
    livekit-recorder

// copy file to host after completion
docker cp <container_name>:app/recording.mp4 .
```

## Recording Options

### Using templates

We have 3 templates available - grid, gallery, and speaker. Just supply your server url, api key, and secret.

Config:
```bash
LIVEKIT_TEMPLATE="grid|gallery|speaker"
LIVEKIT_WS_URL="wss://your-livekit-address.com"
LIVEKIT_API_KEY="your-livekit-api-key"
LIVEKIT_API_SECRET="your-livekit-api-secret"
```
 -- or --

```json
{   
    "Input": {
        "Template": {
            "Type": "grid|gallery|speaker",
            "WSUrl": "wss://your-livekit-address.com",
            "ApiKey": "<key>",
            "ApiSecret": "<secret>"
        }
    }
}
```

### Using a custom webpage

You can also use your own custom recoding webpages - just supply the url.  
```bash
LIVEKIT_URL="your-recording-domain.com"
```
 -- or --
```json
{   
    "Input": {
        "Url": "your-recording-domain.com"
    }
}
```

### Using json config file

To use a config file, supply the full file as a string in `LIVEKIT_RECORDING_CONFIG`:
```bash
LIVEKIT_RECORDING_CONFIG="$(cat config.json)"
```
Input: Either Url or Template required.  
Output: Either File, RTMP, or S3 required.  
All other fields optional.

All config options:
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
            "AccessKey": aws access id
            "Secret": aws secret
            "Bucket": s3 bucket
            "Path": filename
        }
        "Width": scale output width
        "Height": scale output height
        "AudioBitrate": defaults to 128k
        "AudioFrequency": defaults to 44100
        "VideoBitrate": defaults to 2976k
        "VideoBuffer": defaults to 5952k
    }
}
```

## Examples

### Record on custom webpage and upload to S3

s3.json
```json
{
    "Input": {
        "Url": "https://your-recording-domain.com"
    },
    "Output": {
        "S3": {
            "AccessKey": "<aws-access-key>",
            "Secret": "<aws-secret>",
            "Bucket": "bucket-name",
            "Path": "recording.mp4"
        }
    }
}
```

```bash
docker build -t livekit-recorder . \
&& docker run -e LIVEKIT_RECORDING_CONFIG="$(cat s3.json)" livekit-recorder
```

### Stream to Twitch, scaled to 720p

twitch.json
```json
{
    "Input": {
        "Template": {
            "Type": "speaker",
            "WSUrl": "wss://your-livekit-address.com",
            "ApiKey": "<key>",
            "ApiSecret": "<secret>"
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
&& docker run -e LIVEKIT_RECORDING_CONFIG="$(cat twitch.json)" livekit-recorder
```

## Building your own templates

When using this option, you must handle token generation/room connection - the recorder will only open the url and start recording.

To stop the recorder, the page must send a `console.log('END_RECORDING')`.  
For example, our templates do the following:
```  
const onParticipantDisconnected = (room: Room) => {
    updateParticipantSize(room)

    /* Special rule for recorder */
    if (recorder && parseInt(recorder, 10) === 1 && room.participants.size === 0) {
      console.log("END_RECORDING")
    }
}
```
