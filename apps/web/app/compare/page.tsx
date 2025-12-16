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
    <div className="text-base text-gray-900 space-y-2 bg-white p-3">
      <div className="font-bold text-lg border-b-2 border-gray-400 pb-1">内訳</div>
      <div className="space-y-2 pt-2">
        <div className="flex justify-between">
          <span>商品価格:</span>
          <span className="font-semibold">{formatPrice(price)}</span>
        </div>
        <div className="flex justify-between">
          <span>送料:</span>
          <span className="font-semibold">{formatPrice(baseShipping)}</span>
        </div>
        {fee > 0 && (
          <div className="flex justify-between">
            <span>手数料:</span>
            <span className="font-semibold">{formatPrice(fee)}</span>
          </div>
        )}
        <div className="border-t-2 border-gray-400 pt-2 flex justify-between font-bold text-green-700 text-lg">
          <span>合計:</span>
          <span>{formatPrice(offer.total_to_us_amount)}</span>
        </div>
      </div>
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
      <main className="min-h-screen p-8 bg-white">
        <div className="max-w-6xl mx-auto">
          <p className="text-gray-700 text-base">読み込み中...</p>
        </div>
      </main>
    )
  }

  if (error || !product) {
    return (
      <main className="min-h-screen p-8 bg-white">
        <div className="max-w-6xl mx-auto">
          <div className="mb-8">
            <Link href="/search" className="text-blue-600 hover:underline text-base font-medium">
              ← 検索に戻る
            </Link>
          </div>
          <div className="p-4 bg-red-50 border-2 border-red-300 text-red-800 rounded-lg">
            <p className="font-medium">{error || '商品が見つかりませんでした'}</p>
          </div>
        </div>
      </main>
    )
  }

  return (
    <main className="min-h-screen p-8 bg-white">
      <div className="max-w-6xl mx-auto">
        <div className="mb-8">
          <Link href="/search" className="text-blue-600 hover:underline text-base font-medium">
            ← 検索に戻る
          </Link>
        </div>

        <div className="mb-8">
          {product.image_url && (
            <img
              src={product.image_url}
              alt={product.title}
              className="w-32 h-32 object-cover rounded mb-4 border-2 border-gray-200"
            />
          )}
          <h1 className="text-4xl font-bold mb-4 text-gray-900">{product.title}</h1>
          {product.brand && (
            <p className="text-lg text-gray-700 mb-2">
              ブランド: {product.brand}
            </p>
          )}
          {product.model && (
            <p className="text-lg text-gray-700 mb-4">
              モデル: {product.model}
            </p>
          )}
        </div>

        <h2 className="text-2xl font-semibold mb-4 text-gray-900">価格比較 ({offers.length}件)</h2>

        {offers.length === 0 ? (
          <p className="text-gray-600 dark:text-gray-400">
            この商品のオファーが見つかりませんでした。価格更新ジョブを実行してください。
          </p>
        ) : (
          <div className="overflow-x-auto shadow-lg rounded-lg border-2 border-gray-300 bg-white">
            <table className="w-full border-collapse text-base">
              <thead>
                <tr className="bg-gray-200 border-b-2 border-gray-400">
                  <th className="border border-gray-400 p-4 text-left font-bold text-gray-900 text-base">販売元</th>
                  <th className="border border-gray-400 p-4 text-left font-bold text-gray-900 text-base">ソース</th>
                  <th className="border border-gray-400 p-4 text-right font-bold text-gray-900 text-base">商品価格</th>
                  <th className="border border-gray-400 p-4 text-right font-bold text-gray-900 text-base">US送料</th>
                  <th className="border border-gray-400 p-4 text-right font-bold text-gray-900 text-base">合計</th>
                  <th className="border border-gray-400 p-4 text-center font-bold text-gray-900 text-base">推定到着</th>
                  <th className="border border-gray-400 p-4 text-center font-bold text-gray-900 text-base">在庫</th>
                  <th className="border border-gray-400 p-4 text-left font-bold text-gray-900 text-base">詳細</th>
                </tr>
              </thead>
              <tbody>
                {offers.map((offer) => (
                  <tr key={offer.id} className="hover:bg-blue-50 transition-colors border-b border-gray-300">
                    <td className="border border-gray-400 p-4 text-gray-900 font-medium">{offer.seller}</td>
                    <td className="border border-gray-400 p-4">
                      <span className="px-3 py-1 bg-blue-200 text-blue-900 rounded font-semibold text-sm">
                        {offer.source}
                      </span>
                    </td>
                    <td className="border border-gray-400 p-4 text-right text-gray-900 font-semibold text-base">{formatPrice(offer.price_amount)}</td>
                    <td className="border border-gray-400 p-4 text-right text-gray-900 font-semibold text-base">
                      <div className="group relative">
                        {formatPrice(offer.shipping_to_us_amount)}
                        <div className="absolute left-0 top-full mt-1 hidden group-hover:block z-10 bg-white border-2 border-gray-400 rounded p-3 shadow-xl">
                          <ShippingBreakdown offer={offer} />
                        </div>
                      </div>
                    </td>
                    <td className="border border-gray-400 p-4 text-right font-bold text-green-700 text-lg bg-green-50">
                      {formatPrice(offer.total_to_us_amount)}
                    </td>
                    <td className="border border-gray-400 p-4 text-center text-gray-900 font-medium text-base">
                      {offer.est_delivery_days_min !== null && offer.est_delivery_days_max !== null
                        ? `${offer.est_delivery_days_min}-${offer.est_delivery_days_max}日`
                        : 'N/A'}
                    </td>
                    <td className="border border-gray-400 p-4 text-center">
                      {offer.in_stock ? (
                        <span className="text-green-700 font-bold bg-green-100 px-2 py-1 rounded">在庫あり</span>
                      ) : (
                        <span className="text-red-700 font-bold bg-red-100 px-2 py-1 rounded">在庫なし</span>
                      )}
                    </td>
                    <td className="border border-gray-400 p-4">
                      {offer.url ? (
                        <a
                          href={offer.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-blue-700 hover:underline font-semibold text-base"
                        >
                          詳細を見る
                        </a>
                      ) : (
                        <span className="text-gray-600">-</span>
                      )}
                      <div className="text-sm text-gray-700 mt-2 font-medium">
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

