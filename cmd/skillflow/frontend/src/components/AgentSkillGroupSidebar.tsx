import { useState } from 'react'
import { Layers3, Plus, Tags } from 'lucide-react'
import ContextMenu from './ContextMenu'
import { useLanguage } from '../contexts/LanguageContext'

export type AgentSkillGroupSidebarItem = {
  key: string
  label: string
  count: number
  kind: 'all' | 'ungrouped' | 'group'
}

type Props = {
  items: AgentSkillGroupSidebarItem[]
  selectedKey: string
  draggingSkillName: string | null
  onSelect: (key: string) => void
  onCreateGroup: () => void
  onRenameGroup: (groupName: string) => void
  onDeleteGroup: (groupName: string) => void
  onDropSkill: (skillName: string, target: AgentSkillGroupSidebarItem) => void
}

export default function AgentSkillGroupSidebar({
  items,
  selectedKey,
  draggingSkillName,
  onSelect,
  onCreateGroup,
  onRenameGroup,
  onDeleteGroup,
  onDropSkill,
}: Props) {
  const { t } = useLanguage()
  const [menu, setMenu] = useState<{ x: number; y: number; item: AgentSkillGroupSidebarItem } | null>(null)
  const [dragTargetKey, setDragTargetKey] = useState<string | null>(null)

  const acceptsDrop = draggingSkillName !== null

  const handleDragOver = (event: React.DragEvent, item: AgentSkillGroupSidebarItem) => {
    if (!acceptsDrop || item.kind === 'all') return
    event.preventDefault()
    event.dataTransfer.dropEffect = 'move'
    setDragTargetKey(item.key)
  }

  const handleDragLeave = (item: AgentSkillGroupSidebarItem) => {
    if (dragTargetKey === item.key) {
      setDragTargetKey(null)
    }
  }

  const handleDrop = (event: React.DragEvent, item: AgentSkillGroupSidebarItem) => {
    if (item.kind === 'all') return
    event.preventDefault()
    setDragTargetKey(null)
    const draggedSkillName = event.dataTransfer.getData('application/x-skillflow-agent-skill-name')
      || event.dataTransfer.getData('text/plain')
      || draggingSkillName
      || ''
    if (draggedSkillName) {
      onDropSkill(draggedSkillName, item)
    }
  }

  const getItemStyle = (active: boolean, dragTarget: boolean) => {
    if (dragTarget) {
      return {
        background: 'var(--active-surface)',
        color: 'var(--active-text)',
        border: '1px solid var(--active-border)',
        boxShadow: 'var(--active-shadow)',
      }
    }
    if (active) {
      return {
        background: 'var(--active-surface)',
        color: 'var(--active-text)',
        border: '1px solid var(--active-border)',
        boxShadow: 'var(--active-shadow)',
      }
    }
    return {
      color: 'var(--text-muted)',
      border: '1px solid transparent',
    }
  }

  return (
    <div
      className="w-52 shrink-0 h-full min-h-0 overflow-y-auto p-3 flex flex-col gap-1"
      style={{ borderRight: '1px solid var(--border-base)' }}
    >
      <div className="px-3 py-1.5 text-xs font-medium tracking-wide uppercase" style={{ color: 'var(--text-muted)' }}>
        {t('toolSkills.groupSidebarTitle')}
      </div>

      {items.map(item => (
        <button
          key={item.key}
          type="button"
          onClick={() => onSelect(item.key)}
          onContextMenu={event => {
            if (item.kind !== 'group') return
            event.preventDefault()
            setMenu({ x: event.clientX, y: event.clientY, item })
          }}
          onDragEnter={event => handleDragOver(event, item)}
          onDragOver={event => handleDragOver(event, item)}
          onDragLeave={() => handleDragLeave(item)}
          onDrop={event => handleDrop(event, item)}
          className="flex items-center gap-2 rounded-lg px-3 py-2 text-sm text-left transition-all duration-150"
          style={getItemStyle(selectedKey === item.key, dragTargetKey === item.key)}
        >
          {item.kind === 'all' ? <Layers3 size={16} /> : <Tags size={16} />}
          <span className="truncate flex-1">{item.label}</span>
          <span className="text-[11px]" style={{ color: selectedKey === item.key ? 'var(--active-text)' : 'var(--text-disabled)' }}>
            {item.count}
          </span>
        </button>
      ))}

      <button
        type="button"
        onClick={onCreateGroup}
        className="mt-2 flex items-center gap-1.5 px-3 py-2 text-sm transition-colors"
        style={{ color: 'var(--text-muted)' }}
      >
        <Plus size={14} />
        {t('toolSkills.createGroup')}
      </button>

      {menu && (
        <ContextMenu
          x={menu.x}
          y={menu.y}
          items={[
            { label: t('category.rename'), onClick: () => onRenameGroup(menu.item.label) },
            { label: t('common.delete'), onClick: () => onDeleteGroup(menu.item.label), danger: true },
          ]}
          onClose={() => setMenu(null)}
        />
      )}
    </div>
  )
}
