import { useEffect, useState } from 'react'
import { GetEnabledTools, ListSkills, ListCategories, PushToTools, PushToToolsForce } from '../../wailsjs/go/main/App'
import ConflictDialog from '../components/ConflictDialog'
import { ArrowUpFromLine } from 'lucide-react'
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

  const push = async () => {
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

  return (
    <div className="p-8 max-w-2xl">
      <h2 className="text-lg font-semibold mb-6 flex items-center gap-2"><ArrowUpFromLine size={18} /> 推送到工具</h2>

      {/* Tool selection */}
      <section className="mb-6">
        <p className="text-sm text-gray-400 mb-3">目标工具</p>
        <div className="flex flex-wrap gap-2">
          {tools.map(t => (
            <button
              key={t.name}
              onClick={() => toggleTool(t.name)}
              className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm border transition-colors ${selectedTools.has(t.name) ? 'bg-indigo-600 border-indigo-500 text-white' : 'bg-gray-800 border-gray-700 text-gray-300 hover:border-gray-500'}`}
            >
              <ToolIcon name={t.name} size={20} />
              {t.name}
            </button>
          ))}
        </div>
      </section>

      {/* Scope selection */}
      <section className="mb-6">
        <p className="text-sm text-gray-400 mb-3">同步范围</p>
        <div className="space-y-2">
          {([['all', '全部 Skills'], ['category', '按分类'], ['manual', '手动选择']] as [Scope, string][]).map(([v, label]) => (
            <label key={v} className="flex items-center gap-3 cursor-pointer">
              <input type="radio" checked={scope === v} onChange={() => setScope(v)} className="accent-indigo-500" />
              <span className="text-sm">{label}</span>
            </label>
          ))}
        </div>

        {scope === 'category' && (
          <select value={selectedCategory} onChange={e => setSelectedCategory(e.target.value)}
            className="mt-3 bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm w-48">
            <option value="">选择分类</option>
            {categories.map(c => <option key={c} value={c}>{c}</option>)}
          </select>
        )}

        {scope === 'manual' && (
          <div className="mt-3 max-h-52 overflow-y-auto space-y-1 border border-gray-700 rounded-xl p-3">
            {skills.map(sk => (
              <label key={sk.ID} className="flex items-center gap-3 px-2 py-1.5 hover:bg-gray-800 rounded-lg cursor-pointer">
                <input type="checkbox" checked={selectedSkills.has(sk.ID)} onChange={() => toggleSkill(sk.ID)} className="accent-indigo-500" />
                <span className="text-sm">{sk.Name}</span>
                <span className="text-xs text-gray-500">{sk.Category || '未分类'}</span>
              </label>
            ))}
          </div>
        )}
      </section>

      <button
        onClick={push}
        disabled={pushing || selectedTools.size === 0}
        className="px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50"
      >{pushing ? '推送中...' : '开始推送'}</button>

      {done && <p className="mt-4 text-sm text-green-400">推送完成</p>}

      {conflicts.length > 0 && (
        <ConflictDialog
          conflicts={conflicts}
          onOverwrite={async (name) => {
            const skill = skills.find(s => s.Name === name)
            if (skill) await PushToToolsForce([skill.ID], [...selectedTools])
            setConflicts(prev => prev.filter(c => c !== name))
          }}
          onSkip={(name) => setConflicts(prev => prev.filter(c => c !== name))}
          onDone={() => setDone(true)}
        />
      )}
    </div>
  )
}
