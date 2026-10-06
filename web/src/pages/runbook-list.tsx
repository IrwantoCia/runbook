import { useMemo, useState } from 'react'
import { Link, useSearchParams } from 'react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { FolderInput, Plus, Terminal } from 'lucide-react'
import { toast } from 'sonner'
import { folderRequests, request, type Folder, type Runbook } from '../api'
import { queryKeys } from '../lib/query-keys'
import { folderPath } from '../lib/folder-path'
import { relativeTime } from '../lib/time'
import { useDocumentTitle } from '../lib/use-document-title'
import { Alert, AlertDescription, AlertTitle } from '../components/ui/alert'
import { Button } from '../components/ui/button'
import { Card, CardContent } from '../components/ui/card'
import { PageHeader } from '../components/page-header'
import { Skeleton } from '../components/ui/skeleton'
import { StatusBadge } from '../components/status-badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../components/ui/table'
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '../components/ui/dialog'
import { FolderNavigation } from '../components/folder-navigation'

type Wrapped<T, K extends string> = Record<K, T>

export function RunbookListPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const client = useQueryClient()
  const [moveBook, setMoveBook] = useState<Runbook | null>(null)
  const [destination, setDestination] = useState('')
  const requestedStatus = searchParams.get('status')
  const status = requestedStatus === 'draft' || requestedStatus === 'published' ? requestedStatus : 'all'
  const folderID = searchParams.get('folder') || null
  const folders = useQuery({ queryKey: queryKeys.folders, queryFn: folderRequests.list })
  const books = useQuery({
    queryKey: [...queryKeys.books, status, folderID],
    queryFn: () => {
      const params = new URLSearchParams()
      if (status !== 'all') params.set('status', status)
      if (folderID) params.set('folder_id', folderID)
      return request<Wrapped<Runbook[], 'runbooks'>>(`/api/runbooks${params.size ? `?${params}` : ''}`).then((data) => data.runbooks)
    },
  })
  const folder = folders.data?.find((item) => item.id === folderID)
  const breadcrumbs = useMemo(() => {
    const chain: Folder[] = []
    let current = folder
    while (current) {
      chain.unshift(current)
      current = folders.data?.find((item) => item.id === current?.parent_id)
    }
    return chain
  }, [folder, folders.data])
  useDocumentTitle(folder ? `${folder.name} · Runbooks` : 'Runbooks · Runbook')

  const selectFolder = (id: string | null) => {
    const next = new URLSearchParams(searchParams)
    if (id) next.set('folder', id)
    else next.delete('folder')
    setSearchParams(next)
  }
  const selectStatus = (nextStatus: string) => {
    const next = new URLSearchParams(searchParams)
    if (nextStatus === 'all') next.delete('status')
    else next.set('status', nextStatus)
    setSearchParams(next)
  }
  const move = useMutation({
    mutationFn: () => moveBook ? folderRequests.moveRunbook(moveBook.id, destination || null) : Promise.reject(new Error('Choose a runbook')),
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: queryKeys.books })
      toast.success('Runbook moved')
      setMoveBook(null)
    },
    onError: (error: Error) => toast.error(error.message),
  })

  return <main className="mx-auto flex w-full max-w-[1440px] flex-1 flex-col gap-6 px-4 py-6 sm:px-7 sm:py-9 lg:flex-row">
    <aside className="w-full shrink-0 lg:w-64"><FolderNavigation selected={folderID} onSelect={selectFolder} /></aside>
    <section className="min-w-0 flex-1">
      <PageHeader eyebrow="Operational knowledge" title={folder?.name ?? 'Runbooks'} description="The steps that make the hard days routine." action={<Button render={<Link to={`/runbooks/new${folderID ? `?folder=${encodeURIComponent(folderID)}` : ''}`} />}><Plus />New runbook</Button>} />
      <nav aria-label="Folder breadcrumbs" className="mb-4 flex min-w-0 flex-wrap items-center gap-2 text-sm text-muted-foreground"><button type="button" className="cursor-pointer hover:text-foreground" onClick={() => selectFolder(null)}>All runbooks</button>{breadcrumbs.map((item) => <span key={item.id} className="flex min-w-0 items-center gap-2"><span aria-hidden="true">/</span><button type="button" className={`max-w-48 truncate cursor-pointer hover:text-foreground ${item.id === folderID ? 'font-medium text-foreground' : ''}`} onClick={() => selectFolder(item.id)}>{item.name}</button></span>)}</nav>
      <div role="group" aria-label="Filter runbooks by status" className="mb-5 flex w-full flex-wrap gap-1 rounded-lg border border-border bg-card/60 p-1 sm:w-fit">
        {(['all', 'draft', 'published'] as const).map((filter) => <Button key={filter} size="sm" variant={status === filter ? 'secondary' : 'ghost'} aria-pressed={status === filter} onClick={() => selectStatus(filter)} className="capitalize">{filter === 'all' ? 'All' : filter}</Button>)}
      </div>
      {books.isLoading && <Card className="border-border/80 bg-card/80"><CardContent className="space-y-4 p-5">{[0, 1, 2, 3].map((row) => <Skeleton key={row} className="h-8 w-full" />)}</CardContent></Card>}
      {books.isError && <Alert variant="destructive"><AlertTitle>Runbooks unavailable</AlertTitle><AlertDescription>{(books.error as Error).message}</AlertDescription></Alert>}
      {books.data?.length ? <Card className="overflow-hidden border-border/80 bg-card/80 py-0"><CardContent className="overflow-x-auto p-0"><Table>
        <TableHeader><TableRow><TableHead>Runbook</TableHead><TableHead>Status</TableHead><TableHead>Last updated</TableHead><TableHead><span className="sr-only">Actions</span></TableHead></TableRow></TableHeader>
        <TableBody>{books.data.map((book) => <TableRow key={book.id}>
          <TableCell className="font-medium"><Link className="hover:text-primary" to={`/runbooks/${book.id}${folderID ? `?folder=${encodeURIComponent(folderID)}` : ''}`}>{book.title}</Link></TableCell>
          <TableCell><StatusBadge status={book.status} /></TableCell>
          <TableCell className="whitespace-nowrap text-muted-foreground">{relativeTime(book.updated_at)}</TableCell>
          <TableCell><Button type="button" variant="ghost" size="sm" onClick={() => { setMoveBook(book); setDestination(book.folder_id ?? '') }}><FolderInput />Move</Button></TableCell>
        </TableRow>)}</TableBody>
      </Table></CardContent></Card> : null}
      {books.data && !books.data.length && <Card className="border-border/80 bg-card/80"><CardContent className="flex flex-col items-center px-6 py-14 text-center">
        <Terminal className="size-7 text-primary" /><h2 className="mt-4 text-lg font-semibold">{status === 'all' ? 'A calm page is a good start.' : `No ${status} runbooks.`}</h2><p className="mt-1 max-w-md text-sm text-muted-foreground">{status === 'all' ? 'Capture the first repeatable fix and share it with your team.' : 'Try another status filter or create a new runbook.'}</p>
        {status === 'all' && <Button className="mt-5" render={<Link to={`/runbooks/new${folderID ? `?folder=${encodeURIComponent(folderID)}` : ''}`} />}><Plus />Write your first runbook</Button>}
      </CardContent></Card>}
    </section>
    <Dialog open={Boolean(moveBook)} onOpenChange={(open) => !open && setMoveBook(null)}><DialogContent><DialogHeader><DialogTitle>Move runbook</DialogTitle><DialogDescription>Choose a folder or move this runbook to the root.</DialogDescription></DialogHeader><div className="grid gap-2"><label htmlFor="runbook-folder" className="text-sm font-medium">Destination</label><select id="runbook-folder" value={destination} onChange={(event) => setDestination(event.target.value)} className="h-9 rounded-md border border-input bg-background px-3 text-sm"><option value="">Root</option>{(folders.data ?? []).map((item) => <option key={item.id} value={item.id}>{folderPath(folders.data ?? [], item.id)}</option>)}</select></div><DialogFooter><DialogClose render={<Button type="button" variant="outline">Cancel</Button>} /><Button type="button" disabled={move.isPending} onClick={() => move.mutate()}>{move.isPending ? 'Moving…' : 'Move runbook'}</Button></DialogFooter></DialogContent></Dialog>
  </main>
}
