import Link from 'next/link'

export default function Home() {
  return (
    <main className="min-h-screen p-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-4xl font-bold mb-8">Price Compare</h1>
        <div className="space-y-4">
          <Link
            href="/search"
            className="block p-4 border rounded-lg hover:bg-gray-50 dark:hover:bg-gray-900"
          >
            <h2 className="text-2xl font-semibold mb-2">商品検索</h2>
            <p className="text-gray-600 dark:text-gray-400">
              キーワードで商品を検索
            </p>
          </Link>
          <Link
            href="/admin/jobs"
            className="block p-4 border rounded-lg hover:bg-gray-50 dark:hover:bg-gray-900"
          >
            <h2 className="text-2xl font-semibold mb-2">管理画面</h2>
            <p className="text-gray-600 dark:text-gray-400">
              価格更新ジョブの実行
            </p>
          </Link>
        </div>
      </div>
    </main>
  )
}

