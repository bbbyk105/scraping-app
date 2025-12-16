import Link from 'next/link'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Search, Settings } from 'lucide-react'

export default function Home() {
  return (
    <main className="min-h-screen p-8 bg-gradient-to-b from-gray-50 to-white">
      <div className="max-w-4xl mx-auto">
        <div className="mb-12 text-center">
          <h1 className="text-5xl font-bold mb-4 bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
            Price Compare
          </h1>
          <p className="text-xl text-gray-600">
            複数のソースから商品価格を比較
          </p>
        </div>
        <div className="grid gap-6 md:grid-cols-2">
          <Link href="/search">
            <Card className="h-full hover:shadow-lg transition-shadow cursor-pointer border-2 hover:border-blue-300">
              <CardHeader>
                <div className="flex items-center gap-3 mb-2">
                  <div className="p-2 bg-blue-100 rounded-lg">
                    <Search className="h-6 w-6 text-blue-600" />
                  </div>
                  <CardTitle className="text-2xl">商品検索</CardTitle>
                </div>
                <CardDescription className="text-base">
                  キーワードで商品を検索して価格を比較
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-gray-600">
                  複数のソースから最適な価格を見つけましょう
                </p>
              </CardContent>
            </Card>
          </Link>
          <Link href="/admin/jobs">
            <Card className="h-full hover:shadow-lg transition-shadow cursor-pointer border-2 hover:border-purple-300">
              <CardHeader>
                <div className="flex items-center gap-3 mb-2">
                  <div className="p-2 bg-purple-100 rounded-lg">
                    <Settings className="h-6 w-6 text-purple-600" />
                  </div>
                  <CardTitle className="text-2xl">管理画面</CardTitle>
                </div>
                <CardDescription className="text-base">
                  価格更新ジョブの実行と管理
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-gray-600">
                  データソースの更新を手動で実行できます
                </p>
              </CardContent>
            </Card>
          </Link>
        </div>
      </div>
    </main>
  )
}
