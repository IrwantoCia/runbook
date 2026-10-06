export type User = { id: string; email: string }
export type Runbook = { id: string; slug: string; title: string; body: string; status: 'draft' | 'published'; created_by: string; created_at: string; updated_at: string; published_at?: string; folder_id?: string }
export type Folder = { id: string; name: string; parent_id: string | null }
export type APIKey = { id: string; name: string; prefix: string; created_at: string; last_used_at: string | null; revoked_at: string | null }

export async function request<T>(url: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(url, { ...init, credentials: 'same-origin', headers: { 'Content-Type': 'application/json', ...init.headers } })
  if (!response.ok) { const data = await response.json().catch(() => ({})); throw new Error(data.error ?? `Request failed (${response.status})`) }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

export const folderRequests = {
  list: () => request<{ folders: Folder[] }>('/api/folders').then((data) => data.folders),
  create: (name: string, parent_id: string | null) => request<{ folder: Folder }>('/api/folders', { method: 'POST', body: JSON.stringify({ name, parent_id }) }),
  rename: (id: string, name: string) => request<{ folder: Folder }>(`/api/folders/${id}`, { method: 'PATCH', body: JSON.stringify({ name }) }),
  move: (id: string, parent_id: string | null) => request<void>(`/api/folders/${id}/move`, { method: 'POST', body: JSON.stringify({ parent_id }) }),
  remove: (id: string) => request<void>(`/api/folders/${id}`, { method: 'DELETE' }),
  moveRunbook: (id: string, folder_id: string | null) => request<{ runbook: Runbook }>(`/api/runbooks/${id}/move`, { method: 'POST', body: JSON.stringify({ folder_id }) }),
}
