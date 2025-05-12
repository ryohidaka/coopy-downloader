package browser

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/browser"
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

// 指定された企画回の注文ページから注文書をダウンロードする
//
// 引数:
//
//	ctx context.Context: chromedpコンテキスト
//	kikaku string: 対象となる企画回のID
//	downloadPath string: ダウンロード先ディレクトリ
//
// 戻り値:
//
//	error: 処理に失敗した場合はエラーを返却する
func DownloadOrder(ctx context.Context, kikaku string, downloadPath string) error {
	orderURL := constants.OrderBaseURL + kikaku

	// 注文ページへ遷移
	if err := chromedp.Run(ctx, chromedp.Navigate(orderURL)); err != nil {
		return fmt.Errorf("注文ページの遷移に失敗: %w", err)
	}

	var errorVisible, buttonVisible bool

	// エラー表示とダウンロードボタンの有無を同時に確認
	err := chromedp.Run(ctx,
		chromedp.Tasks{
			chromedp.Evaluate(fmt.Sprintf(`document.querySelector("%s") !== null`, constants.ErrorSelector), &errorVisible),
			chromedp.Evaluate(fmt.Sprintf(`document.querySelector("%s") !== null`, constants.ButtonDownloadSelector), &buttonVisible),
		},
	)
	if err != nil {
		return fmt.Errorf("要素の検出に失敗: %w", err)
	}

	if errorVisible {
		return fmt.Errorf("該当の注文書が見つかりません")
	}
	if !buttonVisible {
		return fmt.Errorf("ダウンロードボタンがありません")
	}

	// ダウンロードの進捗監視
	done := make(chan string, 1)

	chromedp.ListenTarget(ctx, func(v interface{}) {
		if ev, ok := v.(*browser.EventDownloadProgress); ok {
			completed := "(unknown)"
			if ev.TotalBytes != 0 {
				completed = fmt.Sprintf("%0.2f%%", float64(ev.ReceivedBytes)/float64(ev.TotalBytes)*100.0)
			}
			slog.Info("ダウンロード進捗", slog.String("state", ev.State.String()), slog.String("completed", completed))

			if ev.State == browser.DownloadProgressStateCompleted {
				done <- ev.GUID
				close(done)
			}
		}
	})

	// ダウンロード動作とボタンクリック
	if err := chromedp.Run(ctx,
		browser.
			SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllowAndName).
			WithDownloadPath(downloadPath).
			WithEventsEnabled(true),
		chromedp.Click(constants.ButtonDownloadSelector, chromedp.ByQuery),
	); err != nil {
		return fmt.Errorf("ダウンロードクリックに失敗: %w", err)
	}

	// 完了待ち
	guid := <-done
	slog.Info("ダウンロード完了", slog.String("guid", guid), slog.String("path", downloadPath))

	// ファイルをリネーム
	filename := kikaku + ".csv"

	files, err := os.ReadDir(downloadPath)
	if err != nil {
		return fmt.Errorf("ダウンロードディレクトリの読み取りに失敗: %w", err)
	}

	for _, f := range files {
		if !f.IsDir() && strings.HasPrefix(f.Name(), guid) {
			oldPath := filepath.Join(downloadPath, f.Name())
			newPath := filepath.Join(downloadPath, filename)
			if err := os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("ファイルのリネームに失敗: %w", err)
			}
			break
		}
	}

	return nil
}
