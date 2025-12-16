'use client'

import { useState } from 'react'
import { searchProducts, ProductWithMinPrice } from '@/lib/api'
import Link from 'next/link'

function formatPrice(cents: number | null | undefined): string {
  if (cents === null || cents === undefined) return 'N/A'
  return `$${(cents / 100).toFixed(2)}`
}

export default function SearchPage() {
  const [query, setQuery] = useState('')
  const [products, setProducts] = useState<ProductWithMinPrice[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!query.trim()) return

    setLoading(true)
    setError(null)
    try {
      const results = await searchProducts(query)
      setProducts(results)
    } catch (err) {
      console.error('Search error:', err)
      if (err instanceof TypeError && err.message.includes('fetch')) {
        setError('APIサーバーに接続できません。APIサーバーが起動しているか確認してください。')
      } else {
        setError(err instanceof Error ? err.message : '検索に失敗しました')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <main className="min-h-screen p-8 bg-white">
      <div className="max-w-6xl mx-auto">
        <div className="mb-8">
          <Link href="/" className="text-blue-600 hover:underline text-base font-medium">
            ← ホームに戻る
          </Link>
        </div>

        <h1 className="text-4xl font-bold mb-8 text-gray-900">商品検索</h1>

        <form onSubmit={handleSearch} className="mb-8">
          <div className="flex gap-4">
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="商品名を入力..."
              className="flex-1 px-4 py-3 border-2 border-gray-300 rounded-lg text-base focus:outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-200"
            />
            <button
              type="submit"
              disabled={loading}
              className="px-8 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 font-medium text-base transition-colors"
            >
              {loading ? '検索中...' : '検索'}
            </button>
          </div>
        </form>

        {error && (
          <div className="mb-4 p-4 bg-red-50 border-2 border-red-300 text-red-800 rounded-lg">
            <p className="font-medium">{error}</p>
          </div>
        )}

        {products.length > 0 && (
          <div className="space-y-4">
            <h2 className="text-2xl font-semibold text-gray-900">検索結果 ({products.length}件)</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {products.map((product) => (
                <Link
                  key={product.id}
                  href={`/compare?productId=${product.id}`}
                  className="block p-5 border-2 border-gray-200 rounded-lg hover:bg-blue-50 hover:border-blue-300 transition-colors shadow-sm"
                >
                  {product.image_url && (
                    <img
                      src={product.image_url}
                      alt={product.title}
                      className="w-full h-48 object-cover rounded mb-4"
                    />
                  )}
                  <h3 className="font-semibold mb-2 text-gray-900 text-lg">{product.title}</h3>
                  {product.brand && (
                    <p className="text-sm text-gray-700 mb-2">
                      ブランド: {product.brand}
                    </p>
                  )}
                  <p className="text-lg font-bold text-blue-600">
                    最安値: {formatPrice(product.min_price_cents)}
                  </p>
                </Link>
              ))}
            </div>
          </div>
        )}

        {!loading && products.length === 0 && query && (
          <p className="text-gray-700 text-base">検索結果が見つかりませんでした</p>
        )}
      </div>
    </main>
  )
}

