# coopy-downloader

コープこうべの宅配の注文書(CSV)をダウンロードする CLI ツール

## インストール

```bash
go install github.com/ryohidaka/coopy-downloader@latest
```

## 使用例

```bash
coopy-downloader --help

コープこうべの宅配の注文書(CSV)をダウンロードするCLIツール

Usage:
  coopy-downloader [flags]
  coopy-downloader [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  kikaku      お届け日から企画回を取得する

Flags:
  -h, --help                help for coopy-downloader
  -k, --kikaku string       企画回
  -i, --login-id string     ログインID（必須）
      --no-sandbox          Chromeに --no-sandbox オプションを付けて起動
  -o, --output-dir string   ダウンロード先ディレクトリ (default ".")
  -p, --password string     パスワード（必須）

Use "coopy-downloader [command] --help" for more information about a command.
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
