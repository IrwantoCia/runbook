import type { Folder } from '../api'

export function folderPath(folders: Folder[], id: string | null | undefined): string {
  if (!id) return ''

  const names: string[] = []
  const visited = new Set<string>()
  let currentID: string | null = id

  while (currentID && !visited.has(currentID)) {
    visited.add(currentID)
    const folder = folders.find((item) => item.id === currentID)
    if (!folder) break
    names.unshift(folder.name)
    currentID = folder.parent_id
  }

  return names.join(' / ')
}
