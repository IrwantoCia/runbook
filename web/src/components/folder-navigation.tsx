import { useEffect, useRef, useState, type ReactNode } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ChevronDown, ChevronRight, Folder as FolderIcon, FolderPlus, Pencil, Trash2, MoveRight } from 'lucide-react'
import { toast } from 'sonner'
import { folderRequests, type Folder } from '../api'
import { folderPath } from '../lib/folder-path'
import { queryKeys } from '../lib/query-keys'
import { useMediaQuery } from '../lib/use-media-query'
import { Button } from './ui/button'
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from './ui/dialog'
import { Input } from './ui/input'
import { Label } from './ui/label'

type FolderNavigationProps = { selected: string | null; onSelect: (id: string | null) => void }
type FolderDialog = { kind: 'create' | 'rename' | 'move' | 'delete'; folder?: Folder } | null

export function FolderNavigation({ selected, onSelect }: FolderNavigationProps) {
  const isDesktop = useMediaQuery('(min-width: 1024px)')
  const client = useQueryClient()
  const folders = useQuery({ queryKey: queryKeys.folders, queryFn: folderRequests.list })
  const [dialog, setDialog] = useState<FolderDialog>(null)
  const [name, setName] = useState('')
  const [parentID, setParentID] = useState('')
  const [showFolders, setShowFolders] = useState(() => isDesktop)
  const [expandedFolders, setExpandedFolders] = useState<Set<string>>(new Set())
  const items = folders.data ?? []
  const didInit = useRef(false)
  useEffect(() => {
    if (!isDesktop || !folders.isSuccess || didInit.current) return
    setExpandedFolders(new Set(items.map((item) => item.id)))
    didInit.current = true
  }, [folders.isSuccess, isDesktop, items])
  const descendants = (id: string): Set<string> => {
    const result = new Set<string>()
    const visit = (parent: string) => items.filter((item) => item.parent_id === parent).forEach((item) => { result.add(item.id); visit(item.id) })
    visit(id)
    return result
  }
  const refresh = async (message: string) => {
    await client.invalidateQueries({ queryKey: queryKeys.folders })
    await client.invalidateQueries({ queryKey: queryKeys.books })
    if (dialog?.kind === 'delete' && dialog.folder?.id === selected) onSelect(null)
    toast.success(message)
    setDialog(null)
  }
  const mutation = useMutation({
    mutationFn: async (): Promise<void> => {
      if (!dialog) throw new Error('Choose a folder action')
      if (dialog.kind === 'create') await folderRequests.create(name.trim(), selected)
      else if (dialog.kind === 'rename' && dialog.folder) await folderRequests.rename(dialog.folder.id, name.trim())
      else if (dialog.kind === 'move' && dialog.folder) await folderRequests.move(dialog.folder.id, parentID || null)
      else if (dialog.kind === 'delete' && dialog.folder) await folderRequests.remove(dialog.folder.id)
      else throw new Error('Choose a folder action')
    },
    onSuccess: () => refresh(dialog?.kind === 'create' ? 'Folder created' : dialog?.kind === 'delete' ? 'Folder deleted' : dialog?.kind === 'move' ? 'Folder moved' : 'Folder renamed'),
    onError: (error: Error) => {
      const message = error.message
      if (message.includes('into itself')) {
        toast.error('Cannot move a folder into itself or one of its subfolders.')
      } else if (message.includes('still contains')) {
        toast.error(dialog?.kind === 'delete'
          ? 'Cannot delete: this folder still has runbooks or subfolders. Move them out first.'
          : 'Cannot move: this folder still contains items.')
      } else if (message.includes('already exists')) {
        toast.error('A folder with this name already exists here.')
      } else {
        toast.error(error.message)
      }
    },
  })
  const open = (next: FolderDialog) => {
    setDialog(next)
    setName(next?.kind === 'rename' ? next.folder?.name ?? '' : '')
    setParentID(next?.kind === 'move' ? next.folder?.parent_id ?? '' : '')
  }
  const renderBranch = (parentID: string | null, depth = 0): ReactNode => items.filter((folder) => folder.parent_id === parentID).map((folder) => {
    const children = items.some((child) => child.parent_id === folder.id)
    return <div key={folder.id}>
      <div className={`group flex min-w-0 items-center gap-1 rounded-md pr-1 ${selected === folder.id ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-muted/70 hover:text-foreground'}`} style={{ paddingLeft: `${depth * 14 + 4}px` }}>
              <button type="button" aria-label={`${expandedFolders.has(folder.id) ? 'Collapse' : 'Expand'} ${folder.name}`} disabled={!children} onClick={() => setExpandedFolders((current) => { const next = new Set(current); if (next.has(folder.id)) next.delete(folder.id); else next.add(folder.id); return next })} className="grid size-6 shrink-0 cursor-pointer place-items-center rounded disabled:cursor-default disabled:opacity-40">{children ? expandedFolders.has(folder.id) ? <ChevronDown className="size-3.5" /> : <ChevronRight className="size-3.5" /> : null}</button>
        <button type="button" onClick={() => onSelect(folder.id)} className="flex min-w-0 flex-1 cursor-pointer items-center gap-2 py-2 text-left text-sm"><FolderIcon className="size-4 shrink-0" /><span className="truncate">{folder.name}</span></button>
        <div className="flex shrink-0 opacity-100 sm:opacity-0 sm:group-hover:opacity-100 sm:group-focus-within:opacity-100">
          <Button type="button" aria-label={`Folder actions for ${folder.name}`} variant="ghost" size="icon-xs" onClick={() => open({ kind: 'rename', folder })}><Pencil /></Button>
          <Button type="button" aria-label={`Move ${folder.name}`} variant="ghost" size="icon-xs" onClick={() => open({ kind: 'move', folder })}><MoveRight /></Button>
          <Button type="button" aria-label={`Delete ${folder.name}`} variant="ghost" size="icon-xs" onClick={() => open({ kind: 'delete', folder })}><Trash2 /></Button>
        </div>
      </div>
      {expandedFolders.has(folder.id) && children && renderBranch(folder.id, depth + 1)}
    </div>
  })

  return <>
    <section className="rounded-xl border border-border/80 bg-card/50 p-3" aria-label="Runbook folders">
      <div className="flex items-center justify-between gap-2"><button type="button" className="flex cursor-pointer items-center gap-2 text-sm font-semibold" onClick={() => setShowFolders(!showFolders)}><span className="grid size-7 place-items-center rounded-md bg-primary/10 text-primary"><FolderIcon className="size-4" /></span>Folders <span className="text-xs font-normal text-muted-foreground">{showFolders ? 'Hide' : 'Browse'}</span></button><Button type="button" variant="outline" size="sm" onClick={() => open({ kind: 'create' })}><FolderPlus />New</Button></div>
      {showFolders && <div className="mt-3 space-y-1 border-t border-border pt-2">
        <button type="button" onClick={() => onSelect(null)} className={`flex w-full cursor-pointer items-center gap-2 rounded-md px-2 py-2 text-left text-sm ${selected === null ? 'bg-primary/10 font-medium text-primary' : 'text-muted-foreground hover:bg-muted/70 hover:text-foreground'}`}><FolderIcon className="size-4" />All runbooks</button>
        {folders.isError && <p className="px-2 py-2 text-sm text-destructive">{(folders.error as Error).message}</p>}
        {renderBranch(null)}
      </div>}
    </section>
    <Dialog open={Boolean(dialog)} onOpenChange={(openState) => !openState && setDialog(null)}>
      <DialogContent>
        <DialogHeader><DialogTitle>{dialog?.kind === 'create' ? 'Create folder' : dialog?.kind === 'rename' ? 'Rename folder' : dialog?.kind === 'move' ? 'Move folder' : 'Delete folder?'}</DialogTitle><DialogDescription>{dialog?.kind === 'delete' ? 'Only empty folders can be deleted. Runbooks and subfolders must be moved first.' : dialog?.kind === 'move' ? 'Choose a new parent. Moving into this folder or one of its descendants is not allowed.' : 'Keep folders organized with clear names.'}</DialogDescription></DialogHeader>
        {dialog?.kind === 'create' || dialog?.kind === 'rename' ? <div className="grid gap-2"><Label htmlFor="folder-name">Folder name</Label><Input id="folder-name" autoFocus maxLength={120} value={name} onChange={(event) => setName(event.target.value)} onKeyDown={(event) => event.key === 'Enter' && name.trim() && mutation.mutate()} /></div> : null}
        {dialog?.kind === 'move' && dialog.folder && <div className="grid gap-2"><Label htmlFor="folder-parent">Parent folder</Label><select id="folder-parent" value={parentID} onChange={(event) => setParentID(event.target.value)} className="h-9 rounded-md border border-input bg-background px-3 text-sm"><option value="">Root</option>{items.filter((item) => item.id !== dialog.folder?.id && !descendants(dialog.folder!.id).has(item.id)).map((item) => <option key={item.id} value={item.id}>{folderPath(items, item.id)}</option>)}</select></div>}
        <DialogFooter><DialogClose render={<Button type="button" variant="outline">Cancel</Button>} /><Button type="button" variant={dialog?.kind === 'delete' ? 'destructive' : 'default'} disabled={mutation.isPending || ((dialog?.kind === 'create' || dialog?.kind === 'rename') && !name.trim())} onClick={() => mutation.mutate()}>{mutation.isPending ? 'Saving…' : dialog?.kind === 'delete' ? 'Delete folder' : dialog?.kind === 'move' ? 'Move folder' : dialog?.kind === 'rename' ? 'Save name' : 'Create folder'}</Button></DialogFooter>
      </DialogContent>
    </Dialog>
  </>
}
