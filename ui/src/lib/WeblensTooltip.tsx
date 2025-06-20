import { ReactNode } from 'react'

export default function WeblensTooltip({
    label,
    children,
    className = '',
}: {
    label: ReactNode
    children: ReactNode
    className?: string
}) {
    return (
        <div className={className}>
            {children}
            <div className="border-b-color-border-primary bg-background-secondary text-color-text-primary pointer-events-none absolute z-50 mt-1 rounded-md border p-1 opacity-0 shadow-lg transition hover:block hover:opacity-100">
                {label}
            </div>
        </div>
    )
}
