import { loadConfig } from "./config"
import { Browser, Page, launch } from 'puppeteer'
import { spawn } from 'child_process'
import { S3 } from 'aws-sdk'
import { readFileSync } from 'fs'

const Xvfb = require('xvfb');

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
	if (conf.Input.Template) {
		const token = buildRecorderToken(conf.Input.Template.ApiKey, conf.Input.Template.ApiSecret)
		url = `https://recorder.livekit.io/${conf.Input.Template.Type}?url=${encodeURIComponent(conf.Input.Template.WSUrl)}&token=${token}`
	} else if (conf.Input.Url) {
		url = conf.Input.Url
	} else {
		throw Error('Input url or template required')
	}
	await page.goto(url)

	// For testing
	const [muteAudio] = await page.$x("//button[contains(., 'Mute')]")
	if (muteAudio) {
		await muteAudio.click()
	}

	// ffmpeg output options
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

	// ffmpeg output location
	let ffmpegOutput: string[]
	let uploadFunc: () => void
	if (conf.Output.File) {
		ffmpegOutput = [conf.Output.File]
	} else if (conf.Output.RTMP) {
		ffmpegOutputOpts = ffmpegOutputOpts.concat(['-maxrate', conf.Output.VideoBitrate, '-bufsize', conf.Output.VideoBuffer])
		ffmpegOutput = ['-f', 'flv', conf.Output.RTMP]
	} else if (conf.Output.S3) {
		const filename = 'recording.mp4'

		ffmpegOutput = [filename]
		uploadFunc = function() {
			if (conf.Output.S3) {
				const s3 = new S3({accessKeyId: conf.Output.S3.AccessKey, secretAccessKey: conf.Output.S3.Secret})
				const params = {
					Bucket: conf.Output.S3.Bucket,
					Key: conf.Output.S3.Path,
					Body: readFileSync(filename)
				}
				s3.upload(params, undefined,function(err, data) {
					if (err) {
						console.log(err)
					} else {
						console.log(`file uploaded to ${data.Location}`)
					}
				})
			}
		}
	} else {
		throw Error('Output location required')
	}

	// spawn ffmpeg
	console.log('Start recording')
	const ffmpeg = spawn('ffmpeg', [
		'-fflags', 'nobuffer', // reduce delay
		'-fflags', '+igndts', // generate dts

		// video (x11 grab)
		"-draw_mouse", "0", // don't draw the mouse
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
})().catch((err) => {
	console.log(err)
});
