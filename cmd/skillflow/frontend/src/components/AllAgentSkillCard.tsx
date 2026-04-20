import { useRef } from 'react'
import { Layers3, Tags } from 'lucide-react'
import SkillStatusStrip from './SkillStatusStrip'
import { useLanguage } from '../contexts/LanguageContext'

type Props = {
  name: string
  groupName?: string
  agents: string[]
  instanceCount: number
  dragging: boolean
  onDragStateChange: (skillName: string | null) => void
}

export default function AllAgentSkillCard({
  name,
  groupName,
  agents,
  instanceCount,
  dragging,
  onDragStateChange,
}: Props) {
  const { t } = useLanguage()
  const cardRef = useRef<HTMLDivElement>(null)
  const dragGhostRef = useRef<HTMLDivElement | null>(null)

  const setCardDragImage = (event: React.DragEvent) => {
    if (!cardRef.current) return
    const clone = cardRef.current.cloneNode(true) as HTMLDivElement
    const rect = cardRef.current.getBoundingClientRect()
    clone.style.width = `${Math.max(rect.width * 0.82, 180)}px`
    clone.style.transform = 'scale(0.82)'
    clone.style.transformOrigin = 'top left'
    clone.style.opacity = '0.96'
    clone.style.pointerEvents = 'none'
    clone.style.position = 'fixed'
    clone.style.top = '-1000px'
    clone.style.left = '-1000px'
    clone.style.zIndex = '9999'
    document.body.appendChild(clone)
    dragGhostRef.current = clone
    event.dataTransfer.setDragImage(clone, 24, 18)
  }

  const cleanupDragGhost = () => {
    if (dragGhostRef.current?.parentNode) {
      dragGhostRef.current.parentNode.removeChild(dragGhostRef.current)
    }
    dragGhostRef.current = null
  }

  return (
    <div
      ref={cardRef}
      draggable
      onDragStart={event => {
        event.dataTransfer.setData('text/plain', name)
        event.dataTransfer.setData('application/x-skillflow-agent-skill-name', name)
        event.dataTransfer.effectAllowed = 'move'
        setCardDragImage(event)
        onDragStateChange(name)
      }}
      onDragEnd={() => {
        cleanupDragGhost()
        onDragStateChange(null)
      }}
      className={`card-base relative flex min-h-[132px] flex-col gap-3 p-4 cursor-grab select-none transition-all ${dragging ? 'opacity-55 scale-[0.96]' : ''}`}
    >
      <div className="flex items-center gap-2">
        <SkillStatusStrip
          badges={groupName ? [{
            key: `group:${groupName}`,
            label: groupName,
            tone: 'accent',
          }] : [{
            key: 'group:ungrouped',
            label: t('toolSkills.noGroup'),
            tone: 'muted',
          }]}
          className="min-w-0 flex-1"
          singleLine
        />
        <Layers3 size={14} style={{ color: 'var(--text-muted)' }} className="shrink-0" />
      </div>

      <p className="min-h-[2.75rem] text-sm font-medium leading-snug line-clamp-2" style={{ color: 'var(--text-primary)' }}>
        {name}
      </p>

      <div className="mt-auto space-y-1">
        <div className="flex items-center gap-2 text-xs" style={{ color: 'var(--text-muted)' }}>
          <Tags size={12} className="shrink-0" />
          <span>{t('toolSkills.allSkillMeta', { agents: agents.join(', '), count: instanceCount })}</span>
        </div>
      </div>
    </div>
  )
}
