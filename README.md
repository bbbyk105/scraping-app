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

### 本番環境での追加実装

本番環境にデプロイする前に、以下を必ず実装してください：

1. **公式 API の優先使用**: 可能な限り公式 API や提携 API を使用する
2. **robots.txt チェック**: 各サイトの robots.txt をチェックし、許可されたパスのみ取得する
3. **レートリミット強化**: より厳密なレートリミットとバックオフ戦略の実装
4. **監査ログ**: すべてのリクエストをログに記録し、コンプライアンス監査に対応
5. **法的確認**: 利用するサイトの利用規約を確認し、法的に問題がないことを確認

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

- `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`
- `API_PORT`, `API_HOST`
- `US_SHIP_MODE`, `SHIPPING_FEE_PERCENT`, `FX_USDJPY`
- `USER_AGENT`
- `NEXT_PUBLIC_API_URL` (フロントエンド用)

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

1. まず `/admin/jobs` で価格更新ジョブを実行してデータを取得してください
2. その後、`/search` で商品を検索できるようになります

### マイグレーション

マイグレーションは自動的に実行されますが、手動で実行する場合：

```bash
cd apps/api
go run cmd/migrate/main.go up    # マイグレーション実行
go run cmd/migrate/main.go down  # ロールバック
```

## 使用方法

### 1. 価格データの更新

1. http://localhost:3000/admin/jobs にアクセス
2. ソース（demo / public_html / all）を選択
3. 「価格更新ジョブを実行」をクリック

これにより、バックグラウンドで価格データが取得・更新されます。

### 2. 商品検索

1. http://localhost:3000/search にアクセス
2. キーワードを入力して検索
3. 検索結果から商品を選択

### 3. 価格比較

1. 検索結果から商品をクリック、または `/compare?productId=<product-id>` に直接アクセス
2. 複数のオファーの価格、送料、合計額、推定到着日数を比較

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

1. **demo**: モックデータを使用したテスト用プロバイダ
2. **public_html**: `/samples` 配下の HTML ファイルから価格情報を抽出

新しいプロバイダを追加するには、`apps/api/internal/providers/interface.go` の `Provider` インターフェースを実装し、`cmd/server/main.go` で登録してください。

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
cd apps/api
go test ./...
```

### ログ

ログは Zap を使用しています。本番環境では構造化ログとして出力されます。

## 今後の実装予定（TODO）

- [ ] 画像検索の実装
- [ ] レートリミットの実装強化
- [ ] robots.txt チェック機能
- [ ] 公式 API プロバイダの追加
- [ ] キャッシュ戦略の改善
- [ ] エラーハンドリングの強化
- [ ] E2E テストの追加
- [ ] CI/CD パイプライン
- [ ] 監視とアラートの設定

## ライセンス

このプロジェクトは教育・研究目的で作成されています。実運用に使用する場合は、適切な法的確認とコンプライアンス対応を行ってください。

## 免責事項

表示されている価格は参考情報です。最新の価格や在庫状況は販売元のサイトでご確認ください。このアプリケーションを使用して発生した損害について、開発者は一切の責任を負いません。
