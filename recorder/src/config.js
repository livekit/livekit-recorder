"use strict";
var __assign = (this && this.__assign) || function () {
    __assign = Object.assign || function(t) {
        for (var s, i = 1, n = arguments.length; i < n; i++) {
            s = arguments[i];
            for (var p in s) if (Object.prototype.hasOwnProperty.call(s, p))
                t[p] = s[p];
        }
        return t;
    };
    return __assign.apply(this, arguments);
};
exports.__esModule = true;
exports.loadConfig = void 0;
function loadConfig() {
    var conf = {
        input: {
            width: 1920,
            height: 1080,
            depth: 24,
            framerate: 25
        },
        output: {
            audioBitrate: '128k',
            audioFrequency: '44100',
            videoBitrate: '2976k',
            videoBuffer: '5952k'
        }
    };
    if (process.env.LIVEKIT_RECORDER_CONFIG) {
        // load config from env
        var json = JSON.parse(process.env.LIVEKIT_RECORDER_CONFIG);
        conf.input = __assign(__assign({}, conf.input), json.input);
        conf.output = __assign(__assign({}, conf.output), json.output);
    }
    else {
        throw Error('LIVEKIT_RECORDER_CONFIG, LIVEKIT_URL or Template required');
    }
    // write to file if no output specified
    if (!(conf.output.file || conf.output.rtmp || conf.output.s3)) {
        conf.output.file = 'recording.mp4';
    }
    return conf;
}
exports.loadConfig = loadConfig;
