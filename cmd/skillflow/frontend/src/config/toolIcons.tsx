// Tool icon configuration
// Edit this file to customize the icon and color for each tool.
// - emoji: any emoji character used as the icon
// - color: hex color for the icon background tint
// Tools not listed here will use the default config below.

export interface ToolIconConfig {
  emoji: string
  color: string
}

export const toolIconMap: Record<string, ToolIconConfig> = {
  'claude-code': { emoji: '🤖', color: '#6366f1' },
  'opencode':    { emoji: '💻', color: '#10b981' },
  'codex':       { emoji: '📝', color: '#f59e0b' },
  'gemini-cli':  { emoji: '✨', color: '#8b5cf6' },
  'openclaw':    { emoji: '🦞', color: '#ef4444' },
}

const defaultConfig: ToolIconConfig = { emoji: '🔧', color: '#6b7280' }

export function getToolIconConfig(name: string): ToolIconConfig {
  return toolIconMap[name] ?? defaultConfig
}

interface ToolIconProps {
  name: string
  size?: number
}

export function ToolIcon({ name, size = 28 }: ToolIconProps) {
  const cfg = getToolIconConfig(name)
  return (
    <div
      style={{
        width: size,
        height: size,
        fontSize: Math.round(size * 0.62),
        backgroundColor: cfg.color + '33',
        borderRadius: Math.round(size * 0.28),
        border: `1px solid ${cfg.color}55`,
        flexShrink: 0,
      }}
      className="flex items-center justify-center select-none"
    >
      {cfg.emoji}
    </div>
  )
}
