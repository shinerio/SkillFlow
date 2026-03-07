// Tool icon configuration
// SVG icons sourced from LobeHub icon collection (lobehub.com/icons)
// Add new tools by extending toolIconMap with a color and svg component.

interface ToolIconConfig {
  color: string
  svg: (iconSize: number) => React.ReactNode
}

// --- Individual brand SVG icons ---

function ClaudeIcon({ s }: { s: number }) {
  return (
    <svg viewBox="0 0 24 24" width={s} height={s} xmlns="http://www.w3.org/2000/svg">
      <path
        d="M4.709 15.955l4.72-2.647.08-.23-.08-.128H9.2l-.79-.048-2.698-.073-2.339-.097-2.266-.122-.571-.121L0 11.784l.055-.352.48-.321.686.06 1.52.103 2.278.158 1.652.097 2.449.255h.389l.055-.157-.134-.098-.103-.097-2.358-1.596-2.552-1.688-1.336-.972-.724-.491-.364-.462-.158-1.008.656-.722.881.06.225.061.893.686 1.908 1.476 2.491 1.833.365.304.145-.103.019-.073-.164-.274-1.355-2.446-1.446-2.49-.644-1.032-.17-.619a2.97 2.97 0 01-.104-.729L6.283.134 6.696 0l.996.134.42.364.62 1.414 1.002 2.229 1.555 3.03.456.898.243.832.091.255h.158V9.01l.128-1.706.237-2.095.23-2.695.08-.76.376-.91.747-.492.584.28.48.685-.067.444-.286 1.851-.559 2.903-.364 1.942h.212l.243-.242.985-1.306 1.652-2.064.73-.82.85-.904.547-.431h1.033l.76 1.129-.34 1.166-1.064 1.347-.881 1.142-1.264 1.7-.79 1.36.073.11.188-.02 2.856-.606 1.543-.28 1.841-.315.833.388.091.395-.328.807-1.969.486-2.309.462-3.439.813-.042.03.049.061 1.549.146.662.036h1.622l3.02.225.79.522.474.638-.079.485-1.215.62-1.64-.389-3.829-.91-1.312-.329h-.182v.11l1.093 1.068 2.006 1.81 2.509 2.33.127.578-.322.455-.34-.049-2.205-1.657-.851-.747-1.926-1.62h-.128v.17l.444.649 2.345 3.521.122 1.08-.17.353-.608.213-.668-.122-1.374-1.925-1.415-2.167-1.143-1.943-.14.08-.674 7.254-.316.37-.729.28-.607-.461-.322-.747.322-1.476.389-1.924.315-1.53.286-1.9.17-.632-.012-.042-.14.018-1.434 1.967-2.18 2.945-1.726 1.845-.414.164-.717-.37.067-.662.401-.589 2.388-3.036 1.44-1.882.93-1.086-.006-.158h-.055L4.132 18.56l-1.13.146-.487-.456.061-.746.231-.243 1.908-1.312-.006.006z"
        fill="#D97757"
        fillRule="nonzero"
      />
    </svg>
  )
}

function OpenCodeIcon({ s }: { s: number }) {
  return (
    <svg viewBox="0 0 24 24" width={s} height={s} xmlns="http://www.w3.org/2000/svg" fill="#10b981" fillRule="evenodd">
      <path d="M16 6H8v12h8V6zm4 16H4V2h16v20z" />
    </svg>
  )
}

function CodexIcon({ s }: { s: number }) {
  const gradId = 'codex-grad'
  return (
    <svg viewBox="0 0 24 24" width={s} height={s} xmlns="http://www.w3.org/2000/svg">
      <defs>
        <linearGradient id={gradId} gradientUnits="userSpaceOnUse" x1="12" x2="12" y1="3" y2="21">
          <stop stopColor="#B1A7FF" />
          <stop offset=".5" stopColor="#7A9DFF" />
          <stop offset="1" stopColor="#3941FF" />
        </linearGradient>
      </defs>
      <path d="M19.503 0H4.496A4.496 4.496 0 000 4.496v15.007A4.496 4.496 0 004.496 24h15.007A4.496 4.496 0 0024 19.503V4.496A4.496 4.496 0 0019.503 0z" fill="#fff" />
      <path
        d="M9.064 3.344a4.578 4.578 0 012.285-.312c1 .115 1.891.54 2.673 1.275.01.01.024.017.037.021a.09.09 0 00.043 0 4.55 4.55 0 013.046.275l.047.022.116.057a4.581 4.581 0 012.188 2.399c.209.51.313 1.041.315 1.595a4.24 4.24 0 01-.134 1.223.123.123 0 00.03.115c.594.607.988 1.33 1.183 2.17.289 1.425-.007 2.71-.887 3.854l-.136.166a4.548 4.548 0 01-2.201 1.388.123.123 0 00-.081.076c-.191.551-.383 1.023-.74 1.494-.9 1.187-2.222 1.846-3.711 1.838-1.187-.006-2.239-.44-3.157-1.302a.107.107 0 00-.105-.024c-.388.125-.78.143-1.204.138a4.441 4.441 0 01-1.945-.466 4.544 4.544 0 01-1.61-1.335c-.152-.202-.303-.392-.414-.617a5.81 5.81 0 01-.37-.961 4.582 4.582 0 01-.014-2.298.124.124 0 00.006-.056.085.085 0 00-.027-.048 4.467 4.467 0 01-1.034-1.651 3.896 3.896 0 01-.251-1.192 5.189 5.189 0 01.141-1.6c.337-1.112.982-1.985 1.933-2.618.212-.141.413-.251.601-.33.215-.089.43-.164.646-.227a.098.098 0 00.065-.066 4.51 4.51 0 01.829-1.615 4.535 4.535 0 011.837-1.388zm3.482 10.565a.637.637 0 000 1.272h3.636a.637.637 0 100-1.272h-3.636zM8.462 9.23a.637.637 0 00-1.106.631l1.272 2.224-1.266 2.136a.636.636 0 101.095.649l1.454-2.455a.636.636 0 00.005-.64L8.462 9.23z"
        fill={`url(#${gradId})`}
      />
    </svg>
  )
}

function GeminiIcon({ s }: { s: number }) {
  const g0 = 'gem-g0', g1 = 'gem-g1', g2 = 'gem-g2'
  const pathD = "M20.616 10.835a14.147 14.147 0 01-4.45-3.001 14.111 14.111 0 01-3.678-6.452.503.503 0 00-.975 0 14.134 14.134 0 01-3.679 6.452 14.155 14.155 0 01-4.45 3.001c-.65.28-1.318.505-2.002.678a.502.502 0 000 .975c.684.172 1.35.397 2.002.677a14.147 14.147 0 014.45 3.001 14.112 14.112 0 013.679 6.453.502.502 0 00.975 0c.172-.685.397-1.351.677-2.003a14.145 14.145 0 013.001-4.45 14.113 14.113 0 016.453-3.678.503.503 0 000-.975 13.245 13.245 0 01-2.003-.678z"
  return (
    <svg viewBox="0 0 24 24" width={s} height={s} xmlns="http://www.w3.org/2000/svg">
      <defs>
        <linearGradient id={g0} gradientUnits="userSpaceOnUse" x1="7" x2="11" y1="15.5" y2="12">
          <stop stopColor="#08B962" /><stop offset="1" stopColor="#08B962" stopOpacity="0" />
        </linearGradient>
        <linearGradient id={g1} gradientUnits="userSpaceOnUse" x1="8" x2="11.5" y1="5.5" y2="11">
          <stop stopColor="#F94543" /><stop offset="1" stopColor="#F94543" stopOpacity="0" />
        </linearGradient>
        <linearGradient id={g2} gradientUnits="userSpaceOnUse" x1="3.5" x2="17.5" y1="13.5" y2="12">
          <stop stopColor="#FABC12" /><stop offset=".46" stopColor="#FABC12" stopOpacity="0" />
        </linearGradient>
      </defs>
      <path d={pathD} fill="#3186FF" />
      <path d={pathD} fill={`url(#${g0})`} />
      <path d={pathD} fill={`url(#${g1})`} />
      <path d={pathD} fill={`url(#${g2})`} />
    </svg>
  )
}

function OpenClawIcon({ s }: { s: number }) {
  const f0 = 'claw-f0', f1 = 'claw-f1', f2 = 'claw-f2'
  return (
    <svg viewBox="0 0 24 24" width={s} height={s} xmlns="http://www.w3.org/2000/svg">
      <defs>
        <linearGradient id={f0} gradientUnits="userSpaceOnUse" x1="-.659" x2="27.023" y1=".458" y2="22.855">
          <stop stopColor="#FF4D4D" /><stop offset="1" stopColor="#991B1B" />
        </linearGradient>
        <linearGradient id={f1} gradientUnits="userSpaceOnUse" x1="0" x2="4.311" y1="9.672" y2="14.949">
          <stop stopColor="#FF4D4D" /><stop offset="1" stopColor="#991B1B" />
        </linearGradient>
        <linearGradient id={f2} gradientUnits="userSpaceOnUse" x1="19.385" x2="24.399" y1="9.953" y2="14.462">
          <stop stopColor="#FF4D4D" /><stop offset="1" stopColor="#991B1B" />
        </linearGradient>
      </defs>
      <path d="M12 2.568c-6.33 0-9.495 5.275-9.495 9.495 0 4.22 3.165 8.44 6.33 9.494v2.11h2.11v-2.11s1.055.422 2.11 0v2.11h2.11v-2.11c3.165-1.055 6.33-5.274 6.33-9.494S18.33 2.568 12 2.568z" fill={`url(#${f0})`} />
      <path d="M3.56 9.953C.396 8.898-.66 11.008.396 13.118c1.055 2.11 3.164 1.055 4.22-1.055.632-1.477 0-2.11-1.056-2.11z" fill={`url(#${f1})`} />
      <path d="M20.44 9.953c3.164-1.055 4.22 1.055 3.164 3.165-1.055 2.11-3.164 1.055-4.22-1.055-.632-1.477 0-2.11 1.056-2.11z" fill={`url(#${f2})`} />
      <path d="M5.507 1.875c.476-.285 1.036-.233 1.615.037.577.27 1.223.774 1.937 1.488a.316.316 0 01-.447.447c-.693-.693-1.279-1.138-1.757-1.361-.475-.222-.795-.205-1.022-.069a.317.317 0 01-.326-.542zM16.877 1.913c.58-.27 1.14-.323 1.616-.038a.317.317 0 01-.326.542c-.227-.136-.547-.153-1.022.069-.478.223-1.064.668-1.756 1.361a.316.316 0 11-.448-.447c.714-.714 1.36-1.218 1.936-1.487z" fill="#FF4D4D" />
      <path d="M8.835 9.109a1.266 1.266 0 100-2.532 1.266 1.266 0 000 2.532zM15.165 9.109a1.266 1.266 0 100-2.532 1.266 1.266 0 000 2.532z" fill="#050810" />
      <path d="M9.046 8.16a.527.527 0 100-1.056.527.527 0 000 1.055zM15.376 8.16a.527.527 0 100-1.055.527.527 0 000 1.054z" fill="#00E5CC" />
    </svg>
  )
}

function DefaultToolIcon({ s }: { s: number }) {
  return (
    <svg viewBox="0 0 24 24" width={s} height={s} xmlns="http://www.w3.org/2000/svg" fill="none" stroke="#9ca3af" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
      <path d="M14.7 6.3a1 1 0 000 1.4l1.6 1.6a1 1 0 001.4 0l3.77-3.77a6 6 0 01-7.94 7.94l-6.91 6.91a2.12 2.12 0 01-3-3l6.91-6.91a6 6 0 017.94-7.94l-3.76 3.76z" />
    </svg>
  )
}

// --- Icon registry ---

const toolIconMap: Record<string, ToolIconConfig> = {
  'claude-code': { color: '#D97757', svg: (s) => <ClaudeIcon s={s} /> },
  'opencode':    { color: '#10b981', svg: (s) => <OpenCodeIcon s={s} /> },
  'codex':       { color: '#7A9DFF', svg: (s) => <CodexIcon s={s} /> },
  'gemini-cli':  { color: '#3186FF', svg: (s) => <GeminiIcon s={s} /> },
  'openclaw':    { color: '#FF4D4D', svg: (s) => <OpenClawIcon s={s} /> },
}

const defaultConfig: ToolIconConfig = { color: '#6b7280', svg: (s) => <DefaultToolIcon s={s} /> }

export function getToolIconConfig(name: string): ToolIconConfig {
  return toolIconMap[name] ?? defaultConfig
}

interface ToolIconProps {
  name: string
  size?: number
}

export function ToolIcon({ name, size = 28 }: ToolIconProps) {
  const cfg = getToolIconConfig(name)
  const padding = Math.round(size * 0.18)
  const iconSize = size - padding * 2
  return (
    <div
      style={{
        width: size,
        height: size,
        backgroundColor: cfg.color + '22',
        borderRadius: Math.round(size * 0.28),
        border: `1px solid ${cfg.color}44`,
        flexShrink: 0,
      }}
      className="flex items-center justify-center select-none"
    >
      {cfg.svg(iconSize)}
    </div>
  )
}
