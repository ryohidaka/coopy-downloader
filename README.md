# coopy-downloader

コープこうべの宅配の注文書(CSV)をダウンロードする CLI ツール

## インストール

```bash
go get github.com/ryohidaka/coopy-downloader
```

## 使用例

```bash
coopy-downloader --help

コープこうべの宅配の注文書(CSV)をダウンロードするCLIツール

Usage:
  coopy-downloader [flags]

Flags:
  -h, --help                help for coopy-downloader
  -k, --kikaku string       企画回
  -i, --login-id string     ログインID（必須）
  -o, --output-dir string   ダウンロード先ディレクトリ (default ".")
  -p, --password string     パスワード（必須）

coopy-downloader \
    --login-id <LOGIN_ID> \
    --password <PASSWORD> \
    --kikaku 2025041 \
    --output-dir .output
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
