export function relativeTime(iso: string): string {
  const timestamp = new Date(iso).getTime()
  if (!Number.isFinite(timestamp)) return new Date(iso).toLocaleDateString()

  const seconds = Math.max(0, Math.floor((Date.now() - timestamp) / 1000))
  if (seconds < 60) return 'just now'
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`
  if (seconds < 86_400) return `${Math.floor(seconds / 3600)}h ago`
  if (seconds < 7 * 86_400) return `${Math.floor(seconds / 86_400)}d ago`
  return new Date(iso).toLocaleDateString()
}
