import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { KeyRound, Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { request, type APIKey } from '../api'
import { queryKeys } from '../lib/query-keys'
import { relativeTime } from '../lib/time'
import { useDocumentTitle } from '../lib/use-document-title'
import { Badge } from '../components/ui/badge'
import { Button } from '../components/ui/button'
import { Card, CardContent } from '../components/ui/card'
import { PageHeader } from '../components/page-header'
import { Skeleton } from '../components/ui/skeleton'
import { Input } from '../components/ui/input'
import { Label } from '../components/ui/label'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../components/ui/table'

type Wrapped<T, K extends string> = Record<K, T>

export function KeysPage() {
  useDocumentTitle('API keys · Runbook')
  const client = useQueryClient()
  const [name, setName] = useState('')
  const [newKey, setNewKey] = useState<string | null>(null)
  const keys = useQuery({ queryKey: queryKeys.keys, queryFn: () => request<Wrapped<APIKey[], 'keys'>>('/api/keys').then((data) => data.keys) })
  const create = useMutation({
    mutationFn: () => request<{ key: string }>('/api/keys', { method: 'POST', body: JSON.stringify({ name }) }),
    onSuccess: async (data) => {
      setNewKey(data.key)
      setName('')
      await client.invalidateQueries({ queryKey: queryKeys.keys })
      toast.success('API key created — copy it now')
    },
    onError: (error: Error) => toast.error(error.message),
  })
  const revoke = useMutation({
    mutationFn: (id: string) => request(`/api/keys/${id}`, { method: 'DELETE' }),
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: queryKeys.keys })
      toast.success('API key revoked')
    },
    onError: (error: Error) => toast.error(error.message),
  })

  return (
    <main className="mx-auto w-full max-w-[1120px] px-4 py-8 sm:px-7 sm:py-11">
      <PageHeader eyebrow="Machine access" title="API keys" description="Let trusted systems fetch published runbooks." action={<KeyRound className="size-6 text-primary" />} />
      <Card className="border-border/80 bg-card/80"><CardContent className="p-4 sm:p-6">
        <form className="grid gap-3 sm:grid-cols-[1fr_auto] sm:items-end" onSubmit={(event) => { event.preventDefault(); if (name.trim()) create.mutate() }}>
          <div className="grid gap-2"><Label htmlFor="key-name">Key name</Label><Input id="key-name" value={name} onChange={(event) => setName(event.target.value)} placeholder="Production deploy bot" /></div>
          <Button type="submit" disabled={!name.trim() || create.isPending}><Plus />Create API key</Button>
        </form>
        {newKey && <div className="mt-5 rounded-lg border border-primary/30 bg-primary/5 p-4">
          <strong className="text-sm">Copy this key now. It will not be shown again.</strong>
          <code className="mt-3 block break-all rounded-md bg-background/80 p-3 text-xs text-primary">{newKey}</code>
          <Button variant="outline" className="mt-3" onClick={() => { void navigator.clipboard.writeText(newKey).then(() => toast.success('Copied to clipboard')).catch(() => toast.error('Could not access the clipboard')) }}>Copy key</Button>
        </div>}
      </CardContent></Card>
      <Card className="mt-5 overflow-hidden border-border/80 bg-card/80 py-0">
        {keys.isLoading ? <div className="space-y-4 p-5">{[0, 1, 2].map((row) => <Skeleton key={row} className="h-9 w-full" />)}</div> : keys.data?.length ? <div className="overflow-x-auto"><Table>
          <TableHeader><TableRow><TableHead>Name</TableHead><TableHead>Details</TableHead><TableHead className="text-right">Action</TableHead></TableRow></TableHeader>
          <TableBody>{keys.data.map((key) => <TableRow key={key.id}>
            <TableCell className="font-medium">{key.name}</TableCell>
            <TableCell className="min-w-52 text-xs text-muted-foreground">{key.prefix}… · created {relativeTime(key.created_at)}{key.last_used_at && ` · used ${relativeTime(key.last_used_at)}`}</TableCell>
            <TableCell className="text-right">{key.revoked_at ? <Badge variant="secondary">Revoked</Badge> : <Button size="sm" variant="destructive" onClick={() => revoke.mutate(key.id)}><Trash2 />Revoke</Button>}</TableCell>
          </TableRow>)}</TableBody>
        </Table></div> : keys.data && <div className="flex flex-col items-center px-5 py-10 text-center"><KeyRound className="size-6 text-primary" /><p className="mt-3 text-sm text-muted-foreground">No machine keys yet.</p></div>}
      </Card>
      <p className="mt-4 text-xs text-muted-foreground">Send keys using <code>Authorization: Bearer rb_…</code>. Only published runbooks are available to machines.</p>
    </main>
  )
}
