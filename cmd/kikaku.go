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
	"fmt"
	"log/slog"
	"time"

	k "github.com/ryohidaka/coopy-downloader/internal/kikaku"
	"github.com/spf13/cobra"
)

var dateStr string

// kikakuCmd represents the kikaku command
var kikakuCmd = &cobra.Command{
	Use:   "kikaku",
	Short: "お届け日から企画回を取得する",
	Run: func(cmd *cobra.Command, args []string) {
		result, err := k.CalculateKikakuCode(dateStr)
		if err != nil {
			slog.Error("エラー", "error", err)
			return
		}
		fmt.Println(result)
	},
}

func init() {
	rootCmd.AddCommand(kikakuCmd)

	// Here you will define your flags and configuration settings.
	// 今日の日付をデフォルト値として設定
	defaultDate := time.Now().Format("2006-01-02")
	kikakuCmd.Flags().StringVarP(&dateStr, "date", "d", defaultDate, "お届け日 (YYYY-MM-DD)")
}
