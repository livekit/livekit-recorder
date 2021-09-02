# LiveKit Recorder Service

## How it works

The service listens to a redis subscription and waits for the LiveKit server to make a reservation. Once the reservation 
is made to ensure availability, the service waits for a `START_RECORDING` signal from the server before launching the
[recorder](https://github.com/livekit/livekit-recorder/tree/main/recorder). The recorder will be stopped by 
either a `END_RECORDING` signal from the server, or automatically when the last participant leaves if using our templates.

A single instance can record one room at a time.

## Guides

See guides and deployment docs at https://docs.livekit.io/guides/recording

## Config

The only required field is redis address. This must be the same redis address used by your LiveKit server.  
If you want to use templates without supplying your own tokens, `api_key` and 
`api_secret` are also required.

The `options.preset` field will provide defaults using the following values:

| Preset  | input_width | input_height | framerate | video_bitrate |
|---      |---          |---           |---        |---            |
| 720p30  | 1280        | 720          | 30        | 3000          |
| 720p60  | 1280        | 720          | 60        | 4500          |
| 1080p30 | 1920        | 1080         | 30        | 4500          |
| 1080p60 | 1920        | 1080         | 60        | 6000          |

If you don't supply any options, it defaults to 1080p 30 fps.

```yaml
redis:
    address: redis address, including port (required)
    username: redis username (optional)
    password: redis password (optional)
    db: redis db (optional)
# template config
ws_url: livekit server ws url (required if using templates)
api_key: livekit server api key (required if using templates without supplying tokens)
api_secret: livekit server api secret (required if using templates without supplying tokens)
# default recording options (all optional)
options:
    preset: valid options are "720p30", "720p60", "1080p30", or "1080p60"
    input_width: defaults to 1920
    input_height: defaults to 1080
    depth: defaults to 24
    framerate: defaults to 30
    width: defaults to 0 (no scaling)
    height: defaults to 0 (no scaling)
    audio_bitrate: defaults to 128 (kbps)
    audio_frequency: defaults to 44100 (Hz)
    video_bitrate: defaults to 4500 (kbps)
log_level: valid levels are debug, info, warn, error, fatal, or panic (optional)

```

## Running locally

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
