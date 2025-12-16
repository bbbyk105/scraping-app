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
    <main className="min-h-screen p-8 bg-white">
      <div className="max-w-4xl mx-auto">
        <div className="mb-8">
          <Link href="/" className="text-blue-600 hover:underline text-base font-medium">
            ← ホームに戻る
          </Link>
        </div>

        <h1 className="text-4xl font-bold mb-8 text-gray-900">価格更新ジョブ管理</h1>

        <form onSubmit={handleFetchPrices} className="mb-8">
          <div className="space-y-4">
            <div>
              <label htmlFor="source" className="block text-base font-semibold mb-2 text-gray-900">
                ソースを選択
              </label>
              <select
                id="source"
                value={source}
                onChange={(e) => setSource(e.target.value)}
                className="w-full px-4 py-3 border-2 border-gray-300 rounded-lg text-base focus:outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-200"
              >
                <option value="all">すべて</option>
                <option value="demo">Demo プロバイダ</option>
                <option value="public_html">Public HTML プロバイダ</option>
              </select>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 font-medium text-base transition-colors"
            >
              {loading ? '実行中...' : '価格更新ジョブを実行'}
            </button>
          </div>
        </form>

        {message && (
          <div
            className={`p-4 rounded-lg ${
              message.type === 'success'
                ? 'bg-green-50 border-2 border-green-300 text-green-800'
                : 'bg-red-50 border-2 border-red-300 text-red-800'
            }`}
          >
            <p className="font-medium">{message.text}</p>
          </div>
        )}

        <div className="mt-8 p-6 bg-gray-50 border-2 border-gray-200 rounded-lg">
          <h2 className="text-xl font-semibold mb-4 text-gray-900">説明</h2>
          <ul className="space-y-2 text-base text-gray-700">
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

