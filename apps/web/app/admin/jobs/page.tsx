'use client'

import { useState } from 'react'
import { fetchPrices } from '@/lib/api'
import Link from 'next/link'

export default function AdminJobsPage() {
  const [source, setSource] = useState('all')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  const handleFetchPrices = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setMessage(null)

    try {
      const result = await fetchPrices(source)
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
    <main className="min-h-screen p-8">
      <div className="max-w-4xl mx-auto">
        <div className="mb-8">
          <Link href="/" className="text-blue-600 hover:underline">
            ← ホームに戻る
          </Link>
        </div>

        <h1 className="text-4xl font-bold mb-8">価格更新ジョブ管理</h1>

        <form onSubmit={handleFetchPrices} className="mb-8">
          <div className="space-y-4">
            <div>
              <label htmlFor="source" className="block text-sm font-medium mb-2">
                ソースを選択
              </label>
              <select
                id="source"
                value={source}
                onChange={(e) => setSource(e.target.value)}
                className="w-full px-4 py-2 border rounded-lg"
              >
                <option value="all">すべて</option>
                <option value="demo">Demo プロバイダ</option>
                <option value="public_html">Public HTML プロバイダ</option>
              </select>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
            >
              {loading ? '実行中...' : '価格更新ジョブを実行'}
            </button>
          </div>
        </form>

        {message && (
          <div
            className={`p-4 rounded-lg ${
              message.type === 'success'
                ? 'bg-green-100 border border-green-400 text-green-700'
                : 'bg-red-100 border border-red-400 text-red-700'
            }`}
          >
            {message.text}
          </div>
        )}

        <div className="mt-8 p-4 bg-gray-100 dark:bg-gray-800 rounded-lg">
          <h2 className="text-xl font-semibold mb-4">説明</h2>
          <ul className="space-y-2 text-sm text-gray-600 dark:text-gray-400">
            <li>
              <strong>Demo プロバイダ:</strong> モックデータを使用してテストします
            </li>
            <li>
              <strong>Public HTML プロバイダ:</strong> /samples 配下のHTMLファイルから価格情報を抽出します
            </li>
            <li>
              <strong>すべて:</strong> すべてのプロバイダを実行します
            </li>
          </ul>
        </div>
      </div>
    </main>
  )
}

