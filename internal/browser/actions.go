package browser

import (
	"context"
	"log/slog"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ryohidaka/coopy-downloader/internal/constants"
)

// 指定されたIDとパスワードでログイン処理を行う
//
// 引数:
//
//	ctx context.Context: chromedpコンテキスト
//	loginId string: ログインID
//	passwd string: パスワード
//
// 戻り値:
//
//	error: ログインに失敗した場合はエラーを返却する
func Login(ctx context.Context, loginId, passwd string) error {
	err := chromedp.Run(ctx,
		chromedp.Navigate(constants.LoginURL),
	)
	if err != nil {
		slog.Error("ログインページへの遷移に失敗", "error", err)
		return err
	}

	err = chromedp.Run(ctx,
		chromedp.WaitVisible(constants.InputLoginSelector),
		chromedp.SendKeys(constants.InputLoginSelector, loginId),
	)
	if err != nil {
		slog.Error("ログインID入力に失敗", "error", err)
		return err
	}

	err = chromedp.Run(ctx,
		chromedp.WaitVisible(constants.InputPasswordSelector),
		chromedp.SendKeys(constants.InputPasswordSelector, passwd),
	)
	if err != nil {
		slog.Error("パスワード入力に失敗", "error", err)
		return err
	}

	err = chromedp.Run(ctx,
		chromedp.Click(constants.ButtonLoginSelector),
		chromedp.Sleep(3*time.Second),
	)
	if err != nil {
		slog.Error("ログインボタンクリックに失敗", "error", err)
		return err
	}

	return nil
}
