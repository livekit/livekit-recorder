type Config = {
    Template?: {
        Type: string
        WSUrl: string
        ApiKey: string
        ApiSecret: string
    }
    Url?: string
    Input: {
        Width: number
        Height: number
        Depth: number
        Framerate: number
    }
    Output: {
        Location: string
        Width?: number
        Height?: number
        AudioBitrate: string
        AudioFrequency: string
        VideoBitrate: string
        VideoBuffer: string
    }
}

const defaultConfig: Config = {
    Input: {
        Width: 1920,
        Height: 1080,
        Depth: 24,
        Framerate: 25,
    },
    Output: {
        Location: (process.env.LIVEKIT_OUTPUT || 'recording.mp4'),
        AudioBitrate: '128k',
        AudioFrequency: '44100',
        VideoBitrate: '2976k',
        VideoBuffer: '5952k'
    }
}

export function loadConfig(): Config {
    if (process.env.LIVEKIT_RECORDING_CONFIG) {
        return {...defaultConfig, ...JSON.parse(process.env.LIVEKIT_RECORDING_CONFIG)}
    }

    const conf = defaultConfig
    if (process.env.LIVEKIT_URL) {
        conf.Url = process.env.LIVEKIT_URL
    } else if (process.env.LIVEKIT_WS_URL && process.env.LIVEKIT_API_KEY && process.env.LIVEKIT_API_SECRET) {
        conf.Template = {
            Type: process.env.LIVEKIT_TEMPLATE || 'gallery',
            WSUrl: process.env.LIVEKIT_WS_URL,
            ApiKey: process.env.LIVEKIT_API_KEY,
            ApiSecret: process.env.LIVEKIT_API_SECRET,
        }
    } else {
        conf.Url = "https://example.livekit.io/#/room?url=wss%3A%2F%2Fdemo2.livekit.io&token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjkzMjAyMDQsImlzcyI6IkFQSU1teGlMOHJxdUt6dFpFb1pKVjlGYiIsImp0aSI6InJyMSIsIm5iZiI6MTYyNjcyODIwNCwidmlkZW8iOnsiY2FuU3Vic2NyaWJlIjp0cnVlLCJoaWRkZW4iOnRydWUsInJvb20iOiJMS0hRIiwicm9vbUpvaW4iOnRydWV9fQ.pFg1z89kc47g5YL1bmkycRLl1NQQkHVDUxwnFUWlBBQ&videoEnabled=0&audioEnabled=1&simulcast=0&recorder=1"
    }

    return conf
}
