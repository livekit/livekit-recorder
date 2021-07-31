"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __generator = (this && this.__generator) || function (thisArg, body) {
    var _ = { label: 0, sent: function() { if (t[0] & 1) throw t[1]; return t[1]; }, trys: [], ops: [] }, f, y, t, g;
    return g = { next: verb(0), "throw": verb(1), "return": verb(2) }, typeof Symbol === "function" && (g[Symbol.iterator] = function() { return this; }), g;
    function verb(n) { return function (v) { return step([n, v]); }; }
    function step(op) {
        if (f) throw new TypeError("Generator is already executing.");
        while (_) try {
            if (f = 1, y && (t = op[0] & 2 ? y["return"] : op[0] ? y["throw"] || ((t = y["return"]) && t.call(y), 0) : y.next) && !(t = t.call(y, op[1])).done) return t;
            if (y = 0, t) op = [op[0] & 2, t.value];
            switch (op[0]) {
                case 0: case 1: t = op; break;
                case 4: _.label++; return { value: op[1], done: false };
                case 5: _.label++; y = op[1]; op = [0]; continue;
                case 7: op = _.ops.pop(); _.trys.pop(); continue;
                default:
                    if (!(t = _.trys, t = t.length > 0 && t[t.length - 1]) && (op[0] === 6 || op[0] === 2)) { _ = 0; continue; }
                    if (op[0] === 3 && (!t || (op[1] > t[0] && op[1] < t[3]))) { _.label = op[1]; break; }
                    if (op[0] === 6 && _.label < t[1]) { _.label = t[1]; t = op; break; }
                    if (t && _.label < t[2]) { _.label = t[2]; _.ops.push(op); break; }
                    if (t[2]) _.ops.pop();
                    _.trys.pop(); continue;
            }
            op = body.call(thisArg, _);
        } catch (e) { op = [6, e]; y = 0; } finally { f = t = 0; }
        if (op[0] & 5) throw op[1]; return { value: op[0] ? op[1] : void 0, done: true };
    }
};
var __spreadArray = (this && this.__spreadArray) || function (to, from) {
    for (var i = 0, il = from.length, j = to.length; i < il; i++, j++)
        to[j] = from[i];
    return to;
};
exports.__esModule = true;
var config_1 = require("./config");
var puppeteer_1 = require("puppeteer");
var child_process_1 = require("child_process");
var aws_sdk_1 = require("aws-sdk");
var fs_1 = require("fs");
var livekit_server_sdk_1 = require("livekit-server-sdk");
var Xvfb = require('xvfb');
function buildRecorderToken(room, key, secret) {
    var at = new livekit_server_sdk_1.AccessToken(key, secret, {
        identity: 'livekit-recorder'
    });
    at.addGrant({
        roomJoin: true,
        room: room,
        canPublish: false,
        canSubscribe: true,
        hidden: true
    });
    return at.toJwt();
}
(function () { return __awaiter(void 0, void 0, void 0, function () {
    var conf, xvfb, browser, page, url, template, token, ffmpegOutputOpts, ffmpegOutput, uploadFunc, filename_1, ffmpeg, stopped, stop, _a, _b, _c, _d, _e, _f;
    return __generator(this, function (_g) {
        switch (_g.label) {
            case 0:
                conf = config_1.loadConfig();
                xvfb = new Xvfb({
                    displayNum: 10,
                    silent: true,
                    xvfb_args: ['-screen', '0', conf.input.width + "x" + conf.input.height + "x" + conf.input.depth, '-ac']
                });
                xvfb.start(function (err) { if (err) {
                    console.log(err);
                } });
                return [4 /*yield*/, puppeteer_1.launch({
                        headless: false,
                        defaultViewport: { width: conf.input.width, height: conf.input.height },
                        ignoreDefaultArgs: ["--enable-automation"],
                        args: [
                            '--kiosk',
                            '--no-sandbox',
                            '--autoplay-policy=no-user-gesture-required',
                            "--window-size=" + conf.input.width + "," + conf.input.height,
                            "--display=" + xvfb.display(),
                        ]
                    })
                    // load room
                ];
            case 1:
                browser = _g.sent();
                return [4 /*yield*/, browser.newPage()];
            case 2:
                page = _g.sent();
                template = conf.input.template;
                if (template) {
                    token = void 0;
                    if (template.token) {
                        token = template.token;
                    }
                    else if (template.roomName && conf.apiKey && conf.apiSecret) {
                        token = buildRecorderToken(template.roomName, conf.apiKey, conf.apiSecret);
                    }
                    else {
                        throw Error('Either token, or room name, api key, and secret required');
                    }
                    url = "https://recorder.livekit.io/#/" + template.type + "?url=" + encodeURIComponent(template.wsUrl) + "&token=" + token;
                }
                else if (conf.input.url) {
                    url = conf.input.url;
                }
                else {
                    throw Error('Input url or template required');
                }
                return [4 /*yield*/, page.goto(url, { waitUntil: "load" })
                    // ffmpeg output options
                ];
            case 3:
                _g.sent();
                ffmpegOutputOpts = [
                    // audio
                    '-c:a', 'aac', '-b:a', conf.output.audioBitrate, '-ar', conf.output.audioFrequency,
                    '-ac', '2', '-af', 'aresample=async=1',
                    // video
                    '-c:v', 'libx264', '-preset', 'veryfast', '-tune', 'zerolatency',
                    '-b:v', conf.output.videoBitrate,
                ];
                if (conf.output.width && conf.output.height) {
                    ffmpegOutputOpts = ffmpegOutputOpts.concat('-s', conf.output.width + "x" + conf.output.height);
                }
                if (conf.output.file) {
                    ffmpegOutput = [conf.output.file];
                    console.log("Writing to app/" + conf.output.file);
                }
                else if (conf.output.rtmp) {
                    ffmpegOutputOpts = ffmpegOutputOpts.concat(['-maxrate', conf.output.videoBitrate, '-bufsize', conf.output.videoBuffer]);
                    ffmpegOutput = ['-f', 'flv', conf.output.rtmp];
                    console.log("Streaming to " + conf.output.rtmp);
                }
                else if (conf.output.s3) {
                    filename_1 = 'recording.mp4';
                    ffmpegOutput = [filename_1];
                    uploadFunc = function () {
                        if (conf.output.s3) {
                            var s3 = void 0;
                            if (conf.output.s3.accessKey && conf.output.s3.secret) {
                                s3 = new aws_sdk_1.S3({ accessKeyId: conf.output.s3.accessKey, secretAccessKey: conf.output.s3.secret });
                            }
                            else {
                                s3 = new aws_sdk_1.S3();
                            }
                            var params = {
                                Bucket: conf.output.s3.bucket,
                                Key: conf.output.s3.key,
                                Body: fs_1.readFileSync(filename_1)
                            };
                            s3.upload(params, undefined, function (err, data) {
                                if (err) {
                                    console.log(err);
                                }
                                else {
                                    console.log("file uploaded to " + data.Location);
                                }
                            });
                        }
                    };
                    console.log("Saving to s3://" + conf.output.s3.bucket + "/" + conf.output.s3.key);
                }
                else {
                    throw Error('Output location required');
                }
                // spawn ffmpeg
                console.log('Start recording');
                ffmpeg = child_process_1.spawn('ffmpeg', __spreadArray(__spreadArray([
                    '-fflags', 'nobuffer',
                    '-fflags', '+igndts',
                    // audio (pulse grab)
                    '-thread_queue_size', '1024',
                    '-ac', '2',
                    '-f', 'pulse', '-i', 'grab.monitor',
                    // video (x11 grab)
                    "-draw_mouse", "0",
                    '-thread_queue_size', '1024',
                    '-probesize', '42M',
                    // consider probesize 32 analyzeduration 0 for lower latency
                    '-s',
                    conf.input.width + "x" + conf.input.height,
                    '-r',
                    "" + conf.input.framerate,
                    '-f', 'x11grab', '-i',
                    xvfb.display() + ".0"
                ], ffmpegOutputOpts), ffmpegOutput));
                ffmpeg.stdout.pipe(process.stdout);
                ffmpeg.stderr.pipe(process.stderr);
                ffmpeg.on('error', function (err) { return console.log(err); });
                ffmpeg.on('close', function () {
                    console.log('ffmpeg finished');
                    xvfb.stop();
                    uploadFunc && uploadFunc();
                });
                stopped = false;
                stop = function () { return __awaiter(void 0, void 0, void 0, function () {
                    return __generator(this, function (_a) {
                        switch (_a.label) {
                            case 0:
                                if (stopped) {
                                    return [2 /*return*/];
                                }
                                stopped = true;
                                console.log('End recording');
                                ffmpeg.kill('SIGINT');
                                return [4 /*yield*/, browser.close()];
                            case 1:
                                _a.sent();
                                return [2 /*return*/];
                        }
                    });
                }); };
                _b = (_a = process).once;
                _c = ['SIGINT'];
                return [4 /*yield*/, stop];
            case 4:
                _b.apply(_a, _c.concat([_g.sent()]));
                _e = (_d = process).once;
                _f = ['SIGTERM'];
                return [4 /*yield*/, stop];
            case 5:
                _e.apply(_d, _f.concat([_g.sent()]));
                // wait for END_RECORDING
                page.on('console', function (msg) { return __awaiter(void 0, void 0, void 0, function () {
                    return __generator(this, function (_a) {
                        switch (_a.label) {
                            case 0:
                                if (!(msg.text() === 'END_RECORDING')) return [3 /*break*/, 2];
                                return [4 /*yield*/, stop()];
                            case 1:
                                _a.sent();
                                _a.label = 2;
                            case 2: return [2 /*return*/];
                        }
                    });
                }); });
                return [2 /*return*/];
        }
    });
}); })()["catch"](function (err) {
    console.log(err);
});
