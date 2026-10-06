import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { zodResolver } from '@hookform/resolvers/zod'
import { BookOpenText } from 'lucide-react'
import { toast } from 'sonner'
import { request, type User } from '../api'
import { queryKeys } from '../lib/query-keys'
import { useDocumentTitle } from '../lib/use-document-title'
import { Button } from '../components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card'
import { Input } from '../components/ui/input'
import { Label } from '../components/ui/label'

const loginSchema = z.object({ email: z.email(), password: z.string().min(1, 'Password is required') })
type Wrapped<T, K extends string> = Record<K, T>

export function LoginPage() {
  useDocumentTitle('Sign in · Runbook')
  const client = useQueryClient()
  const navigate = useNavigate()
  const form = useForm<z.infer<typeof loginSchema>>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: '', password: '' },
  })
  const login = useMutation({
    mutationFn: (input: z.infer<typeof loginSchema>) => request<Wrapped<User, 'user'>>('/api/auth/login', { method: 'POST', body: JSON.stringify(input) }),
    onSuccess: (data) => {
      client.setQueryData(queryKeys.me, data.user)
      navigate('/')
      toast.success('Welcome back')
    },
    onError: (error: Error) => toast.error(error.message),
  })

  return (
    <main className="grid min-h-screen place-items-center p-5">
      <Card className="w-full max-w-[420px] border-border/80 bg-card/90 p-1 shadow-2xl shadow-black/20">
        <CardHeader className="px-7 pt-7">
          <div className="flex items-center gap-2.5 text-lg font-bold tracking-tight"><span className="grid size-9 place-items-center rounded-xl bg-primary text-primary-foreground"><BookOpenText className="size-[18px]" /></span>runbook</div>
          <p className="pt-5 text-[11px] font-bold uppercase tracking-[.16em] text-primary">Your operations library</p>
          <CardTitle className="text-3xl tracking-tight">Sign in</CardTitle>
          <CardDescription>Pick up where your team left off.</CardDescription>
        </CardHeader>
        <CardContent>
          <form className="grid gap-4" onSubmit={form.handleSubmit((values) => login.mutate(values))}>
            <div className="grid gap-2">
              <Label htmlFor="email">Email</Label>
              <Input id="email" type="email" autoComplete="username" autoFocus {...form.register('email')} />
              {form.formState.errors.email && <p className="text-sm text-destructive">{form.formState.errors.email.message}</p>}
            </div>
            {login.error && <p role="alert" className="text-sm text-destructive">{login.error.message}</p>}
            <div className="grid gap-2">
              <Label htmlFor="password">Password</Label>
              <Input id="password" type="password" autoComplete="current-password" {...form.register('password')} />
              {form.formState.errors.password && <p className="text-sm text-destructive">{form.formState.errors.password.message}</p>}
            </div>
            <Button type="submit" className="mt-2 w-full" disabled={login.isPending}>{login.isPending ? 'Signing in…' : 'Sign in'}</Button>
          </form>
        </CardContent>
      </Card>
    </main>
  )
}
