package browser

import (
	"context"

	"github.com/chromedp/chromedp"
)

// Chromedpのコンテキストを生成する
func CreateChromedpContext() (context.Context, context.CancelFunc) {
	// create chrome instance
	ctx, cancel := chromedp.NewContext(
		context.Background(),
	)

	return ctx, cancel
}
