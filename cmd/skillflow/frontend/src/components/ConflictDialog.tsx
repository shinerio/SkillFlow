interface Props {
  conflicts: string[]
  onOverwrite: (name: string) => void
  onSkip: (name: string) => void
  onDone: () => void
}

export default function ConflictDialog({ conflicts, onOverwrite, onSkip, onDone }: Props) {
  if (conflicts.length === 0) { onDone(); return null }
  const current = conflicts[0]
  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
      <div className="bg-gray-800 rounded-2xl p-6 w-96 border border-gray-700">
        <h3 className="text-base font-semibold mb-2">冲突检测</h3>
        <p className="text-sm text-gray-400 mb-6">
          <span className="text-white font-medium">{current}</span> 已存在，如何处理？
        </p>
        <div className="flex gap-3 justify-end">
          <button
            onClick={() => onSkip(current)}
            className="px-4 py-2 text-sm rounded-lg bg-gray-700 hover:bg-gray-600"
          >跳过</button>
          <button
            onClick={() => onOverwrite(current)}
            className="px-4 py-2 text-sm rounded-lg bg-indigo-600 hover:bg-indigo-500"
          >覆盖</button>
        </div>
      </div>
    </div>
  )
}
