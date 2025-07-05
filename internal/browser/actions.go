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

	// ダウンロードの挙動設定を事前に有効化する
	if err := chromedp.Run(ctx,
		browser.
			SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllowAndName).
			WithDownloadPath(downloadPath).
			WithEventsEnabled(true),
	); err != nil {
		slog.Error("ダウンロード挙動の設定に失敗", "error", err)
		return fmt.Errorf("ダウンロード設定失敗: %w", err)
	}

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

	// ダウンロードイベントを監視
	var guid string
	var suggestedFilename string
	guidChan := make(chan string, 1)

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if e, ok := ev.(*browser.EventDownloadWillBegin); ok {
			slog.Info("ダウンロード開始検知", slog.String("guid", e.GUID), slog.String("filename", e.SuggestedFilename))
			guid = e.GUID
			suggestedFilename = e.SuggestedFilename
			guidChan <- e.GUID
			close(guidChan)
		}
	})

	// ダウンロードボタンをクリック
	slog.Debug("ダウンロードクリック開始")
	if err := chromedp.Run(ctx,
		chromedp.Click(constants.ButtonDownloadSelector, chromedp.ByQuery),
	); err != nil {
		slog.Error("ダウンロードクリックに失敗", "error", err)
		return fmt.Errorf("ダウンロードクリックに失敗: %w", err)
	}
	slog.Debug("ダウンロードクリック完了、ダウンロード開始待機中")

	// GUID受信を待つ
	select {
	case <-guidChan:
		slog.Debug("GUID受信完了", slog.String("guid", guid), slog.String("suggested", suggestedFilename))
	case <-ctx.Done():
		slog.Error("ダウンロード開始イベントを受信できずタイムアウト", "error", ctx.Err())
		return fmt.Errorf("ダウンロード開始イベントが発生しませんでした: %w", ctx.Err())
	}

	// ダウンロード完了をポーリングで確認（*.crdownloadが消えるまで）
	slog.Debug("ファイルダウンロード完了待機中")
	targetPath := filepath.Join(downloadPath, suggestedFilename)

	timeout := time.After(2 * time.Minute)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			slog.Error("ダウンロードファイルがタイムアウト", slog.String("filename", suggestedFilename))
			return fmt.Errorf("ダウンロードファイルが見つかりません: %s", suggestedFilename)
		case <-tick:
			if _, err := os.Stat(targetPath); err == nil {
				if !strings.HasSuffix(targetPath, ".crdownload") {
					goto FOUND
				}
			}
		}
	}

FOUND:
	// ファイルをkikaku名にリネーム
	newPath := filepath.Join(downloadPath, kikaku+".csv")
	if err := os.Rename(targetPath, newPath); err != nil {
		slog.Error("ファイルのリネームに失敗", "error", err)
		return fmt.Errorf("ファイルのリネームに失敗: %w", err)
	}
	slog.Debug("ファイルのリネーム完了", slog.String("old", targetPath), slog.String("new", newPath))

	return nil
}
