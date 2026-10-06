import type { ReactNode } from 'react'

type PageHeaderProps = {
  eyebrow: string
  title: string
  description?: string
  action?: ReactNode
}

export function PageHeader({ eyebrow, title, description, action }: PageHeaderProps) {
  return (
    <div className="mb-7 flex flex-col items-start justify-between gap-5 sm:flex-row sm:items-end">
      <div className="min-w-0">
        <p className="mb-2 text-[11px] font-bold uppercase tracking-[.16em] text-primary">{eyebrow}</p>
        <h1 className="text-3xl font-semibold tracking-tight sm:text-4xl">{title}</h1>
        {description && <p className="mt-2 text-sm text-muted-foreground">{description}</p>}
      </div>
      {action}
    </div>
  )
}
