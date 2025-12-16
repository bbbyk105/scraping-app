import Link from 'next/link'

export default function Home() {
  return (
    <main className="min-h-screen p-8 bg-white">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-4xl font-bold mb-8 text-gray-900">Price Compare</h1>
        <div className="space-y-4">
          <Link
            href="/search"
            className="block p-6 border-2 border-gray-200 rounded-lg hover:bg-blue-50 hover:border-blue-300 transition-colors shadow-sm"
          >
            <h2 className="text-2xl font-semibold mb-2 text-gray-900">商品検索</h2>
            <p className="text-gray-700 text-base">
              キーワードで商品を検索
            </p>
          </Link>
          <Link
            href="/admin/jobs"
            className="block p-6 border-2 border-gray-200 rounded-lg hover:bg-blue-50 hover:border-blue-300 transition-colors shadow-sm"
          >
            <h2 className="text-2xl font-semibold mb-2 text-gray-900">管理画面</h2>
            <p className="text-gray-700 text-base">
              価格更新ジョブの実行
            </p>
          </Link>
        </div>
      </div>
    </main>
  )
}

