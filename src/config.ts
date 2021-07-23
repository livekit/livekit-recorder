type Config = {
    Input: {
        Url?: string
        Template?: {
            Type: string
            WSUrl: string
            ApiKey: string
            ApiSecret: string
        }
        Width: number
        Height: number
        Depth: number
        Framerate: number
    }
    Output: {
        File?: string
        RTMP?: string
        S3?: {
            AccessID: string
            Secret: string
            Bucket: string
            Key?: string
        }
        Width?: number
        Height?: number
        AudioBitrate: string
        AudioFrequency: string
        VideoBitrate: string
        VideoBuffer: string
    }
}

export function loadConfig(): Config {
    let conf: Config = {
        Input: {
            Width: 1920,
            Height: 1080,
            Depth: 24,
            Framerate: 25,
        },
        Output: {
            AudioBitrate: '128k',
            AudioFrequency: '44100',
            VideoBitrate: '2976k',
            VideoBuffer: '5952k'
        }
    }

    if (process.env.LIVEKIT_RECORDING_CONFIG) {
        // load config from env
        conf = {...conf, ...JSON.parse(process.env.LIVEKIT_RECORDING_CONFIG)}
    } else if (process.env.LIVEKIT_URL) {
        // set url from env
        conf.Input.Url = process.env.LIVEKIT_URL
    } else if (process.env.LIVEKIT_WS_URL && process.env.LIVEKIT_API_KEY && process.env.LIVEKIT_API_SECRET) {
        // set template from env
        conf.Input.Template = {
            Type: process.env.LIVEKIT_TEMPLATE || 'gallery',
            WSUrl: process.env.LIVEKIT_WS_URL,
            ApiKey: process.env.LIVEKIT_API_KEY,
            ApiSecret: process.env.LIVEKIT_API_SECRET,
        }
    } else {
        // TODO: throw Error('LIVEKIT_RECORDING_CONFIG, LIVEKIT_URL or Template required')
        conf.Input.Url = "https://example.livekit.io/#/room?url=wss%3A%2F%2Fdemo2.livekit.io&token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjkzMjAyMDQsImlzcyI6IkFQSU1teGlMOHJxdUt6dFpFb1pKVjlGYiIsImp0aSI6InJyMSIsIm5iZiI6MTYyNjcyODIwNCwidmlkZW8iOnsiY2FuU3Vic2NyaWJlIjp0cnVlLCJoaWRkZW4iOnRydWUsInJvb20iOiJMS0hRIiwicm9vbUpvaW4iOnRydWV9fQ.pFg1z89kc47g5YL1bmkycRLl1NQQkHVDUxwnFUWlBBQ&videoEnabled=0&audioEnabled=1&simulcast=0&recorder=1"
    }

    // write to file if no output specified
    if (!(conf.Output.File || conf.Output.RTMP || conf.Output.S3)) {
        conf.Output.File = 'recording.mp4'
    }

    return conf
}
