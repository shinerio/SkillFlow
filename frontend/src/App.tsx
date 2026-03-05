import { BrowserRouter, Route, Routes, NavLink } from 'react-router-dom'
import { Package, ArrowUpFromLine, ArrowDownToLine, Cloud, Settings, Star } from 'lucide-react'
import Dashboard from './pages/Dashboard'
import SyncPush from './pages/SyncPush'
import SyncPull from './pages/SyncPull'
import Backup from './pages/Backup'
import SettingsPage from './pages/Settings'
import GitHubFavorites from './pages/GitHubFavorites'

export default function App() {
  return (
    <BrowserRouter>
      <div className="flex h-screen bg-gray-950 text-gray-100">
        <aside className="w-56 bg-gray-900 border-r border-gray-800 flex flex-col p-4 gap-1">
          <h1 className="text-lg font-bold mb-6 px-2">SkillFlow</h1>
          <NavItem to="/" icon={<Package size={16} />} label="我的 Skills" />
          <NavItem to="/favorites" icon={<Star size={16} />} label="GitHub 收藏" />
          <p className="text-xs text-gray-500 px-2 mt-3 mb-1">同步管理</p>
          <NavItem to="/sync/push" icon={<ArrowUpFromLine size={16} />} label="推送到工具" />
          <NavItem to="/sync/pull" icon={<ArrowDownToLine size={16} />} label="从工具拉取" />
          <div className="flex-1" />
          <NavItem to="/backup" icon={<Cloud size={16} />} label="云备份" />
          <NavItem to="/settings" icon={<Settings size={16} />} label="设置" />
        </aside>
        <main className="flex-1 overflow-auto">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/favorites" element={<GitHubFavorites />} />
            <Route path="/sync/push" element={<SyncPush />} />
            <Route path="/sync/pull" element={<SyncPull />} />
            <Route path="/backup" element={<Backup />} />
            <Route path="/settings" element={<SettingsPage />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  )
}

function NavItem({ to, icon, label }: { to: string; icon: React.ReactNode; label: string }) {
  return (
    <NavLink
      to={to}
      end
      className={({ isActive }) =>
        `flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${
          isActive ? 'bg-indigo-600 text-white' : 'text-gray-400 hover:bg-gray-800 hover:text-white'
        }`
      }
    >
      {icon}
      {label}
    </NavLink>
  )
}
