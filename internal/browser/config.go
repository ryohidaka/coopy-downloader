package browser

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

// Chromedpのコンテキストを生成する
// noSandbox: true の場合、--no-sandbox を付与
func CreateChromedpContext(noSandbox bool) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "chrome"),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true), // /dev/shm の容量不足対策
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("no-zygote", true),
		chromedp.Flag("single-process", true),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("mute-audio", true),
	)

	if noSandbox {
		opts = append(opts, chromedp.Flag("no-sandbox", true))
	}

	// Create allocator context
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// Set timeout to 3 minute
	timeoutCtx, timeoutCancel := context.WithTimeout(allocCtx, 3*time.Minute)

	// Create chromedp context with timeout
	ctx, cancelCtx := chromedp.NewContext(timeoutCtx)

	// Return composite cancel function
	return ctx, func() {
		cancelCtx()
		timeoutCancel()
		allocCancel()
	}
}
