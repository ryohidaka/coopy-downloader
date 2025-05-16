/*
Copyright © 2025 ryohidaka<39184410+ryohidaka@users.noreply.github.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"log"
	"log/slog"
	"os"

	"github.com/ryohidaka/coopy-downloader/internal/browser"
	"github.com/spf13/cobra"
)

type Params struct {
	LoginID   string // ログインID
	Password  string // パスワード
	Kikaku    string // 企画回
	OutputDir string // 出力先ディレクトリ
}

var (
	loginId   string
	password  string
	kikaku    string
	outputDir string
	noSandbox bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "coopy-downloader",
	Short: "コープこうべの宅配の注文書(CSV)をダウンロードするCLIツール",
	Run:   run,
}

// CLI 実行時に呼び出されるメイン処理。
//
// 引数:
//   - cmd: 実行された cobra コマンド
//   - args: コマンドライン引数
//
// 返り値:
//   - なし（エラーが発生した場合は log.Fatal により強制終了）
func run(cmd *cobra.Command, args []string) {
	// chromedp用のコンテキストを作成（no-sandboxオプションを指定）
	ctx, cancel := browser.CreateChromedpContext(noSandbox)
	defer cancel()

	// フラグからログイン情報を構造体に格納
	params := Params{
		LoginID:   loginId,
		Password:  password,
		Kikaku:    kikaku,
		OutputDir: outputDir,
	}

	// ログイン処理を実行し、失敗した場合はエラーを出力して終了
	if err := browser.Login(ctx, params.LoginID, params.Password); err != nil {
		slog.Error("ログイン失敗", "error", err)
	}

	// 注文ページからダウンロード
	if err := browser.DownloadOrder(ctx, params.Kikaku, params.OutputDir); err != nil {
		log.Fatalf("注文書のダウンロード失敗: %v", err)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.Flags().StringVarP(&loginId, "login-id", "i", "", "ログインID（必須）")
	rootCmd.Flags().StringVarP(&password, "password", "p", "", "パスワード（必須）")
	rootCmd.Flags().StringVarP(&kikaku, "kikaku", "k", "", "企画回")
	rootCmd.Flags().StringVarP(&outputDir, "output-dir", "o", ".", "ダウンロード先ディレクトリ")
	rootCmd.Flags().BoolVar(&noSandbox, "no-sandbox", false, "Chromeに --no-sandbox オプションを付けて起動")

	rootCmd.MarkFlagRequired("login-id")
	rootCmd.MarkFlagRequired("password")
	rootCmd.MarkFlagRequired("kikaku")
}
