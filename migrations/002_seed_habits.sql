-- 初回起動時のみ実行される。起動ごとに保証したい場合は INSERT OR IGNORE をアプリ層で実行すること。
INSERT INTO habits (id, name, exp_per_done) VALUES
    (1, '早起き',   10),
    (2, '英語学習', 10),
    (3, '運動',     10);
