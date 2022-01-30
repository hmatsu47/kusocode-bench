# kusocode-bench

個人研究用クソコード置き場（スピードアップコンテストのネタにするようなやつ）のベンチマークプログラム

- クソコード置き場にあるサンプル用のベンチマークプログラムです
- Go で書かれていますが Go らしい書き方をしていない部分が多いです（たぶん）
- 今のところクソコード 3 用に合わせてありますが、クソコード 1 用のパスに書き換えればクソコード 1 でもそのまま動きます
- クソコード 2 でも動きますが、ページネーションに対応していないので不完全です
- ベンチマーク実行環境には各クソコードと同じデータベースを用意するとともに、以下のテーブルを追加します

```sql:team
CREATE TABLE `team` (
  `id` int(11) NOT NULL,
  `name` varchar(40) NOT NULL,
  `ip_address` varchar(15) NOT NULL,
  `score` int(11) NOT NULL DEFAULT 0,
  `exec_flag` int(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `ip_address_UNIQUE` (`ip_address`),
  UNIQUE KEY `name_UNIQUE` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

- `team` テーブルには参加チームの情報を登録します
- ベンチマーク対象のポート番号は 8080 で直書きしてありますが、必要に応じて [bench.go](service/bench.go) を「:8080」で検索して書き換えてください
  - コードから外して DB に IP アドレスと一緒に入れても OK
