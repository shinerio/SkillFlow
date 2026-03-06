import { useEffect, useRef } from 'react'

interface MenuItem { label: string; onClick: () => void; danger?: boolean }
interface Props { x: number; y: number; items: MenuItem[]; onClose: () => void }

export default function ContextMenu({ x, y, items, onClose }: Props) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) onClose()
    }
    document.addEventListener('click', handler)
    return () => document.removeEventListener('click', handler)
  }, [onClose])

  return (
    <div
      ref={ref}
      onClick={e => e.stopPropagation()}
      style={{ position: 'fixed', top: y, left: x, zIndex: 9999 }}
      className="bg-gray-800 border border-gray-700 rounded-lg shadow-xl py-1 min-w-36"
    >
      {items.map((item, i) => (
        <button
          key={i}
          onClick={() => { item.onClick(); onClose() }}
          className={`w-full text-left px-4 py-2 text-sm hover:bg-gray-700 ${item.danger ? 'text-red-400' : 'text-gray-200'}`}
        >
          {item.label}
        </button>
      ))}
    </div>
  )
}
