'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { searchProducts, ProductWithMinPrice, resolveURL } from '@/lib/api'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { ArrowLeft, Search as SearchIcon, Package, Link as LinkIcon } from 'lucide-react'

function formatPrice(cents: number | null | undefined): string {
  if (cents === null || cents === undefined) return 'N/A'
  return `$${(cents / 100).toFixed(2)}`
}

type TabType = 'keyword' | 'url'

export default function SearchPage() {
  const router = useRouter()
  const [activeTab, setActiveTab] = useState<TabType>('keyword')
  const [query, setQuery] = useState('')
  const [url, setUrl] = useState('')
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

  const handleResolveURL = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!url.trim()) return

    setLoading(true)
    setError(null)
    try {
      const result = await resolveURL(url)
      // URL解決後、compareページにリダイレクト
      router.push(`/compare?productId=${result.product.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'URLの解決に失敗しました')
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
          <p className="text-gray-600">キーワードまたは商品URLで検索してください</p>
        </div>

        <Card className="mb-8">
          <CardHeader>
            <div className="flex gap-2 border-b">
              <button
                type="button"
                onClick={() => {
                  setActiveTab('keyword')
                  setError(null)
                }}
                className={`px-4 py-2 font-medium transition-colors border-b-2 ${
                  activeTab === 'keyword'
                    ? 'border-blue-600 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700'
                }`}
              >
                <SearchIcon className="inline h-4 w-4 mr-2" />
                キーワード検索
              </button>
              <button
                type="button"
                onClick={() => {
                  setActiveTab('url')
                  setError(null)
                }}
                className={`px-4 py-2 font-medium transition-colors border-b-2 ${
                  activeTab === 'url'
                    ? 'border-blue-600 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700'
                }`}
              >
                <LinkIcon className="inline h-4 w-4 mr-2" />
                URL入力
              </button>
            </div>
          </CardHeader>
          <CardContent className="pt-6">
            {activeTab === 'keyword' ? (
              <form onSubmit={handleSearch} className="flex gap-4">
                <div className="flex-1 relative">
                  <SearchIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
                  <Input
                    type="text"
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    placeholder="商品名、型番、JANコードを入力..."
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
            ) : (
              <form onSubmit={handleResolveURL} className="flex gap-4">
                <div className="flex-1 relative">
                  <LinkIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
                  <Input
                    type="url"
                    value={url}
                    onChange={(e) => setUrl(e.target.value)}
                    placeholder="商品URLを貼り付けてください (例: https://www.amazon.com/dp/B08N5WRWNW)"
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
                    <>解析中...</>
                  ) : (
                    <>
                      <LinkIcon className="mr-2 h-4 w-4" />
                      解析
                    </>
                  )}
                </Button>
              </form>
            )}
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
