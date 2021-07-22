const puppeteer = require('puppeteer');
const { exec, spawn } = require("child_process");

(async () => {
	const browser = await puppeteer.launch({
		headless: false,
		defaultViewport: {width: 1920, height: 1080},
		ignoreDefaultArgs: ["--enable-automation"],
		args: ['--kiosk', '--no-sandbox', '--window-size=1920,1080']
	});
	const page = await browser.newPage();
	await page.goto('https://example.livekit.io/#/room?url=wss%3A%2F%2Fdemo2.livekit.io&token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjkzMjAyMDQsImlzcyI6IkFQSU1teGlMOHJxdUt6dFpFb1pKVjlGYiIsImp0aSI6InJyMSIsIm5iZiI6MTYyNjcyODIwNCwidmlkZW8iOnsiY2FuU3Vic2NyaWJlIjp0cnVlLCJoaWRkZW4iOnRydWUsInJvb20iOiJMS0hRIiwicm9vbUpvaW4iOnRydWV9fQ.pFg1z89kc47g5YL1bmkycRLl1NQQkHVDUxwnFUWlBBQ&videoEnabled=1&audioEnabled=1&simulcast=0');
	const [muteAudio] = await page.$x("//button[contains(., 'Mute')]");
	if (muteAudio) {
    	await muteAudio.click();
	}
	const [muteVideo] = await page.$x("//button[contains(., 'Stop Video')]");
	if (muteVideo) {
    	await muteVideo.click();
	}

	console.log('Start recording');
	const ffmpeg = spawn('ffmpeg',  [
		'-video_size', '1920x1080', '-framerate', '25',
		'-f', 'x11grab', '-thread_queue_size', '1024', '-i', ':10.0', 
		// '-vcodec', 'libx264', '-preset', 'ultrafast', '-tune', 'zerolatency',
		'-f', 'pulse', '-thread_queue_size', '1024', '-i', 'grab.monitor', '-ac', '2',
		'-t', '15',
		'recording.mp4'])
	ffmpeg.stdout.pipe(process.stdout);
    ffmpeg.stderr.pipe(process.stderr);
	ffmpeg.on('error', (err) => console.log(err));
	ffmpeg.on('close', () => {
		console.log('Closed')
	});
	
	setTimeout(async () => {
		console.log('Closing');
		ffmpeg.stdin.end();
		await browser.close();
	}, 1000 * 15);

  })();