package recorder

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/chromedp/chromedp"
	"github.com/livekit/protocol/logger"
)

const Display = ":99"

func (r *Recorder) LaunchXvfb(width, height, depth int) error {
	logger.Debugw("launching xvfb")

	dims := fmt.Sprintf("%dx%dx%d", width, height, depth)
	cmd := exec.Command("Xvfb", Display, "-screen", "0", dims, "-ac", "-nolisten", "tcp")
	if err := cmd.Start(); err != nil {
		return err
	}

	r.xvfb = cmd
	return nil
}

func (r *Recorder) LaunchChrome(url string, width, height int) error {
	logger.Debugw("launching chrome")

	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.WindowSize(width, height),

		// puppeteer default behavior
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("excludeSwitches", "enable-automation"),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-client-side-phishing-detection", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-features", "site-per-process,TranslateUI,BlinkGenPropertyTrees"),
		chromedp.Flag("disable-hang-monitor", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-prompt-on-repost", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("force-color-profile", "srgb"),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("safebrowsing-disable-auto-update", true),
		chromedp.Flag("password-store", "basic"),
		chromedp.Flag("use-mock-keychain", true),

		// custom args
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("autoplay-policy", "no-user-gesture-required"),
		chromedp.Flag("window-position", "0,0"),
		chromedp.Flag("display", Display),
	}

	ctx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	r.chromeCtx, r.chromeCancel = chromedp.NewContext(ctx)

	return chromedp.Run(r.chromeCtx,
		chromedp.Navigate(url),
		// TODO: wait?
	)
}
