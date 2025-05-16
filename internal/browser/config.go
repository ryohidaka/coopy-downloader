package browser

import (
	"context"

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

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancelCtx := chromedp.NewContext(allocCtx)

	return ctx, func() {
		cancelCtx()
		cancel()
	}
}
