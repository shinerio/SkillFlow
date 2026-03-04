import { useState } from 'react'
import { ScanGitHub, InstallFromGitHub, ListCategories } from '../../wailsjs/go/main/App'
import { Github, X } from 'lucide-react'

interface Props { onClose: () => void; onDone: () => void }

export default function GitHubInstallDialog({ onClose, onDone }: Props) {
  const [url, setUrl] = useState('')
  const [candidates, setCandidates] = useState<any[]>([])
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [categories, setCategories] = useState<string[]>([])
  const [category, setCategory] = useState('')
  const [scanning, setScanning] = useState(false)
  const [installing, setInstalling] = useState(false)

  const scan = async () => {
    setScanning(true)
    const [c, cats] = await Promise.all([ScanGitHub(url), ListCategories()])
    setCandidates(c ?? [])
    setCategories(cats ?? [])
    // 默认选择第一个分类（Imported）
    if ((cats ?? []).length > 0 && category === "") {
      setCategory(cats[0])
    }
    setSelected(new Set((c ?? []).filter((x: any) => !x.Installed).map((x: any) => x.Name)))
    setScanning(false)
  }

  const install = async () => {
    setInstalling(true)
    const toInstall = candidates.filter(c => selected.has(c.Name))
    await InstallFromGitHub(url, toInstall, category)
    setInstalling(false)
    onDone()
  }

  const toggle = (name: string) => {
    const next = new Set(selected)
    next.has(name) ? next.delete(name) : next.add(name)
    setSelected(next)
  }

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
      <div className="bg-gray-800 rounded-2xl p-6 w-[520px] border border-gray-700">
        <div className="flex justify-between items-center mb-4">
          <h3 className="font-semibold flex items-center gap-2"><Github size={16} /> 从 GitHub 安装</h3>
          <button onClick={onClose}><X size={16} className="text-gray-400" /></button>
        </div>

        <div className="flex gap-2 mb-4">
          <input
            value={url} onChange={e => setUrl(e.target.value)}
            placeholder="https://github.com/user/repo"
            className="flex-1 bg-gray-900 border border-gray-700 rounded-lg px-3 py-2 text-sm outline-none focus:border-indigo-500"
          />
          <button onClick={scan} disabled={scanning || !url} className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50">
            {scanning ? '扫描中...' : '扫描'}
          </button>
        </div>

        {candidates.length > 0 && (
          <>
            <div className="max-h-52 overflow-y-auto space-y-1 mb-4">
              {candidates.map(c => (
                <label key={c.Name} className="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-gray-700 cursor-pointer">
                  <input type="checkbox" checked={selected.has(c.Name)} onChange={() => toggle(c.Name)} className="accent-indigo-500" />
                  <span className="text-sm flex-1">{c.Name}</span>
                  {c.Installed && <span className="text-xs bg-blue-900/50 text-blue-300 px-2 py-0.5 rounded">已安装</span>}
                </label>
              ))}
            </div>
            <div className="flex items-center gap-3 mb-4">
              <span className="text-sm text-gray-400">安装到分类</span>
              <select
                value={category} onChange={e => setCategory(e.target.value)}
                className="bg-gray-900 border border-gray-700 rounded-lg px-3 py-1.5 text-sm flex-1"
              >
                {categories.map(c => <option key={c} value={c}>{c}</option>)}
              </select>
            </div>
            <button
              onClick={install} disabled={installing || selected.size === 0}
              className="w-full py-2 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50"
            >{installing ? '安装中...' : `安装 ${selected.size} 个 Skill`}</button>
          </>
        )}
      </div>
    </div>
  )
}
