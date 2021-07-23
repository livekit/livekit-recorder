const Xvfb = require('xvfb');
import { Browser, Page, launch } from 'puppeteer'
import { spawn } from 'child_process'

(async () => {
	// var config: {}
	// const confString = process.env.LIVEKIT_RECORDING_CONFIG

	// JSON.parse(confString)

	const xvfb = new Xvfb({
		displayNum: 10,
		silent: true,
		xvfb_args: ['-screen', '0', '1920x1080x24', '-ac']
	})
	xvfb.start((err: Error) => {
		if (err) {
			console.log(err)
		}
	})

	// launch puppeteer
	const browser: Browser = await launch({
		headless: false,
		defaultViewport: {width: 1920, height: 1080},
		ignoreDefaultArgs: ["--enable-automation"],
		args: ['--kiosk', '--no-sandbox', '--window-size=1920,1080', '--display='+xvfb.display()]
	})
	const page: Page = await browser.newPage()
	await page.goto('https://example.livekit.io/#/room?url=wss%3A%2F%2Fdemo2.livekit.io&token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjkzMjAyMDQsImlzcyI6IkFQSU1teGlMOHJxdUt6dFpFb1pKVjlGYiIsImp0aSI6InJyMSIsIm5iZiI6MTYyNjcyODIwNCwidmlkZW8iOnsiY2FuU3Vic2NyaWJlIjp0cnVlLCJoaWRkZW4iOnRydWUsInJvb20iOiJMS0hRIiwicm9vbUpvaW4iOnRydWV9fQ.pFg1z89kc47g5YL1bmkycRLl1NQQkHVDUxwnFUWlBBQ&videoEnabled=1&audioEnabled=1&simulcast=0');
	
	// mute audio and video for example.livekit.io
	const [muteAudio] = await page.$x("//button[contains(., 'Mute')]")
	if (muteAudio) {
		await muteAudio.click()
	}
	const [muteVideo] = await page.$x("//button[contains(., 'Stop Video')]")
	if (muteVideo) {
		await muteVideo.click()
	}

	// spawn ffmpeg
	console.log('Start recording')
	const ffmpeg = spawn('ffmpeg',  [
		// generate DTS
		'-fflags', '+igndts',
		// video options
		'-video_size', '1920x1080', '-framerate', '25',
		// x11 grab
		'-f', 'x11grab', '-thread_queue_size', '1024', '-i', ':10.0',
		// pulse grab
		'-f', 'pulse', '-thread_queue_size', '1024', '-i', 'grab.monitor', '-ac', '2',
		// output options
		'-preset', 'ultrafast', '-vcodec', 'libx264', '-tune', 'zerolatency',
		// output
		'recording.mp4'])
	ffmpeg.stdout.pipe(process.stdout)
	ffmpeg.stderr.pipe(process.stderr)
	ffmpeg.on('error', (err) => console.log(err))
	ffmpeg.on('close', () => {
		console.log('ffmpeg finished')
		xvfb.stop()
	});

	// stop recording
	setTimeout(async () => {
		console.log('Closing')
		ffmpeg.kill('SIGINT')
		await browser.close()
	}, 1000 * 5);

	// page.on('console', async (msg) => {
	// 	if (msg.text() === 'END_RECORDING') {
	// 		console.log('End recording')
	// 		ffmpeg.kill('SIGINT')
	// 		await browser.close()
	// 	}
	// })
})();