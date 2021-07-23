const Xvfb = require('xvfb');
import { Browser, Page, launch } from 'puppeteer'
import { spawn } from 'child_process'

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
	Url: process.env.LIVEKIT_URL || "https://example.livekit.io/#/room?url=wss%3A%2F%2Fdemo2.livekit.io&token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjkzMjAyMDQsImlzcyI6IkFQSU1teGlMOHJxdUt6dFpFb1pKVjlGYiIsImp0aSI6InJyMSIsIm5iZiI6MTYyNjcyODIwNCwidmlkZW8iOnsiY2FuU3Vic2NyaWJlIjp0cnVlLCJoaWRkZW4iOnRydWUsInJvb20iOiJMS0hRIiwicm9vbUpvaW4iOnRydWV9fQ.pFg1z89kc47g5YL1bmkycRLl1NQQkHVDUxwnFUWlBBQ&videoEnabled=0&audioEnabled=1&simulcast=0&recorder=1",
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

function loadConfig(): Config {
	const confString = process.env.LIVEKIT_RECORDING_CONFIG
	if (confString) {
		return {...defaultConfig, ...JSON.parse(confString)}
	}
	return defaultConfig
}

function buildRecorderToken(key: string, secret: string): string {
	return key+secret // TODO
}

(async () => {
	const conf = loadConfig()

	// start xvfb
	const xvfb = new Xvfb({
		displayNum: 10,
		silent: true,
		xvfb_args: ['-screen', '0', `${conf.Input.Width}x${conf.Input.Height}x${conf.Input.Depth}`, '-ac']
	})
	xvfb.start((err: Error) => { if (err) { console.log(err) } })

	// launch puppeteer
	const browser: Browser = await launch({
		headless: false,
		defaultViewport: {width: conf.Input.Width, height: conf.Input.Height},
		ignoreDefaultArgs: ["--enable-automation"],
		args: [
			'--kiosk', // full screen, no info bar
			'--no-sandbox', // required when running as root
			`--window-size=${conf.Input.Width},${conf.Input.Height}`,
			`--display=${xvfb.display()}`,
		]
	})

	// load room
	const page: Page = await browser.newPage()
	let url: string
	if (conf.Template) {
		const token = buildRecorderToken(conf.Template.ApiKey, conf.Template.ApiSecret)
		url = `https://recorder.livekit.io/${conf.Template.Type}?url=${encodeURIComponent(conf.Template.WSUrl)}&token=${token}`
	} else if (conf.Url) {
		url = conf.Url
	} else {
		throw Error('url or template required')
	}
	await page.goto(url)

	// For testing
	const [muteAudio] = await page.$x("//button[contains(., 'Mute')]")
	if (muteAudio) {
		await muteAudio.click()
	}

	// prepare ffmpeg output
	let ffmpegOutputOpts = [
		// audio
		'-c:a', 'aac', '-b:a', conf.Output.AudioBitrate, '-ar', conf.Output.AudioFrequency,
		'-ac', '2', '-af', 'aresample=async=1',
		// video
		'-c:v', 'libx264', '-preset', 'veryfast', '-tune', 'zerolatency',
		'-b:v', conf.Output.VideoBitrate,
	]
	if (conf.Output.Width && conf.Output.Height) {
		ffmpegOutputOpts = ffmpegOutputOpts.concat('-s', `${conf.Output.Width}x${conf.Output.Height}`)
	}

	let ffmpegOutput: string[]
	let uploadFunc: () => void
	if (conf.Output.Location.startsWith('rtmp')) {
		ffmpegOutputOpts = ffmpegOutputOpts.concat(['-maxrate', conf.Output.VideoBitrate, '-bufsize', conf.Output.VideoBuffer])
		ffmpegOutput = ['-f', 'flv', conf.Output.Location]
	} else if (conf.Output.Location.startsWith('s3://')) {
		const filename = 'recording.mp4'
		ffmpegOutput = [filename]
		uploadFunc = function() {
			// TODO: upload to s3
		}
	} else {
		ffmpegOutput = [conf.Output.Location]
	}

	// spawn ffmpeg
	console.log('Start recording')
	const ffmpeg = spawn('ffmpeg', [
		'-fflags', 'nobuffer', // reduce delay
		'-fflags', '+igndts', // generate dts

		// video (x11 grab)
		'-thread_queue_size', '1024', // avoid thread message queue blocking
		'-probesize', '42M', // increase probe size for bitrate estimation
		// consider probesize 32 analyzeduration 0 for lower latency
		'-s', `${conf.Input.Width}x${conf.Input.Height}`,
		'-r', `${conf.Input.Framerate}`,
		'-f', 'x11grab', '-i', `${xvfb.display()}.0`,

		// audio (pulse grab)
		'-thread_queue_size', '1024', // avoid thread message queue blocking
		'-ac', '2', // 2 channels
		'-f', 'pulse', '-i', 'grab.monitor',

		// output
		...ffmpegOutputOpts, ...ffmpegOutput,
	])
	ffmpeg.stdout.pipe(process.stdout)
	ffmpeg.stderr.pipe(process.stderr)
	ffmpeg.on('error', (err) => console.log(err))
	ffmpeg.on('close', () => {
		console.log('ffmpeg finished')
		xvfb.stop()
		uploadFunc && uploadFunc()
	});

	// wait for END_RECORDING
	page.on('console', async (msg) => {
		if (msg.text() === 'END_RECORDING') {
			console.log('End recording')
			ffmpeg.kill('SIGINT')
			await browser.close()
		}
	})
})();
