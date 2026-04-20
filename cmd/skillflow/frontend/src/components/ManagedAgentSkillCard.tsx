import { FolderOpenDot } from 'lucide-react'
import SkillStatusStrip from './SkillStatusStrip'
import { useLanguage } from '../contexts/LanguageContext'

type Props = {
  name: string
  groupName?: string
  pathCount: number
  enabled: boolean
  primaryActionLabel: string
  onPrimaryAction: () => void
  onOpen?: () => void
}

export default function ManagedAgentSkillCard({
  name,
  groupName,
  pathCount,
  enabled,
  primaryActionLabel,
  onPrimaryAction,
  onOpen,
}: Props) {
  const { t } = useLanguage()

  return (
    <div className="card-base relative flex min-h-[148px] flex-col gap-3 p-4">
      <div className="pr-10">
        <SkillStatusStrip
          badges={[
            {
              key: enabled ? 'enabled' : 'disabled',
              label: enabled ? t('toolSkills.enabled') : t('toolSkills.disabled'),
              tone: enabled ? 'success' : 'muted',
            },
            ...(groupName ? [{
              key: `group:${groupName}`,
              label: groupName,
              tone: 'accent' as const,
            }] : [{
              key: 'group:ungrouped',
              label: t('toolSkills.noGroup'),
              tone: 'muted' as const,
            }]),
          ]}
          className="min-w-0"
          singleLine
        />
      </div>

      <p className="min-h-[2.75rem] text-sm font-medium leading-snug line-clamp-2" style={{ color: 'var(--text-primary)' }}>
        {name}
      </p>

      <p className="text-xs" style={{ color: 'var(--text-muted)' }}>
        {t('toolSkills.skillPathCount', { count: pathCount })}
      </p>

      <div className="mt-auto flex items-center gap-2">
        <button
          type="button"
          onClick={onPrimaryAction}
          className="rounded-lg px-3 py-1.5 text-sm transition-all"
          style={{
            background: enabled ? 'rgba(248,113,113,0.12)' : 'var(--accent-glow)',
            color: enabled ? 'var(--color-error)' : 'var(--accent-primary)',
            border: enabled ? '1px solid rgba(248,113,113,0.25)' : '1px solid var(--border-accent)',
          }}
        >
          {primaryActionLabel}
        </button>
        {onOpen && (
          <button
            type="button"
            onClick={onOpen}
            className="ml-auto rounded-lg p-2"
            style={{ background: 'var(--bg-elevated)' }}
            title={t('toolSkills.openDir')}
          >
            <FolderOpenDot size={14} />
          </button>
        )}
      </div>
    </div>
  )
}
