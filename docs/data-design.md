# データ設計

## habits テーブル

習慣マスタを管理する。

| カラム名 | 型 | 説明 |
|---|---|---|
| id | INTEGER | 主キー |
| name | TEXT | 習慣名 |
| exp_per_done | INTEGER | 1回達成時の基本経験値 |
| created_at | DATETIME | 作成日時 |

---

## daily_records テーブル

日ごとの達成記録を管理する。達成した日のみレコードを作成し、未達成日のレコードは保存しない。

| カラム名 | 型 | 説明 |
|---|---|---|
| id | INTEGER | 主キー |
| habit_id | INTEGER | habits.id への参照 |
| date | DATE | 達成した対象日 |
| created_at | DATETIME | 作成日時 |

### 制約

- `habit_id + date` は一意制約を持たせる
- 同一習慣の同日重複登録を防ぐ

### 備考

- `done` カラムは持たない。レコードの存在そのものが達成を意味する
- レコードが存在しない日は未達成とみなして計算する
