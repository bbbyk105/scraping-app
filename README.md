# Price Compare - 価格比較 Web アプリケーション

実運用（製品レベル）を前提にした価格比較 Web アプリケーション。Next.js + Go + Postgres + Redis で構築されています。

## ⚠️ 重要: コンプライアンス/規約

このアプリケーションは以下の規約を厳守するよう設計されています：

- **ログインが必要なページ、購入フロー、個人情報、決済ページへのアクセスは絶対にしない**
- **robots.txt を必ず尊重し、許可されないパスは取得しない**
- **取得対象は「公開されている商品検索結果/商品詳細の最低限情報」または「公式/提携 API」のみ**
- **リアルタイム取得をしない。必ずキャッシュ/DB に保存し、更新頻度を制御する**
- **リクエスト頻度制限（レートリミット）、バックオフ、User-Agent 明示、タイムアウトを実装する**
- **スクレイピングは "許可されたサイトのみのプラガブル実装" とし、デフォルトはダミープロバイダ＋公式 API プロバイダのみ有効にする**

### コンプライアンス機能（実装済み）

本アプリケーションには、外部 HTTP アクセスを行う際のコンプライアンス機能が実装されています：

1. **robots.txt チェック**: 外部 URL アクセス前に、対象サイトの robots.txt を自動チェックし、Disallow されているパスへのアクセスをブロックします。robots.txt はドメインごとに Redis にキャッシュされます（TTL: 24 時間、環境変数で変更可能）。

2. **レートリミット**: プロバイダごとに設定可能なレートリミットを実装しています。デフォルトでは、live プロバイダは 1 RPS、demo/public_html プロバイダは 10 RPS に設定されています。環境変数で各プロバイダの RPS とバースト値を個別に設定できます。

3. **監査ログ**: すべての外部 HTTP リクエストを JSON 形式で監査ログに記録します。ログには、タイムスタンプ、プロバイダ、URL、ステータスコード、robots.txt の許可/拒否状態、リトライ回数などが含まれます。

4. **ALLOW_LIVE_FETCH 制御**: デフォルトでは`ALLOW_LIVE_FETCH=false`となっており、外部 URL へのアクセスはブロックされます。外部サイトにアクセスする場合は、環境変数で`ALLOW_LIVE_FETCH=true`に設定する必要があります（**自己責任で、許可サイトのみにアクセスすること**）。

5. **リトライとバックオフ**: HTTP リクエストが 429（Too Many Requests）や 5xx エラーを返した場合、指数バックオフ＋ jitter で最大 3 回まで自動リトライします。

これらの機能は、`internal/httpclient`パッケージに集約されており、すべてのプロバイダが外部 HTTP アクセスを行う際に自動的に適用されます。

### 本番環境での追加確認事項

本番環境にデプロイする前に、以下を必ず確認してください：

1. **公式 API の優先使用**: 可能な限り公式 API や提携 API を使用する
2. **ALLOW_LIVE_FETCH 設定**: 外部サイトアクセスが必要な場合のみ`ALLOW_LIVE_FETCH=true`に設定し、許可されたサイトのみにアクセスすること
3. **法的確認**: 利用するサイトの利用規約を確認し、法的に問題がないことを確認
4. **監査ログの確認**: 監査ログを定期的に確認し、コンプライアンス違反がないことを確認

## 技術スタック

- **Frontend**: Next.js 16 (App Router) + TypeScript + Tailwind CSS
- **Backend**: Go 1.22+ + Fiber
- **Database**: PostgreSQL 16
- **Cache/Queue**: Redis 7 + Asynq
- **Infrastructure**: Docker + Docker Compose

## プロジェクト構成

```
.
├── apps/
│   ├── api/              # Go API サーバー
│   │   ├── cmd/
│   │   │   ├── server/   # メインサーバー
│   │   │   └── migrate/  # DBマイグレーション
│   │   ├── internal/
│   │   │   ├── config/   # 設定管理
│   │   │   ├── handlers/ # HTTPハンドラー
│   │   │   ├── jobs/     # バックグラウンドジョブ
│   │   │   ├── models/   # データモデル
│   │   │   ├── providers/ # 価格取得プロバイダ
│   │   │   ├── repository/ # DBリポジトリ
│   │   │   └── shipping/  # 送料計算
│   │   └── migrations/    # DBマイグレーションSQL
│   └── web/               # Next.js フロントエンド
│       ├── app/           # App Router
│       └── lib/           # ユーティリティ
├── samples/               # サンプルHTMLファイル
├── docs/                  # ドキュメント
├── docker-compose.yml     # Docker Compose設定
└── Makefile              # 便利コマンド
```

## セットアップと起動

### 前提条件

- Docker & Docker Compose
- Make (オプション)

### 環境変数の設定

環境変数は現在 docker-compose.yml に直接設定されています。本番環境では、以下の環境変数を設定してください：

**基本設定:**

- `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`
- `API_PORT`, `API_HOST`
- `US_SHIP_MODE`, `SHIPPING_FEE_PERCENT`, `FX_USDJPY`
- `USER_AGENT`
- `NEXT_PUBLIC_API_URL` (フロントエンド用)

**公式 API 設定（本番用）:**

- `WALMART_API_KEY`: Walmart API キー
- `AMAZON_ACCESS_KEY`, `AMAZON_SECRET_KEY`, `AMAZON_ASSOCIATE_TAG`: Amazon API 認証情報

**開発用設定（本番では無効化推奨）:**

- `ENABLE_DEMO_PROVIDERS`: 開発用プロバイダ（demo/public_html）を有効化（デフォルト: `false`）

詳細は `docs/API_KEYS.md` を参照してください。

### 起動

```bash
# Docker Composeで全サービスを起動
make dev

# または直接
docker-compose up --build
```

**注意**: 初回起動時は、依存関係のダウンロードとビルドに時間がかかります（数分）。

サービスが起動したら：

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- API Health Check: http://localhost:8080/health

**初回使用時**:

1. API キーを設定（`docs/API_KEYS.md` を参照）
   - Walmart API: `WALMART_API_KEY` を設定
   - Amazon API: `AMAZON_ACCESS_KEY`, `AMAZON_SECRET_KEY`, `AMAZON_ASSOCIATE_TAG` を設定
2. `/admin/jobs` で価格更新ジョブを実行してデータを取得
   - プロバイダを選択（Walmart / Amazon / All）
   - 「価格更新ジョブを実行」ボタンをクリック
3. 数秒〜数十秒待ってから `/search` で商品を検索
4. 検索結果から商品を選択して `/compare` で価格比較

**注意**: API キーが設定されていないプロバイダは自動的に無効化され、選択できません。

### マイグレーション

マイグレーションは自動的に実行されますが、手動で実行する場合：

```bash
cd apps/api
go run cmd/migrate/main.go up    # マイグレーション実行
go run cmd/migrate/main.go down  # ロールバック
```

## 使用方法（価格.com 風フロー）

### 1. 価格データの更新（管理画面）

1. ブラウザで `http://localhost:3000/admin/jobs` にアクセス
2. プロバイダを選択（Walmart / Amazon / All）
3. 「価格更新ジョブを実行」ボタンをクリック
   - 選択したプロバイダから商品情報を取得して DB に保存します
   - API キーが設定されていないプロバイダは選択できません（無効化されています）
4. 数秒〜数十秒待つと、各商品の `price_updated_at` が更新されます

### 2. 商品検索（キーワード / 型番 / JAN）

1. `http://localhost:3000/search` にアクセス
2. 「キーワード検索」タブで、商品名・型番・JAN 等を入力して検索
3. 検索結果カードから商品をクリックすると、`/compare?productId=...` に遷移します

### 3. URL から商品解決 → 価格比較

1. `/search` の「URL 入力」タブを選択
2. 以下の形式のいずれかで URL を入力:
   - `https://www.amazon.com/dp/ASIN`
   - `www.amazon.com/dp/ASIN`
   - `amazon.com/dp/ASIN`
3. 「解析」をクリック
4. 正常な URL の場合:
   - API が ASIN を抽出し、`products` / `product_identifiers` / `source_products` を作成
   - ブラウザが自動的に `/compare?productId=...` にリダイレクト
5. 対応外/不正な URL の場合:
   - 画面上に「URL の形式が正しくありません」「この URL は現在のバージョンでは解析対象外です」等のエラーメッセージが表示されます（コンソールにはスタックトレースを出さない）

### 4. 価格比較画面 `/compare`

1. `/search` から商品をクリック、または `/compare?productId=<product-id>` に直接アクセス
2. 以下の情報でオファーを比較できます:
   - **商品価格 / 送料 / 手数料 / 税 / 合計**
   - **推定到着日数（min-max 日）**
   - **在庫ステータス（在庫あり / 在庫なし）**
   - **更新日時（`price_updated_at`）**
3. 右上のプルダウンから並び替え:
   - 「総額が安い順」
   - 「納期が早い順」
   - 「更新日時が新しい順」
   - 「在庫あり優先」

## API エンドポイント

詳細は `docs/openapi.yaml` を参照してください。

### 主要エンドポイント

- `GET /health` - ヘルスチェック
- `GET /api/search?query=<keyword>` - 商品検索
- `GET /api/products/:id` - 商品詳細取得
- `GET /api/products/:id/offers` - 商品のオファー一覧
- `POST /api/admin/jobs/fetch_prices` - 価格更新ジョブ実行
- `POST /api/image-search` - 画像検索（スタブ実装）

## プロバイダ

現在実装されているプロバイダ：

1. **walmart**: Walmart 公式 API を使用したプロバイダ（本番用）
2. **amazon**: Amazon Product Advertising API 5.0 を使用したプロバイダ（本番用）
3. **live**: 外部サイトからのライブ取得用プロバイダ（実装済み）
4. **demo**: モックデータを使用したテスト用プロバイダ（開発用、`ENABLE_DEMO_PROVIDERS=true`で有効化）
5. **public_html**: `/samples` 配下の HTML ファイルから価格情報を抽出（開発用、`ENABLE_DEMO_PROVIDERS=true`で有効化）

### 公式 API プロバイダ（推奨）

#### Walmart 公式 API

Walmart 公式 API を使用して商品情報を取得します。

- **環境変数**: `WALMART_API_KEY`（必須）
- **設定方法**: `docs/API_KEYS.md` を参照
- **レートリミット**: デフォルト 5 RPS（`PROVIDER_RATE_LIMIT_WALMART_RPS`で変更可能）

#### Amazon Product Advertising API

Amazon Product Advertising API 5.0 を使用して商品情報を取得します。

- **環境変数**: `AMAZON_ACCESS_KEY`, `AMAZON_SECRET_KEY`, `AMAZON_ASSOCIATE_TAG`（すべて必須）
- **設定方法**: `docs/API_KEYS.md` を参照
- **レートリミット**: デフォルト 1 RPS（`PROVIDER_RATE_LIMIT_AMAZON_RPS`で変更可能）

**重要**: API キーが設定されていない場合、該当プロバイダは自動的に無効化されます。

### Live Provider（実際のスクレイピング）

Live Provider は実際の外部サイトから商品情報を取得します。以下の機能が自動的に適用されます：

- **robots.txt チェック**: アクセス前に自動チェック
- **レートリミット**: 1 RPS（デフォルト、環境変数で変更可能）
- **監査ログ**: すべてのリクエストを記録
- **ALLOW_LIVE_FETCH 制御**: デフォルトでは`false`でブロック

**使用方法：**

1. `ALLOW_LIVE_FETCH=true`に設定（自己責任で、許可サイトのみ）
2. `LIVE_PROVIDER_BASE_URL`にスクレイピング対象サイトのベース URL を設定
3. 管理画面で「Live プロバイダ」を選択してジョブを実行

**注意事項：**

- サイトの利用規約を必ず確認してください
- robots.txt を尊重してください（自動チェックされます）
- レートリミットを守ってください（自動適用されます）
- 監査ログを定期的に確認してください

新しいプロバイダを追加するには、`apps/api/internal/providers/interface.go` の `Provider` インターフェースを実装し、`cmd/server/main.go` で登録してください。

**重要**: プロバイダが外部 HTTP アクセスを行う場合は、必ず`internal/httpclient.Client`を使用してください。これにより、robots.txt チェック、レートリミット、監査ログが自動的に適用されます。

## 送料計算

現在は簡易テーブル方式を実装：

- 価格 < $20: 送料 $9.99
- $20 ≤ 価格 < $50: 送料 $14.99
- $50 ≤ 価格: 送料 $19.99
- 手数料: 商品価格の 3%（環境変数で変更可能）
- 為替レート: USD/JPY = 150（環境変数で変更可能）

## 開発

### テスト

```bash
# すべてのテストを実行
cd apps/api
go test ./...

# 特定のパッケージのテストを実行（詳細出力）
go test ./internal/compliance/robots/... -v
go test ./internal/ratelimit/... -v
go test ./internal/httpclient/... -v

# カバレッジを確認
go test ./... -cover
```

**詳細なテスト手順は `TESTING.md` を参照してください。**

### ログ

ログは Zap を使用しています。本番環境では構造化ログとして出力されます。

監査ログは `slog` を使用して JSON 形式で出力されます。外部 HTTP アクセスが発生した場合、以下の情報が記録されます：

- タイムスタンプ、プロバイダ、URL、ステータスコード
- robots.txt の許可/拒否状態、リトライ回数、エラー情報

監査ログを確認するには：

```bash
# Dockerコンテナのログを確認
docker logs pricecompare-api -f | grep "HTTP request audit"
```

## 規約順守の方針

本アプリケーションは、外部サイトへのアクセスを行う際に、以下の規約順守機能を自動的に適用します：

### robots.txt チェック

- **場所**: `internal/compliance/robots/checker.go`
- **動作**: 外部 URL アクセス前に、対象サイトの`/robots.txt`を取得し、現在の User-Agent とパスが許可されているかチェック
- **キャッシュ**: Redis にドメインごとにキャッシュ（デフォルト TTL: 24 時間）
- **失敗時の動作**: robots.txt が取得できない場合や、パースエラーが発生した場合は、安全側に倒してアクセスをブロック

### レートリミット

- **場所**: `internal/ratelimit/manager.go`
- **動作**: プロバイダごとに独立したレートリミッターを管理
- **設定**: 環境変数で各プロバイダの RPS（Requests Per Second）とバースト値を設定可能
- **実装**: `golang.org/x/time/rate`を使用したトークンバケット方式

### 監査ログ

- **場所**: `internal/audit/log.go`
- **出力形式**: JSON 形式の構造化ログ（stdout）
- **記録内容**: タイムスタンプ、プロバイダ、HTTP メソッド、URL、ホスト、パス、ステータスコード、処理時間、User-Agent、robots.txt の許可/拒否状態、リトライ回数、エラー情報

### ALLOW_LIVE_FETCH 制御

- **デフォルト**: `false`（外部 URL アクセスをブロック）
- **設定方法**: 環境変数`ALLOW_LIVE_FETCH=true`で有効化
- **注意**: `true`に設定する場合は、**自己責任で、許可されたサイトのみにアクセスすること**

## 開発用プロバイダの無効化について

本番環境では、開発用プロバイダ（demo/public_html）はデフォルトで無効化されています。

- `ENABLE_DEMO_PROVIDERS=false`（デフォルト）: 開発用プロバイダは使用不可
- `ENABLE_DEMO_PROVIDERS=true`: 開発用プロバイダを有効化（開発・テスト時のみ）

本番環境では、公式 API プロバイダ（walmart/amazon）のみを使用することを推奨します。

## 今後の実装予定（TODO）

- [ ] 画像検索の実装
- [x] レートリミットの実装強化
- [x] robots.txt チェック機能
- [x] 公式 API プロバイダの追加（Walmart/Amazon）
- [ ] 商品識別子（ASIN/UPC/GTIN 等）の自動保存と統合
- [ ] キャッシュ戦略の改善
- [ ] エラーハンドリングの強化
- [ ] E2E テストの追加
- [ ] CI/CD パイプライン
- [ ] 監視とアラートの設定

## 動作確認手順（5 分で完了）

### 1. 環境変数の設定（API キーが無い場合でも動作確認可能）

```bash
# docker-compose.yml を編集してAPIキーを設定
# または .env ファイルを作成
```

### 2. アプリケーションの起動

```bash
docker-compose up --build
```

### 3. ヘルスチェック

```bash
curl http://localhost:8080/health
# 期待される結果: {"status":"ok"}
```

### 4. 価格更新ジョブの実行

1. ブラウザで `http://localhost:3000/admin/jobs` にアクセス
2. プロバイダを選択（API キーが設定されているプロバイダのみ選択可能）
3. 「価格更新ジョブを実行」ボタンをクリック
4. ジョブ ID が表示されることを確認

### 5. 商品検索

1. `http://localhost:3000/search` にアクセス
2. 検索キーワードを入力（例: "headphones", "laptop"）
3. 検索結果が表示されることを確認

### 6. 価格比較

1. 検索結果から商品をクリック
2. `/compare` 画面でオファーが表示されることを確認
3. ソート機能（総額が安い順、納期が早い順等）が動作することを確認

### トラブルシューティング

- **プロバイダが選択できない**: API キーが設定されていない可能性があります。ログを確認してください。
- **商品が検索できない**: 価格更新ジョブを実行してデータを取得してください。
- **エラーが発生する**: `docker logs pricecompare-api -f` でログを確認してください。

詳細は `docs/OPERATION.md` を参照してください。

## ライセンス

このプロジェクトは教育・研究目的で作成されています。実運用に使用する場合は、適切な法的確認とコンプライアンス対応を行ってください。

## 免責事項

表示されている価格は参考情報です。最新の価格や在庫状況は販売元のサイトでご確認ください。このアプリケーションを使用して発生した損害について、開発者は一切の責任を負いません。
