const { launch, getStream } = require("puppeteer-stream");
const fs = require("fs");
const { exec } = require("child_process");

async function test() {
	const browser = await launch({
		defaultViewport: {
			width: 1280,
			height: 720,
		},
	});

	const page = await browser.newPage();
	await page.goto("https://example.livekit.io/#/room?url=wss%3A%2F%2Fdemo2.livekit.io&token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjkzMjAyMDQsImlzcyI6IkFQSU1teGlMOHJxdUt6dFpFb1pKVjlGYiIsImp0aSI6InJyMSIsIm5iZiI6MTYyNjcyODIwNCwidmlkZW8iOnsiY2FuU3Vic2NyaWJlIjp0cnVlLCJoaWRkZW4iOnRydWUsInJvb20iOiJMS0hRIiwicm9vbUpvaW4iOnRydWV9fQ.pFg1z89kc47g5YL1bmkycRLl1NQQkHVDUxwnFUWlBBQ&videoEnabled=1&audioEnabled=1&simulcast=0");
	const [muteAudio] = await page.$x("//button[contains(., 'Mute')]");
	if (muteAudio) {
    	await muteAudio.click();
	}
	const [muteVideo] = await page.$x("//button[contains(., 'Mute')]");
	if (muteVideo) {
    	await muteVideo.click();
	}
	const stream = await getStream(page, { audio: true, video: true });
	console.log("recording");
	// this will pipe the stream to ffmpeg and convert the webm to mp4 format
	const ffmpeg = exec(`ffmpeg -y -i - output.mp4`);
	ffmpeg.stderr.on("data", (chunk) => {
		console.log(chunk.toString());
	});

	stream.pipe(ffmpeg.stdin);

	setTimeout(async () => {
		await stream.destroy();
		stream.on("end", () => {});
		// ffmpeg.stdin.setEncoding("utf8");
		// ffmpeg.stdin.write("q");
		// ffmpeg.stdin.end();
		// ffmpeg.kill();

		console.log("finished");
	}, 1000 * 30);
}

test();