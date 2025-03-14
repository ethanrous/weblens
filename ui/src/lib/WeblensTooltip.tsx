export default function WeblensTooltip({ label, children }) {
    return (
        <div>
            {children}
            <div className="pointer-events-none absolute z-50 mt-1 rounded-md border border-b-color-border-primary bg-background-secondary p-1 text-color-text-primary opacity-0 shadow-lg transition hover:block hover:opacity-100">
                {label}
            </div>
        </div>
    )
}
