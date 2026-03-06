import { useEffect, useState } from 'react'
import { GetEnabledTools, ListSkills, ListCategories, PushToTools, PushToToolsForce, CheckMissingPushDirs } from '../../wailsjs/go/main/App'
import ConflictDialog from '../components/ConflictDialog'
import SyncSkillCard from '../components/SyncSkillCard'
import { ArrowUpFromLine, CheckSquare, Square, FolderPlus, X } from 'lucide-react'
import { ToolIcon } from '../config/toolIcons'

type Scope = 'all' | 'category' | 'manual'

export default function SyncPush() {
  const [tools, setTools] = useState<any[]>([])
  const [selectedTools, setSelectedTools] = useState<Set<string>>(new Set())
  const [scope, setScope] = useState<Scope>('all')
  const [categories, setCategories] = useState<string[]>([])
  const [selectedCategory, setSelectedCategory] = useState('')
  const [skills, setSkills] = useState<any[]>([])
  const [selectedSkills, setSelectedSkills] = useState<Set<string>>(new Set())
  const [conflicts, setConflicts] = useState<string[]>([])
  const [pushing, setPushing] = useState(false)
  const [done, setDone] = useState(false)
  const [missingDirs, setMissingDirs] = useState<{name: string, dir: string}[]>([])
  const [pendingPush, setPendingPush] = useState(false)

  useEffect(() => {
    Promise.all([GetEnabledTools(), ListSkills(), ListCategories()]).then(([t, s, c]) => {
      setTools(t ?? [])
      setSkills(s ?? [])
      setCategories(c ?? [])
    })
  }, [])

  const getSkillIDs = () => {
    if (scope === 'all') return skills.map(s => s.ID)
    if (scope === 'category') return skills.filter(s => s.Category === selectedCategory).map(s => s.ID)
    return [...selectedSkills]
  }

  const manualSkills = scope === 'manual' ? skills : scope === 'category' && selectedCategory
    ? skills.filter(s => s.Category === selectedCategory)
    : skills

  const doPush = async () => {
    setPushing(true)
    setDone(false)
    const ids = getSkillIDs()
    const toolNames = [...selectedTools]
    const result = await PushToTools(ids, toolNames)
    if (result && result.length > 0) {
      setConflicts(result)
    } else {
      setDone(true)
    }
    setPushing(false)
  }

  const push = async () => {
    const toolNames = [...selectedTools]
    const missing = await CheckMissingPushDirs(toolNames)
    if (missing && missing.length > 0) {
      setMissingDirs(missing as {name: string, dir: string}[])
      setPendingPush(true)
    } else {
      await doPush()
    }
  }

  const confirmMkdirAndPush = async () => {
    setMissingDirs([])
    setPendingPush(false)
    await doPush()
  }

  const toggleTool = (name: string) => {
    const next = new Set(selectedTools)
    next.has(name) ? next.delete(name) : next.add(name)
    setSelectedTools(next)
  }

  const toggleSkill = (id: string) => {
    const next = new Set(selectedSkills)
    next.has(id) ? next.delete(id) : next.add(id)
    setSelectedSkills(next)
  }

  const toggleAllManual = () => {
    if (selectedSkills.size === skills.length) {
      setSelectedSkills(new Set())
    } else {
      setSelectedSkills(new Set(skills.map(s => s.ID)))
    }
  }

  const allManualSelected = skills.length > 0 && selectedSkills.size === skills.length

  // Computed: how many skills will be pushed
  const pushCount = scope === 'all'
    ? skills.length
    : scope === 'category'
      ? skills.filter(s => s.Category === selectedCategory).length
      : selectedSkills.size

  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div className="p-8 pb-0 shrink-0">
        <h2 className="text-lg font-semibold mb-6 flex items-center gap-2">
          <ArrowUpFromLine size={18} /> 推送到工具
        </h2>

        {/* Tool selection */}
        <section className="mb-6">
          <p className="text-sm text-gray-400 mb-3">目标工具</p>
          <div className="flex flex-wrap gap-2">
            {tools.map(t => (
              <button
                key={t.name}
                onClick={() => toggleTool(t.name)}
                className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm border transition-colors ${
                  selectedTools.has(t.name)
                    ? 'bg-indigo-600 border-indigo-500 text-white'
                    : 'bg-gray-800 border-gray-700 text-gray-300 hover:border-gray-500'
                }`}
              >
                <ToolIcon name={t.name} size={20} />
                {t.name}
              </button>
            ))}
          </div>
        </section>

        {/* Scope selection */}
        <section className="mb-4">
          <p className="text-sm text-gray-400 mb-3">同步范围</p>
          <div className="flex gap-3">
            {([['all', '全部 Skills'], ['category', '按分类'], ['manual', '手动选择']] as [Scope, string][]).map(([v, label]) => (
              <button
                key={v}
                onClick={() => setScope(v)}
                className={`px-4 py-1.5 rounded-lg text-sm border transition-colors ${
                  scope === v
                    ? 'bg-indigo-600 border-indigo-500 text-white'
                    : 'bg-gray-800 border-gray-700 text-gray-300 hover:border-gray-500'
                }`}
              >
                {label}
              </button>
            ))}
          </div>

          {scope === 'category' && (
            <select
              value={selectedCategory}
              onChange={e => setSelectedCategory(e.target.value)}
              className="mt-3 bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm w-48"
            >
              <option value="">选择分类</option>
              {categories.map(c => <option key={c} value={c}>{c}</option>)}
            </select>
          )}
        </section>
      </div>

      {/* Card grid for manual / category preview */}
      {(scope === 'manual' || (scope === 'category' && selectedCategory)) && (
        <>
          <div className="px-8 mb-3 flex items-center gap-4 shrink-0">
            <p className="text-sm text-gray-400">
              {scope === 'manual' ? '选择要推送的 Skills' : `分类「${selectedCategory}」中的 Skills`}
              <span className="ml-1 text-gray-500">
                （{scope === 'manual' ? `${selectedSkills.size}/` : ''}{manualSkills.length}）
              </span>
            </p>
            {scope === 'manual' && (
              <button
                onClick={toggleAllManual}
                className="flex items-center gap-1.5 text-xs text-gray-400 hover:text-white transition-colors"
              >
                {allManualSelected ? <CheckSquare size={13} /> : <Square size={13} />}
                {allManualSelected ? '取消全选' : '全选'}
              </button>
            )}
          </div>

          <div className="flex-1 overflow-y-auto px-8">
            <div className="grid grid-cols-3 xl:grid-cols-4 gap-3 pb-4">
              {manualSkills.map((sk: any) => (
                <SyncSkillCard
                  key={sk.ID}
                  id={sk.ID}
                  name={sk.Name}
                  subtitle={sk.Category || undefined}
                  source={sk.Source}
                  path={sk.Path}
                  selected={scope === 'manual' ? selectedSkills.has(sk.ID) : true}
                  onToggle={() => scope === 'manual' && toggleSkill(sk.ID)}
                />
              ))}
            </div>
          </div>
        </>
      )}

      {/* "全部" summary when no grid shown */}
      {scope === 'all' && (
        <div className="px-8 mt-2 flex-1">
          <p className="text-sm text-gray-500">
            将推送全部 <span className="text-white">{skills.length}</span> 个 Skills
          </p>
        </div>
      )}

      {/* Bottom action bar */}
      <div className="px-8 py-4 border-t border-gray-800 shrink-0 flex items-center gap-4">
        <button
          onClick={push}
          disabled={pushing || selectedTools.size === 0 || pushCount === 0}
          className="px-6 py-2 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50"
        >
          {pushing ? '推送中...' : `开始推送 (${pushCount})`}
        </button>
        {done && <span className="text-sm text-green-400">推送完成 ✓</span>}
      </div>

      {conflicts.length > 0 && (
        <ConflictDialog
          conflicts={conflicts}
          onOverwrite={async (name) => {
            const sk = skills.find(s => s.Name === name)
            if (sk) await PushToToolsForce([sk.ID], [...selectedTools])
            setConflicts(prev => prev.filter(c => c !== name))
          }}
          onSkip={(name) => setConflicts(prev => prev.filter(c => c !== name))}
          onDone={() => setDone(true)}
        />
      )}

      {pendingPush && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-2xl p-6 w-[460px] border border-gray-700">
            <div className="flex justify-between items-center mb-1">
              <h3 className="font-semibold flex items-center gap-2"><FolderPlus size={16} /> 目录不存在</h3>
              <button onClick={() => { setMissingDirs([]); setPendingPush(false) }}><X size={16} className="text-gray-400" /></button>
            </div>
            <p className="text-xs text-gray-500 mb-3">以下推送目录尚未创建，是否自动创建后继续推送？</p>
            <ul className="space-y-1.5 mb-4 max-h-40 overflow-y-auto">
              {missingDirs.map(d => (
                <li key={d.name} className="text-sm bg-gray-900 rounded-lg px-3 py-2">
                  <span className="text-gray-300 font-medium">{d.name}</span>
                  <span className="text-gray-500 text-xs block truncate" title={d.dir}>{d.dir}</span>
                </li>
              ))}
            </ul>
            <div className="flex gap-3">
              <button onClick={confirmMkdirAndPush}
                className="flex-1 py-2 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm">
                创建并推送
              </button>
              <button onClick={() => { setMissingDirs([]); setPendingPush(false) }}
                className="flex-1 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm">取消</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
