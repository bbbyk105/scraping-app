# 環境変数一覧

## 現在の状況

現在、環境変数は`docker-compose.yml`に直接設定されています。**ローカル開発環境ではそのまま動作します**。

本番環境や、設定を変更したい場合は、以下の環境変数を設定してください。

---

## 必須設定（本番環境）

### データベース設定

```bash
POSTGRES_HOST=postgres           # デフォルト: localhost
POSTGRES_PORT=5432               # デフォルト: 5432
POSTGRES_USER=pricecompare       # デフォルト: pricecompare
POSTGRES_PASSWORD=password       # ⚠️ 本番環境では必ず変更
POSTGRES_DB=pricecompare         # デフォルト: pricecompare
POSTGRES_SSLMODE=disable         # デフォルト: disable（本番では require 推奨）
```

### Redis設定

```bash
REDIS_HOST=redis                 # デフォルト: localhost
REDIS_PORT=6379                  # デフォルト: 6379
REDIS_PASSWORD=                  # デフォルト: 空（本番環境では設定推奨）
REDIS_DB=0                       # デフォルト: 0
```

### API設定

```bash
API_PORT=8080                    # デフォルト: 8080
API_HOST=0.0.0.0                 # デフォルト: 0.0.0.0
```

---

## 推奨設定（カスタマイズ可能）

### 送料計算設定

```bash
US_SHIP_MODE=TABLE               # デフォルト: TABLE（送料計算方式）
SHIPPING_FEE_PERCENT=3.0         # デフォルト: 3.0（手数料パーセンテージ）
FX_USDJPY=150.0                  # デフォルト: 150.0（USD/JPY為替レート）
```

### User Agent

```bash
USER_AGENT=PriceCompareBot/1.0   # デフォルト: PriceCompareBot/1.0
                                 # ⚠️ 実際のサイトにアクセスする場合は適切に設定
```

### レートリミット（現在未実装、将来用）

```bash
RATE_LIMIT_REQUESTS_PER_SECOND=10  # デフォルト: 10
RATE_LIMIT_BURST=20                # デフォルト: 20
```

---

## フロントエンド設定

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080  # デフォルト: http://localhost:8080
                                           # ⚠️ 本番環境では実際のAPI URLを設定
```

---

## 設定方法

### 方法1: docker-compose.ymlを直接編集（現在の方法）

`docker-compose.yml`の`environment:`セクションを編集します。

### 方法2: .envファイルを使用（推奨）

1. プロジェクトルートに`.env`ファイルを作成：

```bash
# .env
POSTGRES_PASSWORD=your_secure_password
FX_USDJPY=155.0
NEXT_PUBLIC_API_URL=http://your-api-domain.com:8080
```

2. `docker-compose.yml`で環境変数を読み込む：

```yaml
services:
  api:
    environment:
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-password}
      FX_USDJPY: ${FX_USDJPY:-150.0}
```

### 方法3: 環境変数を直接エクスポート

```bash
export POSTGRES_PASSWORD=your_secure_password
export FX_USDJPY=155.0
docker-compose up
```

---

## 本番環境で特に注意すべき設定

1. **POSTGRES_PASSWORD**: 必ず強力なパスワードに変更
2. **POSTGRES_SSLMODE**: `require` または `verify-full` に設定
3. **REDIS_PASSWORD**: セキュリティのため設定を推奨
4. **NEXT_PUBLIC_API_URL**: 実際のAPIのURLに変更
5. **USER_AGENT**: 実際のサイトにアクセスする場合は適切なUser-Agentを設定

---

## 現在のデフォルト値（変更不要な場合）

以下の環境変数はデフォルト値が設定されているため、特に変更する必要がなければ設定不要です：

- `API_PORT=8080`
- `API_HOST=0.0.0.0`
- `POSTGRES_HOST=postgres`（Docker内では`postgres`、ローカルでは`localhost`）
- `POSTGRES_PORT=5432`
- `POSTGRES_USER=pricecompare`
- `POSTGRES_DB=pricecompare`
- `POSTGRES_SSLMODE=disable`
- `REDIS_HOST=redis`（Docker内では`redis`、ローカルでは`localhost`）
- `REDIS_PORT=6379`
- `REDIS_DB=0`
- `US_SHIP_MODE=TABLE`
- `SHIPPING_FEE_PERCENT=3.0`
- `FX_USDJPY=150.0`
- `USER_AGENT=PriceCompareBot/1.0`
- `RATE_LIMIT_REQUESTS_PER_SECOND=10`
- `RATE_LIMIT_BURST=20`
- `NEXT_PUBLIC_API_URL=http://localhost:8080`

