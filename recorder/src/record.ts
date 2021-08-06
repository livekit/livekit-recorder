import { loadConfig } from "./config"
import { Browser, Page, launch } from 'puppeteer'
import { spawn } from 'child_process'
import { S3 } from 'aws-sdk'
import { readFileSync } from 'fs'
import { AccessToken } from 'livekit-server-sdk'

const Xvfb = require('xvfb');

function buildRecorderToken(room: string, key: string, secret: string): string {
	const at = new AccessToken(key, secret, {
		identity: 'recorder-'+(Math.random()+1).toString(36).substring(2),
	})
	at.addGrant({
		roomJoin: true,
		room: room,
		canPublish: false,
		canSubscribe: true,
		hidden: true,
	})
	return at.toJwt()
}

(async () => {
	const conf = loadConfig()

	// start xvfb
	const xvfb = new Xvfb({
		displayNum: 10,
		silent: true,
		xvfb_args: ['-screen', '0', `${conf.options.input_width}x${conf.options.input_height}x${conf.options.depth}`, '-ac']
	})
	xvfb.start((err: Error) => { if (err) { console.log(err) } })

	// launch puppeteer
	const browser: Browser = await launch({
		headless: false,
		defaultViewport: {width: conf.options.input_width, height: conf.options.input_height},
		ignoreDefaultArgs: ["--enable-automation"],
		args: [
			'--kiosk', // full screen, no info bar
			'--no-sandbox', // required when running as root
			'--autoplay-policy=no-user-gesture-required', // autoplay
			'--window-position=0,0',
			`--window-size=${conf.options.input_width},${conf.options.input_height}`,
			`--display=${xvfb.display()}`,
		]
	})

	// load room
	const page: Page = await browser.newPage()
	let url: string
	const template = conf.input.template
	if (template) {
		let token: string
		if (template.token) {
			token = template.token
		} else if (template.room_name && conf.api_key && conf.api_secret) {
			token = buildRecorderToken(template.room_name, conf.api_key, conf.api_secret)
		} else {
			throw Error('Either token, or room name, api key, and secret required')
		}
		url = `https://recorder.livekit.io/#/${template.layout}?url=${encodeURIComponent(template.ws_url)}&token=${token}`
	} else if (conf.input.url) {
		url = conf.input.url
	} else {
		throw Error('Input url or template required')
	}
	await page.goto(url, {waitUntil: "load"})
	await new Promise(resolve => {setTimeout(resolve, 15000)})

	// ffmpeg output options
	let ffmpegOutputOpts = [
		// audio
		'-c:a', 'aac', '-b:a', `${conf.options.audio_bitrate}k`, '-ar', `${conf.options.audio_frequency}`,
		'-ac', '2', '-af', 'aresample=async=1',
		// video
		'-c:v', 'libx264', '-preset', 'veryfast', '-tune', 'zerolatency',
		'-b:v', `${conf.options.video_bitrate}k`,
	]
	if (conf.options.output_width && conf.options.output_height) {
		ffmpegOutputOpts = ffmpegOutputOpts.concat('-s', `${conf.options.output_width}x${conf.options.output_height}`)
	}

	// ffmpeg output location
	let ffmpegOutput: string[]
	let uploadFunc: () => void
	if (conf.output.rtmp) {
		ffmpegOutputOpts = ffmpegOutputOpts.concat([
			'-maxrate', `${conf.options.video_bitrate}k`,
			'-bufsize', `${conf.options.video_bitrate * 2}k`
		])
		ffmpegOutput = ['-f', 'flv', conf.output.rtmp]
		console.log(`Streaming to ${conf.output.rtmp}`)
	} else if (conf.output.file) {
		const filename = conf.output.file
		ffmpegOutput = [filename]
		if (conf.output.s3) {
			uploadFunc = function() {
				if (conf.output.s3) {
					let s3: S3
					if (conf.output.s3.access_key && conf.output.s3.secret) {
						s3 = new S3({accessKeyId: conf.output.s3.access_key, secretAccessKey: conf.output.s3.secret})
					} else {
						s3 = new S3()
					}
					const params = {
						Bucket: conf.output.s3.bucket,
						Key: conf.output.s3.key,
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
			console.log(`Saving to s3://${conf.output.s3.bucket}/${conf.output.s3.key}`)
		} else if (filename.startsWith('/')) {
			console.log(`Writing to ${filename}`)
		} else {
			console.log(`Writing to /app/${filename}`)
		}
	} else {
		throw Error("Missing ffmpeg output")
	}

	// spawn ffmpeg
	console.log('Start recording')
	const ffmpeg = spawn('ffmpeg', [
		'-fflags', 'nobuffer', // reduce delay
		'-fflags', '+igndts', // generate dts
		'-y', // automatically overwrite

		// audio (pulse grab)
		'-thread_queue_size', '1024', // avoid thread message queue blocking
		'-ac', '2', // 2 channels
		'-f', 'pulse', '-i', 'grab.monitor',

		// video (x11 grab)
		"-draw_mouse", "0", // don't draw the mouse
		'-thread_queue_size', '1024', // avoid thread message queue blocking
		'-s', `${conf.options.input_width}x${conf.options.input_height}`,
		'-r', `${conf.options.framerate}`,
		'-f', 'x11grab', '-i', `${xvfb.display()}.0`,

		'-t', '30',
		// output
		...ffmpegOutputOpts, ...ffmpegOutput,
	])
	ffmpeg.stdout.pipe(process.stdout)
	ffmpeg.stderr.pipe(process.stderr)
	ffmpeg.on('error', (err) => {
		console.log(`ffmpeg error: ${err}`)
	})
	ffmpeg.on('close', (code, signal) => {
		console.log(`ffmpeg closed. code: ${code}, signal: ${signal}`)
		xvfb.stop()
		uploadFunc && uploadFunc()
	});

	let stopped = false
	const stop = async () => {
		if (stopped) {
			return
		}
		stopped = true
		console.log('End recording')
		ffmpeg.kill('SIGINT')
		await browser.close()
	}
	process.once('SIGINT', await stop)
	process.once('SIGTERM', await stop)

	// wait for END_RECORDING
	page.on('console', async (msg) => {
		if (msg.text() === 'END_RECORDING') {
			await stop()
		}
	})
})().catch((err) => {
	console.log(err)
});
