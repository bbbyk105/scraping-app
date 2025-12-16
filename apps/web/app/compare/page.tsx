'use client'

import { useState, useEffect } from 'react'
import { useSearchParams } from 'next/navigation'
import { getProduct, getProductOffers, Offer, Product } from '@/lib/api'
import Link from 'next/link'

function formatPrice(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`
}

function ShippingBreakdown({ offer }: { offer: Offer }) {
  const baseShipping = offer.shipping_to_us_amount
  const fee = offer.total_to_us_amount - offer.price_amount - baseShipping
  const price = offer.price_amount

  return (
    <div className="text-xs text-gray-600 dark:text-gray-400 space-y-1">
      <div>商品価格: {formatPrice(price)}</div>
      <div>送料: {formatPrice(baseShipping)}</div>
      {fee > 0 && <div>手数料: {formatPrice(fee)}</div>}
      <div className="font-semibold">合計: {formatPrice(offer.total_to_us_amount)}</div>
    </div>
  )
}

export default function ComparePage() {
  const searchParams = useSearchParams()
  const productId = searchParams.get('productId')

  const [product, setProduct] = useState<Product | null>(null)
  const [offers, setOffers] = useState<Offer[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!productId) {
      setError('商品IDが指定されていません')
      setLoading(false)
      return
    }

    const fetchData = async () => {
      try {
        const [productData, offersData] = await Promise.all([
          getProduct(productId),
          getProductOffers(productId),
        ])
        setProduct(productData)
        setOffers(offersData)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'データの取得に失敗しました')
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [productId])

  if (loading) {
    return (
      <main className="min-h-screen p-8">
        <div className="max-w-6xl mx-auto">
          <p>読み込み中...</p>
        </div>
      </main>
    )
  }

  if (error || !product) {
    return (
      <main className="min-h-screen p-8">
        <div className="max-w-6xl mx-auto">
          <div className="mb-8">
            <Link href="/search" className="text-blue-600 hover:underline">
              ← 検索に戻る
            </Link>
          </div>
          <div className="p-4 bg-red-100 border border-red-400 text-red-700 rounded">
            {error || '商品が見つかりませんでした'}
          </div>
        </div>
      </main>
    )
  }

  return (
    <main className="min-h-screen p-8">
      <div className="max-w-6xl mx-auto">
        <div className="mb-8">
          <Link href="/search" className="text-blue-600 hover:underline">
            ← 検索に戻る
          </Link>
        </div>

        <div className="mb-8">
          {product.image_url && (
            <img
              src={product.image_url}
              alt={product.title}
              className="w-32 h-32 object-cover rounded mb-4"
            />
          )}
          <h1 className="text-4xl font-bold mb-4">{product.title}</h1>
          {product.brand && (
            <p className="text-lg text-gray-600 dark:text-gray-400 mb-2">
              ブランド: {product.brand}
            </p>
          )}
          {product.model && (
            <p className="text-lg text-gray-600 dark:text-gray-400 mb-4">
              モデル: {product.model}
            </p>
          )}
        </div>

        <h2 className="text-2xl font-semibold mb-4">価格比較 ({offers.length}件)</h2>

        {offers.length === 0 ? (
          <p className="text-gray-600 dark:text-gray-400">
            この商品のオファーが見つかりませんでした。価格更新ジョブを実行してください。
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full border-collapse border">
              <thead>
                <tr className="bg-gray-100 dark:bg-gray-800">
                  <th className="border p-4 text-left">販売元</th>
                  <th className="border p-4 text-left">ソース</th>
                  <th className="border p-4 text-right">商品価格</th>
                  <th className="border p-4 text-right">US送料</th>
                  <th className="border p-4 text-right">合計</th>
                  <th className="border p-4 text-center">推定到着</th>
                  <th className="border p-4 text-center">在庫</th>
                  <th className="border p-4 text-left">詳細</th>
                </tr>
              </thead>
              <tbody>
                {offers.map((offer) => (
                  <tr key={offer.id} className="hover:bg-gray-50 dark:hover:bg-gray-900">
                    <td className="border p-4">{offer.seller}</td>
                    <td className="border p-4">
                      <span className="px-2 py-1 bg-blue-100 dark:bg-blue-900 rounded text-sm">
                        {offer.source}
                      </span>
                    </td>
                    <td className="border p-4 text-right">{formatPrice(offer.price_amount)}</td>
                    <td className="border p-4 text-right">
                      <div className="group relative">
                        {formatPrice(offer.shipping_to_us_amount)}
                        <div className="absolute left-0 top-full mt-1 hidden group-hover:block z-10 bg-white dark:bg-gray-800 border rounded p-2 shadow-lg">
                          <ShippingBreakdown offer={offer} />
                        </div>
                      </div>
                    </td>
                    <td className="border p-4 text-right font-bold text-green-600">
                      {formatPrice(offer.total_to_us_amount)}
                    </td>
                    <td className="border p-4 text-center">
                      {offer.est_delivery_days_min !== null && offer.est_delivery_days_max !== null
                        ? `${offer.est_delivery_days_min}-${offer.est_delivery_days_max}日`
                        : 'N/A'}
                    </td>
                    <td className="border p-4 text-center">
                      {offer.in_stock ? (
                        <span className="text-green-600">在庫あり</span>
                      ) : (
                        <span className="text-red-600">在庫なし</span>
                      )}
                    </td>
                    <td className="border p-4">
                      {offer.url ? (
                        <a
                          href={offer.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-blue-600 hover:underline"
                        >
                          詳細を見る
                        </a>
                      ) : (
                        <span className="text-gray-400">-</span>
                      )}
                      <div className="text-xs text-gray-500 mt-1">
                        更新: {new Date(offer.fetched_at).toLocaleString('ja-JP')}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        <footer className="mt-12 pt-8 border-t">
          <p className="text-sm text-gray-600 dark:text-gray-400">
            免責事項: 表示されている価格は参考情報です。最新の価格や在庫状況は販売元のサイトでご確認ください。
          </p>
        </footer>
      </div>
    </main>
  )
}

