# LiveKit Recording

All your live recording needs in one place.  
Record any website using our recorder, or deploy our service to manage it for you.

## How it works

The recorder launches Chrome and navigates to the supplied url, grabs audio from pulse and video from a virtual frame buffer, and feeds them into GStreamer.
You can write the output as mp4 to a file or upload it to s3, or forward the output to one or multiple rtmp streams.

## Config

Whether running the standalone recorder or running in service mode, you'll need a yaml config file. All config options:

```yaml
api_key: livekit server api key (required if using templates without supplying tokens)
api_secret: livekit server api secret (required if using templates without supplying tokens)
ws_url: livekit server ws url (required if using templates)
health_port: http port to serve status (optional)
log_level: valid levels are debug, info, warn, error, fatal, or panic. Defaults to debug
gst_log_level: (string) valid levels are "0" (none) to "9" (memdump). Anything above "3" can be very noisy. Defaults to "3".
redis: (required by service mode)
    address: redis address, including port
    username: redis username (optional)
    password: redis password (optional)
    db: redis db (optional)
s3: (required if using s3 output)
    access_key: s3 access key
    secret: s3 access secret
    region: s3 region
defaults:
    preset: defaults to 0
    input_width: defaults to 1920
    input_height: defaults to 1080
    depth: defaults to 24
    framerate: defaults to 30
    output_width: defaults to 0 (no scaling)
    output_height: defaults to 0 (no scaling)
    audio_bitrate: defaults to 128 (kbps)
    audio_frequency: defaults to 44100 (Hz)
    video_bitrate: defaults to 4500 (kbps)
```

If you will be using templates with your recording requests, `ws_url` is required, and to record by room name instead of token, `api_key` and
`api_secret` are also required.

When running in service mode, redis config is required (it needs to connect to the same db as your LiveKit server).

The `defaults.preset` field will provide the following values as defaults:

| Preset  | input_width | input_height | framerate | video_bitrate |
|---      |---          |---           |---        |---            |
| 1       | 1280        | 720          | 30        | 3000          |
| 2       | 1280        | 720          | 60        | 4500          |
| 3       | 1920        | 1080         | 30        | 4500          |
| 4       | 1920        | 1080         | 60        | 6000          |

If you don't supply any options, it defaults to 1080p 30 fps.

## Request

See StartRecordingRequest [here](https://github.com/livekit/protocol/blob/main/livekit_recording.proto#L16). When using standalone mode, the request can be input as a json file.

### Template input

We currently have 4 templates available - grid or speaker, each available in light or dark.
Just supply your server api key and secret, along with the websocket url.  
Check out our [web README](https://github.com/livekit/livekit-recorder/tree/main/web) to learn more or create your own.

```json
{
    "template": {
        "layout": "<grid|speaker>-<light|dark>",
        "room_name": "room-to-record"
    }
    // output...
}
```
Or, to use your own token instead of having the recorder generate one:
```json
{
    "template": {
        "layout": "<grid|speaker>-<light|dark>",
        "token": "<token>"
    }
    // output...
}
```

### Webpage input

You can also save or stream any other webpage - just supply the url.
```json
{   
    "url": "your-recording-domain.com",
    // output...
}
```

## Output

### Save to file

```json
{
    // input...
    "file": "/app/out/recording.mp4",
}
```
Note: your local mounted directory needs to exist, and the docker directory should match file output (i.e. `/app/out`)
```bash
mkdir -p ~/livekit/output

docker run --rm -e LIVEKIT_RECORDER_CONFIG="$(cat config.json)" \
    -v ~/livekit/output:/app/out \
    livekit/livekit-recorder
```

### Upload to S3

```json
{
    // input...
    "s3Url": "bucket/path/filename.mp4"
}
```

```bash
docker run --rm -e LIVEKIT_RECORDER_CONFIG="$(cat config.json)" livekit/livekit-recorder
```

### RTMP

```json
{
    // input...
    "rtmp": {
        "urls": ["<rtmp://stream-url.com>"]
    }
}
```

```bash
docker run --rm -e LIVEKIT_RECORDER_CONFIG="$(cat config.json)" livekit/livekit-recorder
```

## Ending a recording

Once started, there are a number of ways to end the recording:
* `docker stop <container>`
* if using our templates, the recorder will stop automatically when the last participant leaves
* if using your own webpage, logging `END_RECORDING` to the console

With any of these methods, the recorder will stop ffmpeg and finish uploading before shutting down.

## Examples

### Basic recording

basic.json:
```json
{
    "template": {
        "layout": "speaker-dark",
        "ws_url": "<wss://livekit.your-domain.com>",
        "room_name": "<my-room>"
    },
    "file": "/app/out/recording.mp4"
}
```
```bash
mkdir -p ~/livekit/output

docker run --rm -e LIVEKIT_RECORDER_CONFIG="$(cat basic.json)" \
    -v ~/livekit/output:/app/out \
    livekit/livekit-recorder
```

### Record custom url at 720p, 60fps and upload to s3

s3.json:
```json
{
    "url": "https://your-recording-domain.com",
    "s3Url": "bucket/path/filename.mp4",
    "options": {
        "preset": "720p60"
    }
}
```
```bash
docker run --rm --name my-recorder -e LIVEKIT_RECORDER_CONFIG="$(cat s3.json)" livekit/livekit-recorder
```
```bash
docker stop my-recorder
```

### Stream to Twitch, scaling output from 1080p to 720p

twitch.json:
```json
{
    "template": {
        "layout": "speaker-dark",
        "token": "<recording-token>"
    },
    "rtmp": {
        "urls": ["rtmp://live.twitch.tv/app/<stream-key>"]
    },
    "options": {
        "input_width": 1920,
        "input_height": 1080,
        "output_width": 1280,
        "output_height": 720
    }
}
```
```bash
docker run --rm -e LIVEKIT_RECORDER_CONFIG="$(cat twitch.json)" livekit/livekit-recorder
```

# Service Mode

Simply deploy the service, and submit requests through your LiveKit server.

### How it works

The service listens to a redis subscription and waits for the LiveKit server to make a reservation. Once the reservation
is made to ensure availability, the service waits for a StartRecording request from the server before launching the recorder.
The recorder will be stopped by either a `END_RECORDING` signal from the server, or automatically when the last participant leaves if using our templates.

A single service instance can record one room at a time.

### Deployment

See guides and deployment docs at https://docs.livekit.io/guides/recording

### Running locally

If you want to try running against a local livekit server, you'll need to make a couple changes:
* open `/usr/local/etc/redis.conf` and comment out the line that says `bind 127.0.0.1`
* change `protected-mode yes` to `protected-mode no` in the same file
* add `--network host` to your `docker run` command
* update your redis address from `localhost` to your host ip as docker sees it:
    * on linux, this should be `172.17.0.1`
    * on mac or windows, run `docker run -it --rm alpine nslookup host.docker.internal` and you should see something like
      `Name:	host.docker.internal
      Address: 192.168.65.2`

These changes allow the service to connect to your local redis instance from inside the docker container.
Finally, to build and run:
```bash
docker build -t recorder-svc . 
docker run --network host -e REDIS_HOST="192.168.65.2:6379" recorder-svc
```

You can then use our [cli](https://github.com/livekit/livekit-cli) to submit recording requests to your server.
