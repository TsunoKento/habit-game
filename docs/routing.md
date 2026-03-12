# API / ルーティング

初期版はサーバレンダリングを前提とする。

| Method | Path | 用途 |
|---|---|---|
| GET | `/` | ダッシュボード表示 |
| POST | `/habits/:id/check` | 習慣の当日達成記録 |
| GET | `/history` | 履歴表示 |
| GET | `/settings` | 設定表示 |

※ Go の標準 `net/http` を使う場合はルーティング実装に合わせてパス設計を調整する。
