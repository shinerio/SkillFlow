import { useState, useEffect } from 'react'
import { Star, Plus, Trash2, ExternalLink, Github, X } from 'lucide-react'
import { ListFavoriteRepos, AddFavoriteRepo, RemoveFavoriteRepo } from '../../wailsjs/go/main/App'
import { config } from '../../wailsjs/go/models'
import RepoInstallDialog from '../components/RepoInstallDialog'

export default function GitHubFavorites() {
  const [repos, setRepos] = useState<config.FavoriteRepo[]>([])
  const [adding, setAdding] = useState(false)
  const [newURL, setNewURL] = useState('')
  const [newDesc, setNewDesc] = useState('')
  const [addError, setAddError] = useState('')
  const [installRepo, setInstallRepo] = useState<config.FavoriteRepo | null>(null)

  useEffect(() => { load() }, [])

  async function load() {
    try {
      const list = await ListFavoriteRepos()
      setRepos(list || [])
    } catch {
      setRepos([])
    }
  }

  async function handleAdd() {
    if (!newURL.trim()) return
    try {
      await AddFavoriteRepo(newURL.trim(), newDesc.trim())
      setNewURL('')
      setNewDesc('')
      setAdding(false)
      setAddError('')
      load()
    } catch (e: any) {
      setAddError(e.message || String(e))
    }
  }

  async function handleRemove(url: string) {
    await RemoveFavoriteRepo(url)
    load()
  }

  return (
    <div className="p-6 h-full overflow-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-2">
          <Star size={20} className="text-yellow-400" />
          <h1 className="text-xl font-semibold">GitHub 收藏</h1>
          {repos.length > 0 && (
            <span className="text-sm text-gray-500">{repos.length} 个仓库</span>
          )}
        </div>
        <button
          onClick={() => { setAdding(true); setAddError('') }}
          className="flex items-center gap-1.5 px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm transition-colors"
        >
          <Plus size={15} />
          添加仓库
        </button>
      </div>

      {/* Add form */}
      {adding && (
        <div className="mb-6 p-4 bg-gray-800 rounded-xl border border-gray-700">
          <div className="flex items-center justify-between mb-3">
            <p className="text-sm font-medium">添加 GitHub 仓库</p>
            <button onClick={() => { setAdding(false); setAddError('') }} className="text-gray-500 hover:text-white transition-colors">
              <X size={16} />
            </button>
          </div>
          <input
            autoFocus
            value={newURL}
            onChange={e => setNewURL(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter') handleAdd(); if (e.key === 'Escape') setAdding(false) }}
            placeholder="https://github.com/owner/repo"
            className="w-full bg-gray-900 border border-gray-600 rounded-lg px-3 py-2 text-sm mb-2 focus:outline-none focus:border-indigo-500"
          />
          <input
            value={newDesc}
            onChange={e => setNewDesc(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter') handleAdd(); if (e.key === 'Escape') setAdding(false) }}
            placeholder="备注（可选）"
            className="w-full bg-gray-900 border border-gray-600 rounded-lg px-3 py-2 text-sm mb-3 focus:outline-none focus:border-indigo-500"
          />
          {addError && <p className="text-red-400 text-xs mb-2">{addError}</p>}
          <div className="flex gap-2">
            <button
              onClick={handleAdd}
              className="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm transition-colors"
            >
              添加
            </button>
            <button
              onClick={() => { setAdding(false); setAddError('') }}
              className="px-3 py-1.5 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm transition-colors"
            >
              取消
            </button>
          </div>
        </div>
      )}

      {/* Empty state */}
      {repos.length === 0 && !adding && (
        <div className="flex flex-col items-center justify-center h-64 text-gray-500 gap-3">
          <Star size={40} className="opacity-30" />
          <p className="text-sm">还没有收藏的仓库</p>
          <p className="text-xs text-gray-600">收藏 GitHub 仓库后，可一键扫描并安装技能</p>
        </div>
      )}

      {/* Repo grid */}
      {repos.length > 0 && (
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
          {repos.map(repo => (
            <div
              key={repo.url}
              className="bg-gray-800 border border-gray-700 rounded-xl p-4 flex flex-col gap-3 hover:border-gray-600 transition-colors"
            >
              {/* Card header */}
              <div className="flex items-start justify-between gap-2">
                <div className="flex items-center gap-2 min-w-0">
                  <Github size={15} className="text-gray-400 shrink-0" />
                  <span className="font-medium text-sm truncate">{repo.name}</span>
                </div>
                <button
                  onClick={() => handleRemove(repo.url)}
                  className="text-gray-500 hover:text-red-400 transition-colors shrink-0 p-0.5"
                  title="从收藏中移除"
                >
                  <Trash2 size={13} />
                </button>
              </div>

              {/* Description */}
              {repo.description && (
                <p className="text-xs text-gray-400 leading-relaxed">{repo.description}</p>
              )}

              {/* URL */}
              <div className="text-xs text-gray-500 truncate">{repo.url}</div>

              {/* Actions */}
              <div className="flex items-center gap-2 mt-auto pt-1">
                <button
                  onClick={() => setInstallRepo(repo)}
                  className="flex-1 px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-xs font-medium transition-colors text-center"
                >
                  扫描并安装
                </button>
                <a
                  href={repo.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="p-1.5 text-gray-400 hover:text-white transition-colors"
                  title="在浏览器中打开"
                >
                  <ExternalLink size={14} />
                </a>
              </div>

              {/* Added date */}
              {repo.addedAt && (
                <p className="text-xs text-gray-600">
                  添加于 {new Date(repo.addedAt).toLocaleDateString('zh-CN')}
                </p>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Install dialog */}
      {installRepo && (
        <RepoInstallDialog
          repoURL={installRepo.url}
          repoName={installRepo.name}
          onClose={() => setInstallRepo(null)}
        />
      )}
    </div>
  )
}
