const Xvfb = require('xvfb');
import { Browser, Page, launch } from 'puppeteer'
import { spawn } from 'child_process'

type Config = {
	Url: string | undefined
	WSUrl: string | undefined
	Token: string | undefined
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
	Url: process.env.LIVEKIT_URL || "https://example.livekit.io/#/room",
	WSUrl: process.env.LIVEKIT_WS_URL || "wss%3A%2F%2Fdemo2.livekit.io",
	Token: process.env.LIVEKIT_TOKEN || "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjkzMjAyMDQsImlzcyI6IkFQSU1teGlMOHJxdUt6dFpFb1pKVjlGYiIsImp0aSI6InJyMSIsIm5iZiI6MTYyNjcyODIwNCwidmlkZW8iOnsiY2FuU3Vic2NyaWJlIjp0cnVlLCJoaWRkZW4iOnRydWUsInJvb20iOiJMS0hRIiwicm9vbUpvaW4iOnRydWV9fQ.pFg1z89kc47g5YL1bmkycRLl1NQQkHVDUxwnFUWlBBQ",
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
		VideoBitrate: '1872k',
		VideoBuffer: '3744k'
	}
}

function loadConfig(): Config {
	const confString = process.env.LIVEKIT_RECORDING_CONFIG
	if (confString) {
		return {...defaultConfig, ...JSON.parse(confString)}
	}
	return defaultConfig
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
			`--display=${xvfb.display()}`]
	})

	// load room
	const page: Page = await browser.newPage()
	const url = `${conf.Url}?url=${conf.WSUrl}&token=${conf.Token}`
	await page.goto(url);
	// mute audio and video for example.livekit.io
	const [muteAudio] = await page.$x("//button[contains(., 'Mute')]")
	if (muteAudio) {
		await muteAudio.click()
	}
	const [muteVideo] = await page.$x("//button[contains(., 'Stop Video')]")
	if (muteVideo) {
		await muteVideo.click()
	}

	// prepare ffmpeg output
	let ffmpegOutputOpts = [
		'-c:v', 'libx264', '-preset', 'veryfast', '-tune', 'zerolatency',
		'-b:v', conf.Output.VideoBitrate, '-maxrate', conf.Output.VideoBitrate, '-bufsize', conf.Output.VideoBuffer,
		'-c:a', 'aac', '-b:a', conf.Output.AudioBitrate, '-ar', conf.Output.AudioFrequency, '-ac', '2',
	]
	if (conf.Output.Width && conf.Output.Height) {
		ffmpegOutputOpts = ffmpegOutputOpts.concat('-s', `${conf.Output.Width}x${conf.Output.Height}`)
	}

	let ffmpegOutput: string[] = []
	let uploadFunc: () => void
	if (conf.Output.Location.startsWith('rtmp')) {
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
		// video (x11 grab)
		'-fflags', '+igndts', // generate dts
		'-thread_queue_size', '64', // avoid thread message queue blocking
		'-probesize', '42M', // increase probe size for bitrate estimation
		'-s', `${conf.Input.Width}x${conf.Input.Height}`,
		'-r', `${conf.Input.Framerate}`,
		'-f', 'x11grab', '-i', `${xvfb.display()}.0`,

		// audio (pulse grab)
		'-fflags', '+igndts', // generate dts
		'-thread_queue_size', '64', // avoid thread message queue blocking
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

	// stop recording
	setTimeout(async () => {
		console.log('Closing')
		ffmpeg.kill('SIGINT')
		await browser.close()
	}, 1000 * 10);

	// page.on('console', async (msg) => {
	// 	if (msg.text() === 'END_RECORDING') {
	// 		console.log('End recording')
	// 		ffmpeg.kill('SIGINT')
	// 		await browser.close()
	// 	}
	// })
})();