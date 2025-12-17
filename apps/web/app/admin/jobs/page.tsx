'use client'

import { useState } from 'react'
import { fetchPrices } from '@/lib/api'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { ArrowLeft, Settings, Play, CheckCircle2, XCircle, Info } from 'lucide-react'

const ENABLE_DEMO_PROVIDERS =
  typeof process !== 'undefined' &&
  process.env.NEXT_PUBLIC_ENABLE_DEMO_PROVIDERS === 'true'

export default function AdminJobsPage() {
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  const handleFetchPrices = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setMessage(null)

    try {
      // source を指定しない → API 側で自動的に "all" と解釈され、有効なプロバイダから取得
      const result = await fetchPrices()
      setMessage({
        type: 'success',
        text: `ジョブが正常にキューに追加されました。Job ID: ${result.job_id}`,
      })
    } catch (err) {
      setMessage({
        type: 'error',
        text: err instanceof Error ? err.message : 'ジョブの実行に失敗しました',
      })
    } finally {
      setLoading(false)
    }
  }

  return (
    <main className="min-h-screen p-4 md:p-8 bg-gradient-to-b from-gray-50 to-white">
      <div className="max-w-4xl mx-auto">
        <div className="mb-8">
          <Link href="/">
            <Button variant="ghost" className="gap-2">
              <ArrowLeft className="h-4 w-4" />
              ホームに戻る
            </Button>
          </Link>
        </div>

        <div className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <div className="p-2 bg-purple-100 rounded-lg">
              <Settings className="h-6 w-6 text-purple-600" />
            </div>
            <h1 className="text-4xl font-bold bg-gradient-to-r from-purple-600 to-blue-600 bg-clip-text text-transparent">
              価格更新ジョブ管理
            </h1>
          </div>
          <p className="text-gray-600">データソースから価格情報を取得・更新します</p>
        </div>

        <Card className="mb-8">
          <CardHeader>
            <CardTitle>ジョブ実行</CardTitle>
            <CardDescription>
              有効になっているプロバイダから自動的に価格情報を取得・更新します
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleFetchPrices} className="space-y-6">
              <Button
                type="submit"
                disabled={loading}
                size="lg"
                className="w-full"
              >
                {loading ? (
                  <>実行中...</>
                ) : (
                  <>
                    <Play className="mr-2 h-4 w-4" />
                    価格更新ジョブを実行（自動判定）
                  </>
                )}
              </Button>
            </form>
          </CardContent>
        </Card>

        {message && (
          <Card className={`mb-8 ${
            message.type === 'success'
              ? 'border-green-200 bg-green-50'
              : 'border-red-200 bg-red-50'
          }`}>
            <CardContent className="pt-6">
              <div className="flex items-start gap-3">
                {message.type === 'success' ? (
                  <CheckCircle2 className="h-5 w-5 text-green-600 mt-0.5 flex-shrink-0" />
                ) : (
                  <XCircle className="h-5 w-5 text-red-600 mt-0.5 flex-shrink-0" />
                )}
                <p className={`font-medium ${
                  message.type === 'success' ? 'text-green-800' : 'text-red-800'
                }`}>
                  {message.text}
                </p>
              </div>
            </CardContent>
          </Card>
        )}

        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Info className="h-5 w-5 text-blue-600" />
              <CardTitle>プロバイダについて</CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {ENABLE_DEMO_PROVIDERS && (
                <>
                  <div>
                    <h3 className="font-semibold mb-2">Demo プロバイダ（開発・テスト用）</h3>
                    <p className="text-sm text-gray-600">
                      モックデータを使用してテストします。本番環境では通常無効化されます。
                    </p>
                  </div>
                  <div>
                    <h3 className="font-semibold mb-2">Public HTML プロバイダ（開発・テスト用）</h3>
                    <p className="text-sm text-gray-600">
                      /samples 配下のHTMLファイルから価格情報を抽出します。本番ビルドでは使用しない想定の検証用プロバイダです。
                    </p>
                  </div>
                </>
              )}
              <div>
                <h3 className="font-semibold mb-2">Live プロバイダ</h3>
                <p className="text-sm text-gray-600">
                  実際の外部サイトから商品情報を取得します。ALLOW_LIVE_FETCH=true に設定する必要があります。
                  robots.txtチェック、レートリミット、監査ログが自動的に適用されます。
                </p>
              </div>
              <div>
                <h3 className="font-semibold mb-2">すべて</h3>
                <p className="text-sm text-gray-600">
                  有効化されているすべてのプロバイダを順次実行します。
                </p>
              </div>
              <div>
                <h3 className="font-semibold mb-2">結果の確認方法</h3>
                <p className="text-sm text-gray-600">
                  ジョブ実行後に
                  <code className="mx-1">/search</code>
                  から商品を選択し、
                  <code className="mx-1">/compare</code>
                  画面の「更新日時」列や「更新日時が新しい順」ソートで価格更新が反映されていることを確認できます。
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </main>
  )
}
