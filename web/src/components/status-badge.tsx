import type { Runbook } from '../api'
import { Badge } from './ui/badge'

export function StatusBadge({ status }: { status: Runbook['status'] }) {
  const styles = {
    draft: { variant: 'secondary' as const, className: '' },
    published: { variant: 'default' as const, className: 'bg-primary/15 text-primary' },
  }[status]

  return <Badge variant={styles.variant} className={styles.className}>{status}</Badge>
}
