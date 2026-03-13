# CLAUDE.md

このファイルは Claude Code がこのリポジトリで作業する際のガイドラインです。

---

## プロジェクト概要

**Habit Growth Tracker** — 習慣をゲーム感覚で継続・可視化するための個人用 Web アプリ。

---

## 技術スタック

- **言語**: Go
- **Web サーバ**: `net/http`
- **テンプレート**: `html/template`
- **DB**: SQLite
- **DB アクセス**: `database/sql`
- **フロントエンド**: HTML / CSS（JavaScript は最小限）

---

## ディレクトリ構成

```
habit-game/
├── cmd/app/main.go
├── internal/
│   ├── handler/       # HTTP ハンドラ
│   ├── service/       # ビジネスロジック
│   ├── repository/    # DB アクセス
│   ├── model/         # データ構造
│   └── db/            # DB 接続・マイグレーション
├── templates/         # HTML テンプレート
├── static/css/        # スタイルシート
├── migrations/        # SQL マイグレーションファイル
├── go.mod
└── docs/specification.md
```

---

## 開発フロー

### TDD（テスト駆動開発）

**Red → Green → Refactor** のサイクルで実装する。

1. **Red**: 失敗するテストを先に書く
2. **Green**: テストが通る最小限の実装をする
3. **Refactor**: コードを整理する

テストなしで実装を進めない。

### テスト方針

- **`internal/service/`**: ビジネスロジックのユニットテストを必ず書く
- **`internal/repository/`**: SQLite を使ったインテグレーションテストを書く（モックは使わない）
- **`internal/handler/`**: `httptest` を使ったハンドラテストを書く
- テストファイルは `*_test.go` の命名規則に従う

```bash
# テスト実行
go test ./...

# 特定パッケージのみ
go test ./internal/service/...

# カバレッジ確認
go test -cover ./...

# ビルド（成果物は cmd/app/app に固定）
go build -o cmd/app/app ./cmd/app/

# サーバ起動
./cmd/app/app
```

---

## コーディング規約

### アーキテクチャ

- `handler` → `service` → `repository` の依存方向を守る
- 逆方向の依存を作らない（例: service が handler を参照しない）
- 各層はインターフェースで結合する（テスタビリティのため）

### 命名

- Go の標準的な命名規則に従う（`camelCase`、`PascalCase`）
- テーブル名・カラム名は `snake_case`

### エラーハンドリング

- エラーは握りつぶさない
- `fmt.Errorf("...: %w", err)` でラップしてコンテキストを付与する

---

## 仕様参照

| 内容 | ファイル |
|---|---|
| 機能・画面要件 | `docs/specification.md` |
| DB スキーマ | `docs/data-design.md` |
| 経験値・レベル・連続日数の計算 | `docs/calculation.md` |
| API / ルーティング | `docs/routing.md` |

