# API キー設定ガイド

本アプリケーションは、Walmart と Amazon の公式 API を使用して商品情報を取得します。

## Walmart Data API 設定（RapidAPI 経由）

### 必要な環境変数

- `WALMART_API_KEY`: RapidAPI キー（X-RapidAPI-Key の値）（必須）
- `WALMART_API_BASE_URL`: API ベース URL（オプション、デフォルト: `https://walmart-data-api.p.rapidapi.com`）
- `WALMART_API_HOST`: RapidAPI ホスト名（オプション、デフォルト: `walmart-data-api.p.rapidapi.com`）

### 取得方法

1. [RapidAPI](https://rapidapi.com/) にアカウントを作成
2. Walmart Data API を検索してサブスクリプションを選択
3. API キー（X-RapidAPI-Key）を取得して `WALMART_API_KEY` 環境変数に設定
4. RapidAPI ダッシュボードからホスト名を確認して `WALMART_API_HOST` に設定

### API エンドポイント

Walmart Data API（RapidAPI 経由）は以下のエンドポイントを提供します：

- `GET /search?q={query}`: 商品検索
- `GET /product-details?url={walmart_product_url}`: 商品詳細取得
- `GET /category-products?url={category_url}`: カテゴリ商品一覧

### 注意事項

- API キーが設定されていない場合、Walmart プロバイダは自動的に無効化されます
- レートリミットはデフォルトで 5 RPS に設定されています（`PROVIDER_RATE_LIMIT_WALMART_RPS`で変更可能）
- RapidAPI 形式では `X-RapidAPI-Key` と `X-RapidAPI-Host` ヘッダーを使用します
- 実際のエンドポイントとホスト名は RapidAPI ダッシュボードで確認してください

## Amazon Product Advertising API 設定

### 必要な環境変数

- `AMAZON_ACCESS_KEY`: AWS アクセスキー ID（必須）
- `AMAZON_SECRET_KEY`: AWS シークレットキー（必須）
- `AMAZON_ASSOCIATE_TAG`: Amazon アソシエイトタグ（必須）
- `AMAZON_API_ENDPOINT`: API エンドポイント（オプション、デフォルト: `webservices.amazon.com`）
- `AMAZON_API_REGION`: API リージョン（オプション、デフォルト: `us-east-1`）

### 取得方法

1. [Amazon Product Advertising API](https://affiliate-program.amazon.com/gp/advertising/api/dashboard/main.html) にアクセス
2. アソシエイトアカウントを作成（既存の場合はログイン）
3. API 認証情報を取得:
   - Access Key ID → `AMAZON_ACCESS_KEY`
   - Secret Access Key → `AMAZON_SECRET_KEY`
   - Associate Tag → `AMAZON_ASSOCIATE_TAG`

### 注意事項

- すべての環境変数が設定されていない場合、Amazon プロバイダは自動的に無効化されます
- レートリミットはデフォルトで 1 RPS に設定されています（`PROVIDER_RATE_LIMIT_AMAZON_RPS`で変更可能）
- PA-API 5.0 の実装は簡略化されています。本番環境では AWS SDK の使用を推奨します

## 環境変数の設定方法

### Docker Compose の場合

`docker-compose.yml` の `api` サービスの `environment` セクションに追加:

```yaml
environment:
  WALMART_API_KEY: "your-rapidapi-key"
  WALMART_API_BASE_URL: "https://walmart-data-api.p.rapidapi.com"
  WALMART_API_HOST: "walmart-data-api.p.rapidapi.com"
  AMAZON_ACCESS_KEY: "your-amazon-access-key"
  AMAZON_SECRET_KEY: "your-amazon-secret-key"
  AMAZON_ASSOCIATE_TAG: "your-associate-tag"
```

### .env ファイルの場合

プロジェクトルートに `.env` ファイルを作成:

```env
WALMART_API_KEY=your-rapidapi-key
WALMART_API_BASE_URL=https://walmart-data-api.p.rapidapi.com
WALMART_API_HOST=walmart-data-api.p.rapidapi.com
AMAZON_ACCESS_KEY=your-amazon-access-key
AMAZON_SECRET_KEY=your-amazon-secret-key
AMAZON_ASSOCIATE_TAG=your-associate-tag
```

## プロバイダの状態確認

アプリケーション起動時に、ログにプロバイダの有効/無効状態が表示されます:

```
INFO    Walmart API provider enabled
INFO    Amazon API provider disabled (AMAZON_ACCESS_KEY, AMAZON_SECRET_KEY, or AMAZON_ASSOCIATE_TAG not set)
```

## トラブルシューティング

### プロバイダが無効になっている

- 環境変数が正しく設定されているか確認
- 環境変数の名前が正確か確認（大文字小文字を区別）
- アプリケーションを再起動

### API リクエストが失敗する

- API キーが有効か確認
- レートリミットに達していないか確認
- API の利用規約を確認（リクエスト数制限など）

### 商品が取得できない

- 検索クエリが適切か確認
- API のレスポンス形式が変更されていないか確認
- ログを確認してエラーメッセージを確認
