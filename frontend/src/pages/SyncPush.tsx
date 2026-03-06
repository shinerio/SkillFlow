import { useEffect, useMemo, useState } from 'react'
import {
  CheckMissingPushDirs,
  GetEnabledTools,
  ListCategories,
  ListSkills,
  PushToTools,
  PushToToolsForce,
} from '../../wailsjs/go/main/App'
import ConflictDialog from '../components/ConflictDialog'
import SyncSkillCard from '../components/SyncSkillCard'
import { ArrowUpFromLine, CheckSquare, FolderPlus, Square, X } from 'lucide-react'
import { ToolIcon } from '../config/toolIcons'

type Scope = 'auto' | 'manual'

export default function SyncPush() {
  const [tools, setTools] = useState<any[]>([])
  const [selectedTools, setSelectedTools] = useState<Set<string>>(new Set())
  const [categories, setCategories] = useState<string[]>([])
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null)
  const [skills, setSkills] = useState<any[]>([])
  const [scope, setScope] = useState<Scope>('auto')
  const [selectedSkills, setSelectedSkills] = useState<Set<string>>(new Set())
  const [conflicts, setConflicts] = useState<string[]>([])
  const [pushing, setPushing] = useState(false)
  const [done, setDone] = useState(false)
  const [missingDirs, setMissingDirs] = useState<{ name: string; dir: string }[]>([])
  const [pendingPush, setPendingPush] = useState(false)

  useEffect(() => {
    Promise.all([GetEnabledTools(), ListSkills(), ListCategories()]).then(([t, s, c]) => {
      setTools(t ?? [])
      setSkills(s ?? [])
      setCategories(c ?? [])
    })
  }, [])

  const filteredSkills = useMemo(
    () => skills.filter((skill: any) => selectedCategory === null || skill.Category === selectedCategory),
    [skills, selectedCategory],
  )

  const pushIDs = useMemo(() => {
    if (scope === 'manual') return Array.from(selectedSkills)
    return filteredSkills.map((skill: any) => skill.ID)
  }, [filteredSkills, scope, selectedSkills])

  const pushCount = pushIDs.length
  const allManualSelected = filteredSkills.length > 0 && selectedSkills.size === filteredSkills.length

  const scopeLabel = scope === 'manual'
    ? `手动选择 ${selectedSkills.size}/${filteredSkills.length}`
    : selectedCategory === null
      ? `全部 Skills (${filteredSkills.length})`
      : `分类「${selectedCategory}」(${filteredSkills.length})`

  const doPush = async () => {
    setPushing(true)
    setDone(false)
    const toolNames = Array.from(selectedTools)
    const result = await PushToTools(pushIDs, toolNames)
    if (result && result.length > 0) {
      setConflicts(result)
    } else {
      setDone(true)
    }
    setPushing(false)
  }

  const push = async () => {
    const toolNames = Array.from(selectedTools)
    const missing = await CheckMissingPushDirs(toolNames)
    if (missing && missing.length > 0) {
      setMissingDirs(missing as { name: string; dir: string }[])
      setPendingPush(true)
      return
    }
    await doPush()
  }

  const confirmMkdirAndPush = async () => {
    setMissingDirs([])
    setPendingPush(false)
    await doPush()
  }

  const toggleTool = (name: string) => {
    setSelectedTools(prev => {
      const next = new Set(prev)
      next.has(name) ? next.delete(name) : next.add(name)
      return next
    })
  }

  const toggleSkill = (id: string) => {
    if (scope !== 'manual') return
    setSelectedSkills(prev => {
      const next = new Set(prev)
      next.has(id) ? next.delete(id) : next.add(id)
      return next
    })
  }

  const toggleAllManual = () => {
    if (allManualSelected) {
      setSelectedSkills(new Set())
      return
    }
    setSelectedSkills(new Set(filteredSkills.map((skill: any) => skill.ID)))
  }

  const setAutoScope = () => {
    setScope('auto')
    setSelectedSkills(new Set())
  }

  const setManualScope = () => {
    setScope('manual')
    setSelectedSkills(new Set(filteredSkills.map((skill: any) => skill.ID)))
  }

  return (
    <div className="flex h-full overflow-hidden">
      <div className="w-48 shrink-0 border-r border-gray-800 p-3 flex flex-col gap-0.5">
        <div className="px-3 py-1.5 text-xs font-medium tracking-wide text-gray-500 uppercase">
          推送范围
        </div>
        <button
          onClick={() => setSelectedCategory(null)}
          className={`px-3 py-2 rounded-lg text-sm text-left transition-colors ${
            selectedCategory === null ? 'bg-indigo-600 text-white' : 'text-gray-400 hover:bg-gray-800'
          }`}
        >
          全部
        </button>
        {categories.map(category => (
          <button
            key={category}
            onClick={() => setSelectedCategory(category)}
            className={`px-3 py-2 rounded-lg text-sm text-left transition-colors ${
              selectedCategory === category ? 'bg-indigo-600 text-white' : 'text-gray-400 hover:bg-gray-800'
            }`}
          >
            {category}
          </button>
        ))}
      </div>

      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-800 flex flex-col gap-4">
          <div className="flex items-center gap-2 text-lg font-semibold">
            <ArrowUpFromLine size={18} />
            推送到工具
          </div>

          <section>
            <p className="text-sm text-gray-400 mb-3">目标工具</p>
            <div className="flex flex-wrap gap-2">
              {tools.map(tool => (
                <button
                  key={tool.name}
                  onClick={() => toggleTool(tool.name)}
                  className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm border transition-colors ${
                    selectedTools.has(tool.name)
                      ? 'bg-indigo-600 border-indigo-500 text-white'
                      : 'bg-gray-800 border-gray-700 text-gray-300 hover:border-gray-500'
                  }`}
                >
                  <ToolIcon name={tool.name} size={20} />
                  {tool.name}
                </button>
              ))}
            </div>
          </section>

          <div className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-2">
              <button
                onClick={setAutoScope}
                className={`px-3 py-1.5 rounded-lg text-sm border transition-colors ${
                  scope === 'auto'
                    ? 'bg-indigo-600 border-indigo-500 text-white'
                    : 'bg-gray-800 border-gray-700 text-gray-300 hover:border-gray-500'
                }`}
              >
                {selectedCategory === null ? '推送全部' : '推送当前分类'}
              </button>
              <button
                onClick={setManualScope}
                className={`px-3 py-1.5 rounded-lg text-sm border transition-colors ${
                  scope === 'manual'
                    ? 'bg-indigo-600 border-indigo-500 text-white'
                    : 'bg-gray-800 border-gray-700 text-gray-300 hover:border-gray-500'
                }`}
              >
                手动选择 Skill
              </button>
            </div>
            <p className="text-sm text-gray-400">{scopeLabel}</p>
          </div>

          {scope === 'manual' && (
            <div className="flex items-center gap-4 text-sm">
              <button
                onClick={toggleAllManual}
                className="flex items-center gap-1.5 text-gray-400 hover:text-white transition-colors"
              >
                {allManualSelected ? <CheckSquare size={14} /> : <Square size={14} />}
                {allManualSelected ? '取消全选' : '全选当前列表'}
              </button>
              <span className="text-gray-500">
                当前可选 {filteredSkills.length} 个 Skill
              </span>
            </div>
          )}
        </div>

        <div className="flex-1 overflow-y-auto p-6">
          <div className="grid grid-cols-3 xl:grid-cols-4 gap-4">
            {filteredSkills.map((skill: any) => (
              <SyncSkillCard
                key={skill.ID}
                id={skill.ID}
                name={skill.Name}
                subtitle={skill.Category || undefined}
                source={skill.Source}
                path={skill.Path}
                selected={scope === 'manual' && selectedSkills.has(skill.ID)}
                showSelection={scope === 'manual'}
                onToggle={() => toggleSkill(skill.ID)}
              />
            ))}
          </div>

          {filteredSkills.length === 0 && (
            <div className="flex flex-col items-center justify-center h-48 text-gray-500">
              <p className="text-sm">当前范围内没有 Skill</p>
              <p className="text-xs mt-1">选择“全部”或切换到其他分类后再试</p>
            </div>
          )}
        </div>

        <div className="px-6 py-4 border-t border-gray-800 shrink-0 flex items-center gap-4">
          <button
            onClick={push}
            disabled={pushing || selectedTools.size === 0 || pushCount === 0}
            className="px-6 py-2 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm disabled:opacity-50"
          >
            {pushing ? '推送中...' : `开始推送 (${pushCount})`}
          </button>
          {done && <span className="text-sm text-green-400">推送完成</span>}
        </div>
      </div>

      {conflicts.length > 0 && (
        <ConflictDialog
          conflicts={conflicts}
          onOverwrite={async (name) => {
            const skill = skills.find(s => s.Name === name)
            if (skill) await PushToToolsForce([skill.ID], Array.from(selectedTools))
            setConflicts(prev => prev.filter(item => item !== name))
          }}
          onSkip={(name) => setConflicts(prev => prev.filter(item => item !== name))}
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
              {missingDirs.map(dir => (
                <li key={dir.name} className="text-sm bg-gray-900 rounded-lg px-3 py-2">
                  <span className="text-gray-300 font-medium">{dir.name}</span>
                  <span className="text-gray-500 text-xs block truncate" title={dir.dir}>{dir.dir}</span>
                </li>
              ))}
            </ul>
            <div className="flex gap-3">
              <button
                onClick={confirmMkdirAndPush}
                className="flex-1 py-2 bg-indigo-600 hover:bg-indigo-500 rounded-lg text-sm"
              >
                创建并推送
              </button>
              <button
                onClick={() => { setMissingDirs([]); setPendingPush(false) }}
                className="flex-1 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm"
              >
                取消
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
