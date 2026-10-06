import { Navigate, Route, Routes } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { request } from './api'
import type { User } from './api'
import { AppShell } from './components/app-shell'
import { LoginPage } from './pages/login'
import { KeysPage } from './pages/keys'
import { RunbookEditorPage } from './pages/runbook-editor'
import { RunbookListPage } from './pages/runbook-list'
import { queryKeys } from './lib/query-keys'
import { Skeleton } from './components/ui/skeleton'

type Wrapped<T, K extends string> = Record<K, T>

export default function App() {
  const me = useQuery({
    queryKey: queryKeys.me,
    queryFn: () => request<Wrapped<User, 'user'>>('/api/auth/me').then((data) => data.user),
  })

  if (me.isLoading) return <div className="grid min-h-screen place-items-center"><div className="flex items-center gap-3" role="status" aria-label="Loading workspace"><Skeleton className="size-8 rounded-lg" /><Skeleton className="h-4 w-32" /></div></div>

  return (
    <Routes>
      <Route path="/login" element={me.data ? <Navigate to="/" replace /> : <LoginPage />} />
      <Route element={me.data ? <AppShell user={me.data} /> : <Navigate to="/login" replace />}>
        <Route index element={<RunbookListPage />} />
        <Route path="/runbooks/new" element={<RunbookEditorPage />} />
        <Route path="/runbooks/:id" element={<RunbookEditorPage />} />
        <Route path="/keys" element={<KeysPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
