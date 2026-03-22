# habit-game

習慣をゲーム感覚で継続・可視化するための個人用 Web アプリ。

## 起動手順（Docker Compose）

```bash
# 開発サーバを起動
docker compose up

# バックグラウンドで起動
docker compose up -d

# 停止
docker compose down
```

起動後、ブラウザで http://localhost:8080 にアクセス。

## テスト実行

```bash
docker compose run --rm dev go test ./...
```
