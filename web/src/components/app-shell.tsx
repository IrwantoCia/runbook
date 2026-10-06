import { Link, NavLink, Outlet, useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { BookOpenText, LogOut } from 'lucide-react'
import { toast } from 'sonner'
import { request, type User } from '../api'
import { queryKeys } from '../lib/query-keys'
import { Button } from './ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from './ui/dropdown-menu'

export function AppShell({ user }: { user: User }) {
  const client = useQueryClient()
  const navigate = useNavigate()
  const logout = useMutation({
    mutationFn: () => request('/api/auth/logout', { method: 'POST' }),
    onSuccess: () => {
      client.setQueryData(queryKeys.me, null)
      void client.invalidateQueries({ queryKey: queryKeys.me })
      navigate('/login')
      toast.success('Signed out')
    },
    onError: (error: Error) => toast.error(error.message),
  })

  return (
    <div className="flex min-h-dvh w-full flex-col">
      <header className="sticky top-0 z-20 flex flex-col gap-3 border-b border-border bg-background/90 px-4 py-3 backdrop-blur-md sm:flex-row sm:items-center sm:justify-between sm:px-7 sm:py-4">
        <Link to="/" className="flex items-center gap-2.5 text-lg font-bold tracking-tight">
          <span className="grid size-9 place-items-center rounded-xl bg-primary text-primary-foreground"><BookOpenText className="size-[18px]" /></span>
          runbook <span className="hidden text-[10px] font-bold tracking-[.16em] text-muted-foreground sm:inline">OPS NOTES</span>
        </Link>
        <nav className="flex min-w-0 items-center justify-between gap-1 sm:justify-start sm:gap-3">
          <NavLink to="/" end className={({ isActive }) => `rounded-md border-b-2 px-2 py-2 text-sm ${isActive ? 'border-primary bg-primary/10 font-medium text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}`}>Runbooks</NavLink>
          <NavLink to="/keys" className={({ isActive }) => `rounded-md border-b-2 px-2 py-2 text-sm ${isActive ? 'border-primary bg-primary/10 font-medium text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}`}><span className="hidden sm:inline">API </span>keys</NavLink>
          <DropdownMenu>
            <DropdownMenuTrigger render={<Button variant="outline" size="sm" className="max-w-36 sm:max-w-52" />}>
              <span className="max-w-[9rem] truncate sm:max-w-44">{user.email}</span>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => logout.mutate()} disabled={logout.isPending}>
                <LogOut /> Sign out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </nav>
      </header>
      <div className="flex min-h-0 flex-1 flex-col">
        <Outlet />
      </div>
    </div>
  )
}
