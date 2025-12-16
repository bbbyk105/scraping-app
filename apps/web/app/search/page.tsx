'use client'

import { useState } from 'react'
import { searchProducts, ProductWithMinPrice } from '@/lib/api'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { ArrowLeft, Search as SearchIcon, Package } from 'lucide-react'

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
    <main className="min-h-screen p-4 md:p-8 bg-gradient-to-b from-gray-50 to-white">
      <div className="max-w-6xl mx-auto">
        <div className="mb-8">
          <Link href="/">
            <Button variant="ghost" className="gap-2">
              <ArrowLeft className="h-4 w-4" />
              ホームに戻る
            </Button>
          </Link>
        </div>

        <div className="mb-8">
          <h1 className="text-4xl font-bold mb-2 bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
            商品検索
          </h1>
          <p className="text-gray-600">キーワードで商品を検索してください</p>
        </div>

        <Card className="mb-8">
          <CardContent className="pt-6">
            <form onSubmit={handleSearch} className="flex gap-4">
              <div className="flex-1 relative">
                <SearchIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
                <Input
                  type="text"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="商品名を入力..."
                  className="pl-10 h-12 text-base"
                />
              </div>
              <Button
                type="submit"
                disabled={loading}
                size="lg"
                className="px-8"
              >
                {loading ? (
                  <>検索中...</>
                ) : (
                  <>
                    <SearchIcon className="mr-2 h-4 w-4" />
                    検索
                  </>
                )}
              </Button>
            </form>
          </CardContent>
        </Card>

        {error && (
          <Card className="mb-4 border-red-200 bg-red-50">
            <CardContent className="pt-6">
              <p className="text-red-800 font-medium">{error}</p>
            </CardContent>
          </Card>
        )}

        {products.length > 0 && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-2xl font-semibold">
                検索結果 ({products.length}件)
              </h2>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {products.map((product) => (
                <Link key={product.id} href={`/compare?productId=${product.id}`}>
                  <Card className="h-full hover:shadow-lg transition-all cursor-pointer border-2 hover:border-blue-300">
                    {product.image_url && (
                      <div className="relative w-full h-48 bg-gray-100 rounded-t-lg overflow-hidden">
                        <img
                          src={product.image_url}
                          alt={product.title}
                          className="w-full h-full object-cover"
                        />
                      </div>
                    )}
                    <CardHeader>
                      <CardTitle className="text-lg leading-tight line-clamp-2">{product.title}</CardTitle>
                      {product.brand && (
                        <CardDescription>
                          <Badge variant="secondary">{product.brand}</Badge>
                        </CardDescription>
                      )}
                    </CardHeader>
                    <CardContent>
                      <div className="flex items-baseline gap-2">
                        <span className="text-2xl font-bold text-green-600">
                          {formatPrice(product.min_price_cents)}
                        </span>
                        <span className="text-sm text-gray-500">から</span>
                      </div>
                    </CardContent>
                  </Card>
                </Link>
              ))}
            </div>
          </div>
        )}

        {!loading && products.length === 0 && query && (
          <Card>
            <CardContent className="pt-6 text-center">
              <Package className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-600 text-lg">検索結果が見つかりませんでした</p>
              <p className="text-gray-500 text-sm mt-2">
                別のキーワードで検索してみてください
              </p>
            </CardContent>
          </Card>
        )}
      </div>
    </main>
  )
}
