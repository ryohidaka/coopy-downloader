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
		chromedp.Flag("headless", true),
	)

	if noSandbox {
		opts = append(opts, chromedp.Flag("no-sandbox", true))
	}

	// Create allocator context
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// Set timeout to 30 seconds
	timeoutCtx, timeoutCancel := context.WithTimeout(allocCtx, 60*time.Second)

	// Create chromedp context with timeout
	ctx, cancelCtx := chromedp.NewContext(timeoutCtx)

	// Return composite cancel function
	return ctx, func() {
		cancelCtx()
		timeoutCancel()
		allocCancel()
	}
}
