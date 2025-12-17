# Price Compare - システム構成説明書

## 目次

1. [システム概要](#システム概要)
2. [アーキテクチャ](#アーキテクチャ)
3. [技術スタック](#技術スタック)
4. [データフロー](#データフロー)
5. [データベーススキーマ](#データベーススキーマ)
6. [プロバイダアーキテクチャ](#プロバイダアーキテクチャ)
7. [商品識別子ベースの統合機能](#商品識別子ベースの統合機能)
8. [ジョブ処理フロー](#ジョブ処理フロー)
9. [コンプライアンス機能](#コンプライアンス機能)
10. [APIエンドポイント](#apiエンドポイント)
11. [フロントエンド構成](#フロントエンド構成)
12. [インフラストラクチャ](#インフラストラクチャ)

---

## システム概要

Price Compareは、複数のECサイトから商品情報を収集し、価格比較を行うWebアプリケーションです。公式APIを優先的に使用し、コンプライアンスを重視した設計となっています。

### 主な特徴

- **公式API優先**: Walmart Data API、Amazon Product Advertising API 5.0を使用
- **コンプライアンス重視**: robots.txtチェック、レートリミット、監査ログを実装
- **非同期処理**: Asynqを使用したバックグラウンドジョブ処理
- **商品統合**: 識別子（itemId、ASIN等）ベースで同一商品を統合
- **リアルタイム検索禁止**: DB検索のみ（外部APIへの直接検索は行わない）

---

## アーキテクチャ

### 全体構成図

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend (Next.js)                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │   Home   │  │  Search │  │  Compare │  │ Admin/Jobs│   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ HTTP/REST API
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Backend API (Go/Fiber)                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Handlers   │  │   Jobs       │  │  Providers   │     │
│  │   (HTTP)     │  │  (Asynq)     │  │  (API/Scrape) │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│         │                  │                  │             │
│         └──────────────────┼──────────────────┘             │
│                            │                                │
│                   ┌────────▼────────┐                       │
│                   │  Repository     │                       │
│                   │  (Data Access)  │                       │
│                   └────────┬────────┘                       │
└────────────────────────────┼──────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  PostgreSQL  │    │    Redis     │    │  HTTP Client │
│  (Database)  │    │ (Cache/Queue)│    │ (Compliance) │
└──────────────┘    └──────────────┘    └──────────────┘
```

### レイヤー構成

1. **Presentation Layer (Frontend)**
   - Next.js 16 App Router
   - TypeScript + Tailwind CSS
   - ユーザーインターフェース

2. **Application Layer (Backend API)**
   - HTTP Handlers (Fiber)
   - Job Processors (Asynq)
   - Business Logic

3. **Domain Layer**
   - Models (Product, Offer, ProductIdentifier, SourceProduct)
   - Providers (Walmart, Amazon, etc.)
   - Shipping Calculator

4. **Infrastructure Layer**
   - Repository (Database Access)
   - HTTP Client (Compliance Features)
   - Configuration Management

5. **Data Layer**
   - PostgreSQL (Primary Database)
   - Redis (Cache & Queue)

---

## 技術スタック

### Frontend

- **Framework**: Next.js 16.0.10 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **UI Components**: shadcn/ui
- **State Management**: React Hooks

### Backend

- **Language**: Go 1.22+
- **Web Framework**: Fiber v2.52.4
- **ORM**: 標準database/sql + カスタムRepository
- **Job Queue**: Asynq (Redis-based)
- **Logging**: Zap (structured logging)
- **Audit Logging**: slog (JSON format)

### Database & Cache

- **Primary Database**: PostgreSQL 16
- **Cache/Queue**: Redis 7
- **Migration Tool**: カスタムマイグレーション（Go）

### Infrastructure

- **Containerization**: Docker + Docker Compose
- **Service Discovery**: Docker Compose networking

### External APIs

- **Walmart Data API**: RapidAPI経由
- **Amazon Product Advertising API**: PA-API 5.0

---

## データフロー

### 1. 価格更新ジョブの実行フロー

```
User (Admin) → POST /api/admin/jobs/fetch_prices
    ↓
Handler: FetchPrices
    ↓
Asynq: Enqueue Job
    ↓
Redis Queue
    ↓
Job Processor: HandleFetchPrices
    ↓
Provider Manager: Get Provider (walmart/amazon)
    ↓
Provider: Search(query)
    ↓
Provider: FetchOffers(product)
    ↓
Repository: Upsert Product/Offer
    ↓
PostgreSQL: Save Data
```

### 2. 商品検索フロー

```
User → GET /api/search?query=keyword
    ↓
Handler: Search
    ↓
Repository: ProductRepository.Search
    ↓
PostgreSQL: SELECT (Full-text search)
    ↓
Response: JSON (Product List)
    ↓
Frontend: Display Results
```

### 3. 価格比較フロー

```
User → GET /api/products/:id/compare
    ↓
Handler: CompareProductOffers
    ↓
Repository: OfferRepository.FindByProductID
    ↓
PostgreSQL: SELECT with ORDER BY
    ↓
Response: JSON (Sorted Offers)
    ↓
Frontend: Display Comparison Table
```

---

## データベーススキーマ

### テーブル構成

#### 1. `products` - 商品マスタ

```sql
CREATE TABLE products (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    brand TEXT,
    model TEXT,
    image_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);
```

**役割**: 統合された商品情報を保持。複数のソースから取得された同一商品は1つのレコードに統合される。

#### 2. `product_identifiers` - 商品識別子

```sql
CREATE TABLE product_identifiers (
    id UUID PRIMARY KEY,
    product_id UUID REFERENCES products(id),
    type TEXT NOT NULL,  -- 'itemId', 'ASIN', 'UPC', 'EAN', etc.
    value TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    UNIQUE (type, value)
);
```

**役割**: 商品の識別子（Walmart itemId、Amazon ASIN等）を保存。識別子ベースで商品を統合する際に使用。

#### 3. `source_products` - ソース別商品情報

```sql
CREATE TABLE source_products (
    id UUID PRIMARY KEY,
    product_id UUID REFERENCES products(id),
    provider TEXT NOT NULL,  -- 'walmart', 'amazon', etc.
    source_id TEXT NOT NULL,  -- Provider-specific ID
    url TEXT NOT NULL,
    title TEXT,
    brand TEXT,
    image_url TEXT,
    raw_json JSONB,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    UNIQUE (provider, source_id)
);
```

**役割**: 各プロバイダから取得した商品情報を保持。1つの商品（products）に対して複数のソース（source_products）が紐づく。

#### 4. `offers` - オファー（価格情報）

```sql
CREATE TABLE offers (
    id UUID PRIMARY KEY,
    product_id UUID REFERENCES products(id),
    source TEXT NOT NULL,  -- 'walmart', 'amazon', etc.
    seller TEXT NOT NULL,
    price_amount INTEGER NOT NULL,  -- cents
    currency TEXT NOT NULL,
    shipping_to_us_amount INTEGER NOT NULL,  -- cents
    total_to_us_amount INTEGER NOT NULL,  -- cents
    est_delivery_days_min INTEGER,
    est_delivery_days_max INTEGER,
    in_stock BOOLEAN NOT NULL,
    url TEXT,
    fetched_at TIMESTAMP WITH TIME ZONE,
    fee_amount INTEGER NOT NULL DEFAULT 0,  -- cents
    tax_amount INTEGER,  -- cents
    availability_status TEXT,  -- 'in_stock', 'out_of_stock', etc.
    estimated_delivery_date TIMESTAMP WITH TIME ZONE,
    price_updated_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);
```

**役割**: 各商品の価格情報、在庫状況、配送情報を保持。1つの商品に対して複数のオファーが存在する。

### リレーションシップ

```
products (1) ──< (N) product_identifiers
products (1) ──< (N) source_products
products (1) ──< (N) offers
```

### インデックス

- `products`: `updated_at` (検索時のソート用)
- `product_identifiers`: `(type, value)` (識別子検索用)
- `offers`: `product_id`, `price_updated_at` (検索・ソート用)

---

## プロバイダアーキテクチャ

### Provider Interface

すべてのプロバイダは以下のインターフェースを実装します：

```go
type Provider interface {
    // Search: キーワード検索で商品候補を取得
    Search(ctx context.Context, query string) ([]ProductCandidate, error)
    
    // FetchOffers: 商品の詳細情報（価格、在庫等）を取得
    FetchOffers(ctx context.Context, product *models.Product) ([]*models.Offer, error)
}
```

### ProductCandidate

```go
type ProductCandidate struct {
    Title      string   // 商品名
    Brand      *string  // ブランド
    Model      *string  // モデル
    ImageURL   *string  // 画像URL
    Source     string   // プロバイダ名
    Identifier *string  // 識別子（itemId, ASIN等）
    SourceURL  *string  // 商品URL
}
```

### 実装済みプロバイダ

#### 1. Walmart Official Provider

- **API**: Walmart Data API (RapidAPI経由)
- **認証**: X-RapidAPI-Key, X-RapidAPI-Host
- **識別子**: itemId (URLから抽出)
- **レートリミット**: 5 RPS (デフォルト)

**実装ファイル**: `apps/api/internal/providers/walmart_official.go`

#### 2. Amazon Official Provider

- **API**: Amazon Product Advertising API 5.0
- **認証**: AWS Signature Version 4
- **識別子**: ASIN
- **レートリミット**: 1 RPS (デフォルト)

**実装ファイル**: `apps/api/internal/providers/amazon_official.go`

#### 3. Demo Provider (開発用)

- **データ**: モックデータ
- **用途**: テスト・開発
- **有効化**: `ENABLE_DEMO_PROVIDERS=true`

**実装ファイル**: `apps/api/internal/providers/demo.go`

#### 4. Public HTML Provider (開発用)

- **データ**: `/samples`配下のHTMLファイル
- **用途**: スクレイピングロジックのテスト
- **有効化**: `ENABLE_DEMO_PROVIDERS=true`

**実装ファイル**: `apps/api/internal/providers/public_html.go`

#### 5. Live Provider (実装済み、使用注意)

- **データ**: 実際の外部サイトから取得
- **コンプライアンス**: robots.txtチェック、レートリミット適用
- **有効化**: `ALLOW_LIVE_FETCH=true` (自己責任)

**実装ファイル**: `apps/api/internal/providers/live.go`

### プロバイダ登録

`apps/api/cmd/server/main.go`でプロバイダを登録：

```go
// 公式APIプロバイダ（常に有効、APIキーが設定されている場合のみ）
providerManager.Register("walmart", walmartProvider)
providerManager.Register("amazon", amazonProvider)

// 開発用プロバイダ（ENABLE_DEMO_PROVIDERS=trueの場合のみ）
if enableDemoProviders {
    providerManager.Register("demo", demoProvider)
    providerManager.Register("public_html", publicHTMLProvider)
}
```

---

## 商品識別子ベースの統合機能

### 概要

同じ商品を複数のソース（Walmart、Amazon等）から取得した場合、識別子（itemId、ASIN等）を使って1つの商品に統合します。

### 統合ロジック

#### 1. 識別子の抽出

- **Walmart**: URLから`itemId`を抽出
  - 例: `https://www.walmart.com/ip/.../5461164337` → `itemId: 5461164337`
- **Amazon**: APIレスポンスから`ASIN`を取得

#### 2. 商品検索・統合フロー

```
Job Processor: processCandidate
    ↓
1. Identifierで既存商品を検索
   └─> ProductIdentifierRepository.FindByTypeAndValue
    ↓
2. 見つかった場合: 既存商品にオファーを追加
   見つからない場合: 新規商品を作成
    ↓
3. 識別子を保存
   └─> ProductIdentifierRepository.Create
```

#### 3. 実装箇所

- **識別子抽出**: `apps/api/internal/providers/walmart_official.go`の`extractWalmartItemId`
- **統合ロジック**: `apps/api/internal/jobs/processor.go`の`processCandidate`
- **識別子保存**: `apps/api/internal/repository/product_identifier.go`

### 統合の優先順位

1. **識別子ベース**: 識別子が一致する場合は統合（推奨）
2. **タイトルベース**: 識別子がない場合はタイトルで検索（フォールバック）

### 注意事項

- 識別子が一致しない場合は統合しない（誤統合防止）
- タイトルが少し違っても、識別子が一致すれば統合される
- 例: "JBL Tune 520BT - White" と "JBL Tune 520BT - Purple" は別商品として扱われる（itemIdが異なるため）

---

## ジョブ処理フロー

### Asynq による非同期処理

#### 1. ジョブの登録

```go
// Handler: POST /api/admin/jobs/fetch_prices
payload := FetchPricesPayload{Source: "walmart"}
task := asynq.NewTask(TypeFetchPrices, payload)
asynqClient.Enqueue(task)
```

#### 2. ジョブの処理

```go
// Processor: HandleFetchPrices
1. ペイロードを解析
2. プロバイダを取得
3. 検索クエリを実行（固定クエリ: headphones, laptop, etc.)
4. 各商品候補を処理（processCandidate）
   - 商品の検索・作成
   - オファーの取得・保存
```

#### 3. エラーハンドリング

- レートリミット（429）: 5秒待機してリトライ
- タイムアウト: エラーログを記録してスキップ
- その他のエラー: ログに記録して続行

### 固定検索クエリ

現在、以下の固定クエリで商品を取得しています：

```go
queries := []string{
    "headphones", "laptop", "smartphone", 
    "tablet", "watch", "minecraft", 
    "game", "toy", "book"
}
```

**注意**: ユーザーが検索したキーワードでジョブを実行する機能は未実装（DB検索のみ）。

---

## コンプライアンス機能

### 1. robots.txt チェック

**実装**: `apps/api/internal/compliance/robots/checker.go`

- 外部URLアクセス前に`/robots.txt`を取得
- User-Agentとパスが許可されているかチェック
- Disallowされている場合はアクセスをブロック
- Redisにキャッシュ（TTL: 24時間）

**使用箇所**: `internal/httpclient.Client`経由で自動適用

### 2. レートリミット

**実装**: `apps/api/internal/ratelimit/manager.go`

- プロバイダごとに独立したレートリミッター
- トークンバケット方式（`golang.org/x/time/rate`）
- 環境変数でRPSとバースト値を設定可能

**設定例**:
- `PROVIDER_RATE_LIMIT_WALMART_RPS=5`
- `PROVIDER_RATE_LIMIT_AMAZON_RPS=1`
- `PROVIDER_RATE_LIMIT_BURST=2`

### 3. 監査ログ

**実装**: `apps/api/internal/audit/log.go`

- すべての外部HTTPリクエストをJSON形式で記録
- 記録内容:
  - タイムスタンプ、プロバイダ、URL
  - ステータスコード、処理時間
  - robots.txtの許可/拒否状態
  - リトライ回数、エラー情報

**出力**: stdout (JSON形式)

### 4. ALLOW_LIVE_FETCH 制御

- **デフォルト**: `false` (外部URLアクセスをブロック)
- **設定**: `ALLOW_LIVE_FETCH=true`で有効化
- **注意**: 自己責任で、許可されたサイトのみにアクセス

### 5. HTTP Client (Compliance Wrapper)

**実装**: `apps/api/internal/httpclient/client.go`

すべての外部HTTPアクセスは`httpclient.Client`経由で行います：

```go
client := httpclient.New(config, logger, redisClient)
resp, err := client.Get(ctx, provider, url)
```

**自動適用機能**:
- robots.txtチェック
- レートリミット
- リトライ（指数バックオフ）
- 監査ログ
- タイムアウト制御

---

## APIエンドポイント

### 公開エンドポイント

#### 1. `GET /health`
ヘルスチェック

**Response**:
```json
{"status": "ok"}
```

#### 2. `GET /api/search?query=<keyword>`
商品検索

**Response**:
```json
{
  "products": [
    {
      "id": "uuid",
      "title": "商品名",
      "brand": "ブランド",
      "image_url": "https://...",
      ...
    }
  ]
}
```

#### 3. `GET /api/products/:id`
商品詳細取得

#### 4. `GET /api/products/:id/offers`
商品のオファー一覧

#### 5. `GET /api/products/:id/compare`
価格比較（ソート済みオファー）

**ソートオプション**:
- 総額が安い順
- 納期が早い順
- 更新日時が新しい順
- 在庫あり優先

#### 6. `POST /api/resolve-url`
URLから商品を解決（ASIN/itemId抽出）

### 管理エンドポイント

#### 7. `POST /api/admin/jobs/fetch_prices`
価格更新ジョブの実行

**Request Body**:
```json
{
  "source": "walmart" | "amazon" | "all"
}
```

**Response**:
```json
{
  "job_id": "uuid",
  "status": "enqueued",
  "source": "walmart"
}
```

---

## フロントエンド構成

### ページ構成

#### 1. `/` (Home)
- アプリケーションの説明
- 検索・比較へのリンク

#### 2. `/search` (商品検索)
- キーワード検索
- URL入力（ASIN/itemId抽出）
- 検索結果のカード表示

#### 3. `/compare?productId=<id>` (価格比較)
- オファー比較テーブル
- ソート機能（総額、納期、更新日時、在庫）
- 各オファーの詳細情報表示

#### 4. `/admin/jobs` (管理画面)
- プロバイダ選択
- ジョブ実行
- ジョブID表示

### コンポーネント構成

```
components/
├── ui/              # shadcn/ui コンポーネント
│   ├── button.tsx
│   ├── card.tsx
│   ├── table.tsx
│   └── ...
└── ...
```

### API通信

**実装**: `apps/web/lib/api.ts`

```typescript
// 検索
export async function searchProducts(query: string)

// 商品取得
export async function getProduct(id: string)

// オファー取得
export async function getOffers(id: string)

// 価格比較
export async function compareProduct(id: string, sortBy?: string)

// ジョブ実行
export async function fetchPrices(source?: string)
```

---

## インフラストラクチャ

### Docker Compose構成

#### サービス一覧

1. **postgres** (PostgreSQL 16)
   - ポート: `5433:5432`
   - データ永続化: `postgres_data` volume

2. **redis** (Redis 7)
   - ポート: `6379:6379`
   - パスワード保護: `password`
   - データ永続化: `redis_data` volume

3. **api** (Go API Server)
   - ポート: `8080:8080`
   - 依存: postgres, redis
   - 自動マイグレーション実行

4. **web** (Next.js Frontend)
   - ポート: `3000:3000`
   - 依存: api
   - ホットリロード対応

### 環境変数

主要な環境変数は`docker-compose.yml`に定義されています。

**基本設定**:
- データベース接続情報
- Redis接続情報
- APIポート

**API設定**:
- `WALMART_API_KEY`
- `AMAZON_ACCESS_KEY`, `AMAZON_SECRET_KEY`, `AMAZON_ASSOCIATE_TAG`

**コンプライアンス設定**:
- `ALLOW_LIVE_FETCH`
- `PROVIDER_RATE_LIMIT_*_RPS`
- `ROBOTS_CACHE_TTL_HOURS`

**開発用設定**:
- `ENABLE_DEMO_PROVIDERS`
- `NEXT_PUBLIC_ENABLE_DEMO_PROVIDERS`

### データ永続化

- **PostgreSQL**: `postgres_data` volume
- **Redis**: `redis_data` volume

### ヘルスチェック

- **PostgreSQL**: `pg_isready`
- **Redis**: `redis-cli ping`

---

## まとめ

### 設計の特徴

1. **公式API優先**: スクレイピングよりも公式APIを優先
2. **コンプライアンス重視**: robots.txt、レートリミット、監査ログを実装
3. **非同期処理**: 重い処理はバックグラウンドジョブで実行
4. **商品統合**: 識別子ベースで同一商品を統合
5. **プラガブル設計**: プロバイダを追加・削除しやすい

### 今後の拡張可能性

- 新しいプロバイダの追加（eBay、Target等）
- 商品識別子の拡張（UPC、EAN、JAN等）
- キャッシュ戦略の改善
- リアルタイム価格更新（WebSocket等）
- 画像検索機能の実装

---

**最終更新**: 2025年12月17日

