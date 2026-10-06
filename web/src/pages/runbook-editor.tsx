import { useState } from 'react'
import { Link, useNavigate, useParams, useSearchParams } from 'react-router'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeSanitize from 'rehype-sanitize'
import { ArrowLeft, Check, FolderInput, Save, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { request, folderRequests, type Folder, type Runbook } from '../api'
import { queryKeys } from '../lib/query-keys'
import { folderPath } from '../lib/folder-path'
import { relativeTime } from '../lib/time'
import { useDocumentTitle } from '../lib/use-document-title'
import { Alert, AlertDescription, AlertTitle } from '../components/ui/alert'
import { Button } from '../components/ui/button'
import { Card, CardContent } from '../components/ui/card'
import { PageHeader } from '../components/page-header'
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '../components/ui/dialog'
import { Input } from '../components/ui/input'
import { Label } from '../components/ui/label'
import { Skeleton } from '../components/ui/skeleton'
import { StatusBadge } from '../components/status-badge'
import { Textarea } from '../components/ui/textarea'

type Wrapped<T, K extends string> = Record<K, T>

export function RunbookEditorPage() {
  const { id } = useParams()
  const editing = Boolean(id)
  const [searchParams] = useSearchParams()
  const [createFolder, setCreateFolder] = useState(() => searchParams.get('folder') ?? '')
  const client = useQueryClient()
  const navigate = useNavigate()
  const [preview, setPreview] = useState(false)
  const [moveOpen, setMoveOpen] = useState(false)
  const [destination, setDestination] = useState('')
  const folders = useQuery({ queryKey: queryKeys.folders, queryFn: folderRequests.list })
  const detail = useQuery({ queryKey: ['runbook', id], enabled: editing, queryFn: () => request<Wrapped<Runbook, 'runbook'>>(`/api/runbooks/${id}`).then((data) => data.runbook) })
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const [loadedID, setLoadedID] = useState('')
  const book = detail.data
  useDocumentTitle(editing ? 'Edit runbook · Runbook' : 'New runbook · Runbook')
  if (detail.data && loadedID !== detail.data.id) {
    setTitle(detail.data.title)
    setBody(detail.data.body)
    setLoadedID(detail.data.id)
  }

  const refresh = async (book: Runbook, message: string) => {
    client.setQueryData(['runbook', book.id], book)
    await client.invalidateQueries({ queryKey: queryKeys.books })
    toast.success(message)
  }
  const save = useMutation<Wrapped<Runbook, 'runbook'>, Error, { returnToList: boolean }>({
    mutationFn: () => editing
      ? request<Wrapped<Runbook, 'runbook'>>(`/api/runbooks/${id}`, { method: 'PATCH', body: JSON.stringify({ title, body }) })
      : request<Wrapped<Runbook, 'runbook'>>('/api/runbooks', { method: 'POST', body: JSON.stringify({ title, body, folder_id: createFolder || null }) }),
    onSuccess: async (data, variables) => {
      client.setQueryData(['runbook', data.runbook.id], data.runbook)
      await client.invalidateQueries({ queryKey: queryKeys.books })
      if (variables.returnToList) {
        toast.success(editing ? 'Runbook saved' : 'Runbook created')
        navigate(data.runbook.folder_id ? `/?folder=${encodeURIComponent(data.runbook.folder_id)}` : '/', { replace: !editing })
      }
    },
    onError: (error: Error) => toast.error(error.message),
  })
  const transition = useMutation({
    mutationFn: async () => {
      if (!editing || !id) throw new Error('Save this runbook before publishing')
      if (title.trim() !== book?.title || body !== book?.body) await save.mutateAsync({ returnToList: false })
      return request<Wrapped<Runbook, 'runbook'>>(`/api/runbooks/${id}/publish`, { method: 'POST' })
    },
    onSuccess: async (data) => refresh(data.runbook, 'Runbook published'),
    onError: (error: Error) => toast.error(error.message),
  })
  const move = useMutation({
    mutationFn: async () => {
      if (!editing || !id) throw new Error('Save this runbook before moving')
      if (title.trim() !== book?.title || body !== book?.body) await save.mutateAsync({ returnToList: false })
      return folderRequests.moveRunbook(id, destination || null)
    },
    onSuccess: async ({ runbook: moved }) => {
      await client.invalidateQueries({ queryKey: queryKeys.books })
      client.setQueryData(['runbook', moved.id], moved)
      toast.success('Runbook moved')
      setMoveOpen(false)
      navigate(`/${moved.folder_id ? `?folder=${encodeURIComponent(moved.folder_id)}` : ''}`)
    },
    onError: (error: Error) => toast.error(error.message),
  })
  const remove = useMutation({
    mutationFn: () => request(`/api/runbooks/${id}`, { method: 'DELETE' }),
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: queryKeys.books })
      toast.success('Runbook deleted')
      navigate('/')
    },
    onError: (error: Error) => toast.error(error.message),
  })

  const returnToFolder = book?.folder_id ?? searchParams.get('folder')
  const listPath = `/${returnToFolder ? `?folder=${encodeURIComponent(returnToFolder)}` : ''}`
  const publish = () => transition.mutate()

  if (editing && detail.isLoading) return <main className="space-y-5 py-10"><Skeleton className="h-8 w-48" /><Skeleton className="h-12 w-full" /><Skeleton className="h-80 w-full" /></main>
  if (editing && (!book || detail.isError)) return <main className="py-10"><p className="mb-4 text-sm text-muted-foreground">Runbook not found.</p><Button variant="outline" render={<Link to="/" />}><ArrowLeft />Back to runbooks</Button></main>

  const readOnly = book?.status === 'published'
  return (
    <main className="flex min-h-0 flex-1 flex-col gap-5 px-4 py-8 sm:px-7 sm:py-11">
      <Link to={listPath} className="inline-block text-sm text-muted-foreground hover:text-foreground">← Runbooks</Link>
      <PageHeader eyebrow={editing ? `/${book?.slug}` : 'New operational note'} title={editing ? 'Edit runbook' : 'New runbook'} description={book ? `Updated ${relativeTime(book.updated_at)}` : undefined} action={book ? <StatusBadge status={book.status} /> : undefined} />
      {readOnly && <Alert className="border-primary/30 bg-primary/5"><AlertTitle>Published runbook</AlertTitle><AlertDescription>Published runbooks are read-only.</AlertDescription></Alert>}
      <Card className="min-h-0 flex-1 border-border/80 bg-card/80"><CardContent className="flex min-h-0 flex-1 flex-col gap-5 p-4 sm:p-6">
        <div className="grid shrink-0 gap-2"><Label htmlFor="title">Title</Label><Input id="title" value={title} onChange={(event) => setTitle(event.target.value)} placeholder="e.g. Recover a stuck deployment" maxLength={120} readOnly={readOnly} /></div>
        {!editing && <div className="grid shrink-0 gap-2"><Label htmlFor="create-folder">Folder</Label><select id="create-folder" value={createFolder} onChange={(event) => setCreateFolder(event.target.value)} className="h-9 rounded-md border border-input bg-background px-3 text-sm"><option value="">Root</option>{(folders.data ?? []).map((folder: Folder) => <option key={folder.id} value={folder.id}>{folderPath(folders.data ?? [], folder.id)}</option>)}</select></div>}
        {!readOnly && <div className="flex w-fit shrink-0 gap-1 rounded-lg border border-border bg-background/70 p-1 lg:hidden" role="group" aria-label="Editor view">
          <Button type="button" size="sm" variant={!preview ? 'secondary' : 'ghost'} aria-pressed={!preview} onClick={() => setPreview(false)}>Write</Button>
          <Button type="button" size="sm" variant={preview ? 'secondary' : 'ghost'} aria-pressed={preview} onClick={() => setPreview(true)}>Preview</Button>
        </div>}
        <div className="grid min-h-0 flex-1 gap-5 lg:grid-cols-2">
          <div className={`flex min-h-0 flex-col gap-2 ${preview && !readOnly ? 'hidden lg:flex' : ''}`}><Label htmlFor="body">Markdown instructions</Label><Textarea id="body" className="h-full min-h-[320px] field-sizing-fixed font-mono text-sm leading-6" value={body} onChange={(event) => setBody(event.target.value)} placeholder={'# Symptoms\nDescribe what someone will notice.\n\n## Recovery\n1. Check the service…'} readOnly={readOnly} /></div>
          <div className={`flex min-h-0 flex-col ${preview || readOnly ? '' : 'hidden lg:flex'}`}><p className="mb-2 shrink-0 text-xs font-semibold uppercase tracking-[.12em] text-muted-foreground">Sanitized preview</p><article className="markdown-preview min-h-0 flex-1 overflow-auto rounded-lg border border-input bg-background/70 p-4 text-sm leading-7"><div className="max-w-[72ch]"><ReactMarkdown remarkPlugins={[remarkGfm]} rehypePlugins={[rehypeSanitize]}>{body || 'Your rendered runbook will appear here.'}</ReactMarkdown></div></article></div>
        </div>
        <div className="flex shrink-0 flex-wrap items-center gap-2 border-t border-border pt-4">
          {!readOnly && <Button disabled={save.isPending || !title.trim()} onClick={() => save.mutate({ returnToList: true })}><Save />{save.isPending ? 'Saving…' : editing ? 'Save changes' : 'Create draft'}</Button>}
          {book?.status === 'draft' && <Button disabled={save.isPending || transition.isPending} onClick={publish}><Check />{transition.isPending ? 'Publishing…' : 'Publish'}</Button>}
          {editing && <Button type="button" variant="outline" onClick={() => { setDestination(book?.folder_id ?? ''); setMoveOpen(true) }}><FolderInput />Move</Button>}
      <Dialog open={moveOpen} onOpenChange={setMoveOpen}><DialogContent><DialogHeader><DialogTitle>Move runbook</DialogTitle><DialogDescription>Choose a folder or move this runbook to the root.</DialogDescription></DialogHeader><div className="grid gap-2"><Label htmlFor="editor-folder">Destination</Label><select id="editor-folder" value={destination} onChange={(event) => setDestination(event.target.value)} className="h-9 rounded-md border border-input bg-background px-3 text-sm"><option value="">Root</option>{(folders.data ?? []).map((folder: Folder) => <option key={folder.id} value={folder.id}>{folderPath(folders.data ?? [], folder.id)}</option>)}</select></div><DialogFooter><DialogClose render={<Button type="button" variant="outline">Cancel</Button>} /><Button type="button" disabled={move.isPending} onClick={() => move.mutate()}>{move.isPending ? 'Moving…' : 'Move runbook'}</Button></DialogFooter></DialogContent></Dialog>
      {editing && <Dialog><DialogTrigger render={<Button variant="destructive" className="sm:ml-auto" />}><Trash2 />Delete</DialogTrigger><DialogContent>
            <DialogHeader><DialogTitle>Delete this runbook?</DialogTitle><DialogDescription>This action is permanent. The runbook will be removed from your library.</DialogDescription></DialogHeader>
            <DialogFooter><DialogClose render={<Button variant="outline">Cancel</Button>} /><Button variant="destructive" disabled={remove.isPending} onClick={() => remove.mutate()}>{remove.isPending ? 'Deleting…' : 'Delete permanently'}</Button></DialogFooter>
          </DialogContent></Dialog>}
        </div>
      </CardContent></Card>
    </main>
  )
}
