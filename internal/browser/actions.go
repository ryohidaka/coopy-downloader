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
	slog.Debug("ログインページへ遷移開始")
	err := chromedp.Run(ctx,
		chromedp.Navigate(constants.LoginURL),
	)
	if err != nil {
		slog.Error("ログインページへの遷移に失敗", "error", err)
		return err
	}
	slog.Debug("ログインページへ遷移完了")

	slog.Debug("ログインID入力開始")
	err = chromedp.Run(ctx,
		chromedp.WaitVisible(constants.InputLoginSelector),
		chromedp.SendKeys(constants.InputLoginSelector, loginId),
	)
	if err != nil {
		slog.Error("ログインID入力に失敗", "error", err)
		return err
	}
	slog.Debug("ログインID入力完了")

	slog.Debug("パスワード入力開始")
	err = chromedp.Run(ctx,
		chromedp.WaitVisible(constants.InputPasswordSelector),
		chromedp.SendKeys(constants.InputPasswordSelector, passwd),
	)
	if err != nil {
		slog.Error("パスワード入力に失敗", "error", err)
		return err
	}
	slog.Debug("パスワード入力完了")

	slog.Debug("ログインボタンクリック開始")
	err = chromedp.Run(ctx,
		chromedp.Click(constants.ButtonLoginSelector),
		chromedp.Sleep(3*time.Second),
	)
	if err != nil {
		slog.Error("ログインボタンクリックに失敗", "error", err)
		return err
	}
	slog.Debug("ログインボタンクリック完了")

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
	// DownloadOrder全体に3分のタイムアウトを適用
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	orderURL := constants.OrderBaseURL + kikaku

	// 注文ページへ遷移
	slog.Debug("注文ページへ遷移開始", slog.String("url", orderURL))

	if err := chromedp.Run(ctx, chromedp.Navigate(orderURL)); err != nil {
		slog.Error("注文ページの遷移に失敗", "error", err)
		return fmt.Errorf("注文ページの遷移に失敗: %w", err)
	}
	slog.Debug("注文ページへ遷移完了")

	var errorVisible, buttonVisible bool
	slog.Debug("注文書エラー要素とダウンロードボタンの存在確認")

	// エラー表示とダウンロードボタンの有無を同時に確認
	err := chromedp.Run(ctx,
		chromedp.Tasks{
			chromedp.Evaluate(fmt.Sprintf(`document.querySelector("%s") !== null`, constants.ErrorSelector), &errorVisible),
			chromedp.Evaluate(fmt.Sprintf(`document.querySelector("%s") !== null`, constants.ButtonDownloadSelector), &buttonVisible),
		},
	)
	if err != nil {
		slog.Error("要素の検出に失敗", "error", err)
		return fmt.Errorf("要素の検出に失敗: %w", err)
	}

	if errorVisible {
		slog.Warn("注文書が存在しない")
		return fmt.Errorf("該当の注文書が見つかりません")
	}
	if !buttonVisible {
		slog.Warn("ダウンロードボタンが存在しない")
		return fmt.Errorf("ダウンロードボタンがありません")
	}

	// ダウンロードの進捗監視
	slog.Debug("ダウンロード準備開始")

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
	slog.Debug("ダウンロードクリック開始")
	if err := chromedp.Run(ctx,
		browser.
			SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllowAndName).
			WithDownloadPath(downloadPath).
			WithEventsEnabled(true),
		chromedp.Click(constants.ButtonDownloadSelector, chromedp.ByQuery),
	); err != nil {
		slog.Error("ダウンロードクリックに失敗", "error", err)
		return fmt.Errorf("ダウンロードクリックに失敗: %w", err)
	}
	slog.Debug("ダウンロードクリック完了、完了待機中")

	// 完了待ち
	var guid string
	select {
	case guid = <-done:
		slog.Debug("ダウンロード完了", slog.String("guid", guid), slog.String("path", downloadPath))
	case <-ctx.Done():
		slog.Error("ダウンロードがタイムアウトまたは中断", "error", ctx.Err())
		return fmt.Errorf("ダウンロードがタイムアウトまたは中断された: %w", ctx.Err())
	}

	// ファイルをリネーム
	slog.Debug("ダウンロードファイルのリネーム処理開始")
	filename := kikaku + ".csv"

	files, err := os.ReadDir(downloadPath)
	if err != nil {
		slog.Error("ダウンロードディレクトリの読み取りに失敗", "error", err)
		return fmt.Errorf("ダウンロードディレクトリの読み取りに失敗: %w", err)
	}

	for _, f := range files {
		if !f.IsDir() && strings.HasPrefix(f.Name(), guid) {
			oldPath := filepath.Join(downloadPath, f.Name())
			newPath := filepath.Join(downloadPath, filename)
			if err := os.Rename(oldPath, newPath); err != nil {
				slog.Error("ファイルのリネームに失敗", "error", err)
				return fmt.Errorf("ファイルのリネームに失敗: %w", err)
			}
			slog.Debug("ファイルのリネーム完了", slog.String("old", oldPath), slog.String("new", newPath))
			break
		}
	}

	return nil
}
