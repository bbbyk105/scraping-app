'use client'

import { useState, useEffect } from 'react'
import { useSearchParams } from 'next/navigation'
import { getProduct, getProductOffersWithSort, Offer, Product } from '@/lib/api'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { ArrowLeft, ExternalLink, Package, Truck, CheckCircle2, XCircle, ArrowUpDown } from 'lucide-react'

function formatPrice(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`
}

function ShippingBreakdown({ offer }: { offer: Offer }) {
  const baseShipping = offer.shipping_to_us_amount
  const fee = offer.fee_amount || (offer.total_to_us_amount - offer.price_amount - baseShipping)
  const price = offer.price_amount

  return (
    <div className="text-sm space-y-2 p-3 bg-white rounded-lg border shadow-lg">
      <div className="font-bold text-base border-b pb-2">内訳</div>
      <div className="space-y-2">
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
        {offer.tax_amount && offer.tax_amount > 0 && (
          <div className="flex justify-between">
            <span>税:</span>
            <span className="font-semibold">{formatPrice(offer.tax_amount)}</span>
          </div>
        )}
        <div className="border-t pt-2 flex justify-between font-bold text-green-700">
          <span>合計:</span>
          <span>{formatPrice(offer.total_to_us_amount)}</span>
        </div>
      </div>
    </div>
  )
}

type SortKey = 'total' | 'fastest' | 'newest' | 'in_stock'

export default function ComparePage() {
  const searchParams = useSearchParams()
  const productId = searchParams.get('productId')

  const [product, setProduct] = useState<Product | null>(null)
  const [offers, setOffers] = useState<Offer[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [sortKey, setSortKey] = useState<SortKey>('total')

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
          getProductOffersWithSort(productId, sortKey),
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
  }, [productId, sortKey])

  const handleSortChange = (newSort: SortKey) => {
    setSortKey(newSort)
    setLoading(true)
    getProductOffersWithSort(productId!, newSort)
      .then(setOffers)
      .catch((err) => {
        setError(err instanceof Error ? err.message : 'データの取得に失敗しました')
      })
      .finally(() => setLoading(false))
  }

  if (loading) {
    return (
      <main className="min-h-screen p-8 bg-gradient-to-b from-gray-50 to-white">
        <div className="max-w-6xl mx-auto">
          <Card>
            <CardContent className="pt-6 text-center">
              <p className="text-gray-700">読み込み中...</p>
            </CardContent>
          </Card>
        </div>
      </main>
    )
  }

  if (error || !product) {
    return (
      <main className="min-h-screen p-8 bg-gradient-to-b from-gray-50 to-white">
        <div className="max-w-6xl mx-auto">
          <div className="mb-8">
            <Link href="/search">
              <Button variant="ghost" className="gap-2">
                <ArrowLeft className="h-4 w-4" />
                検索に戻る
              </Button>
            </Link>
          </div>
          <Card className="border-red-200 bg-red-50">
            <CardContent className="pt-6">
              <p className="font-medium text-red-800">{error || '商品が見つかりませんでした'}</p>
            </CardContent>
          </Card>
        </div>
      </main>
    )
  }

  return (
    <main className="min-h-screen p-4 md:p-8 bg-gradient-to-b from-gray-50 to-white">
      <div className="max-w-6xl mx-auto">
        <div className="mb-8">
          <Link href="/search">
            <Button variant="ghost" className="gap-2">
              <ArrowLeft className="h-4 w-4" />
              検索に戻る
            </Button>
          </Link>
        </div>

        <Card className="mb-8">
          <CardHeader>
            <div className="flex gap-6">
              {product.image_url && (
                <div className="relative w-32 h-32 bg-gray-100 rounded-lg overflow-hidden flex-shrink-0">
                  <img
                    src={product.image_url}
                    alt={product.title}
                    className="w-full h-full object-cover"
                  />
                </div>
              )}
              <div className="flex-1">
                <CardTitle className="text-3xl mb-4">{product.title}</CardTitle>
                <div className="flex flex-wrap gap-3">
                  {product.brand && (
                    <Badge variant="secondary" className="text-base px-3 py-1">
                      {product.brand}
                    </Badge>
                  )}
                  {product.model && (
                    <Badge variant="outline" className="text-base px-3 py-1">
                      {product.model}
                    </Badge>
                  )}
                </div>
              </div>
            </div>
          </CardHeader>
        </Card>

        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-2xl font-semibold">
            価格比較 ({offers.length}件)
          </h2>
          <div className="flex items-center gap-2">
            <ArrowUpDown className="h-4 w-4 text-gray-500" />
            <Select value={sortKey} onValueChange={handleSortChange}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="並び替え" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="total">総額が安い順</SelectItem>
                <SelectItem value="fastest">納期が早い順</SelectItem>
                <SelectItem value="newest">更新日時が新しい順</SelectItem>
                <SelectItem value="in_stock">在庫あり優先</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        {offers.length === 0 ? (
          <Card>
            <CardContent className="pt-6 text-center">
              <Package className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-600 text-lg">
                この商品のオファーが見つかりませんでした
              </p>
              <p className="text-gray-500 text-sm mt-2">
                価格更新ジョブを実行してください
              </p>
            </CardContent>
          </Card>
        ) : (
          <Card>
            <CardContent className="p-0">
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>販売元</TableHead>
                      <TableHead>ソース</TableHead>
                      <TableHead className="text-right">商品価格</TableHead>
                      <TableHead className="text-right">送料</TableHead>
                      <TableHead className="text-right">手数料</TableHead>
                      <TableHead className="text-right">合計</TableHead>
                      <TableHead className="text-center">推定到着</TableHead>
                      <TableHead className="text-center">在庫</TableHead>
                      <TableHead>更新日時</TableHead>
                      <TableHead>詳細</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {offers.map((offer) => (
                      <TableRow key={offer.id} className="hover:bg-muted/50">
                        <TableCell className="font-medium">{offer.seller}</TableCell>
                        <TableCell>
                          <Badge variant="secondary">{offer.source}</Badge>
                        </TableCell>
                        <TableCell className="text-right font-semibold">
                          {formatPrice(offer.price_amount)}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="group relative inline-block">
                            <span className="font-semibold cursor-help">
                              {formatPrice(offer.shipping_to_us_amount)}
                            </span>
                            <div className="absolute left-0 top-full mt-2 hidden group-hover:block z-50">
                              <ShippingBreakdown offer={offer} />
                            </div>
                          </div>
                        </TableCell>
                        <TableCell className="text-right">
                          {offer.fee_amount ? (
                            <span className="font-semibold">{formatPrice(offer.fee_amount)}</span>
                          ) : (
                            <span className="text-gray-400">-</span>
                          )}
                        </TableCell>
                        <TableCell className="text-right">
                          <span className="text-xl font-bold text-green-600">
                            {formatPrice(offer.total_to_us_amount)}
                          </span>
                        </TableCell>
                        <TableCell className="text-center">
                          {offer.est_delivery_days_min !== null && offer.est_delivery_days_max !== null ? (
                            <div className="flex items-center justify-center gap-1">
                              <Truck className="h-4 w-4 text-gray-500" />
                              <span>
                                {offer.est_delivery_days_min}-{offer.est_delivery_days_max}日
                              </span>
                            </div>
                          ) : (
                            <span className="text-gray-500">N/A</span>
                          )}
                        </TableCell>
                        <TableCell className="text-center">
                          {offer.in_stock ? (
                            <Badge className="bg-green-100 text-green-800 hover:bg-green-100">
                              <CheckCircle2 className="h-3 w-3 mr-1" />
                              在庫あり
                            </Badge>
                          ) : (
                            <Badge variant="destructive">
                              <XCircle className="h-3 w-3 mr-1" />
                              在庫なし
                            </Badge>
                          )}
                        </TableCell>
                        <TableCell className="text-sm text-gray-600">
                          {offer.price_updated_at
                            ? new Date(offer.price_updated_at).toLocaleString('ja-JP', {
                                year: 'numeric',
                                month: '2-digit',
                                day: '2-digit',
                                hour: '2-digit',
                                minute: '2-digit',
                              })
                            : '-'}
                        </TableCell>
                        <TableCell>
                          {offer.url ? (
                            <div>
                              <a
                                href={offer.url}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="inline-flex items-center gap-1 text-blue-600 hover:underline font-medium"
                              >
                                詳細を見る
                                <ExternalLink className="h-3 w-3" />
                              </a>
                            </div>
                          ) : (
                            <span className="text-gray-400">-</span>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        )}

        <Card className="mt-8 border-blue-200 bg-blue-50">
          <CardContent className="pt-6">
            <p className="text-sm text-gray-700">
              <strong>免責事項:</strong> 表示されている価格は参考情報です。最新の価格や在庫状況は販売元のサイトでご確認ください。
            </p>
          </CardContent>
        </Card>
      </div>
    </main>
  )
}
