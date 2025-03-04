export default function WeblensTooltip({ label, children }) {
    return (
        <div>
            {children}
            <div className="pointer-events-none absolute z-50 mt-1 rounded-md border border-b-wl-border-color-primary bg-wl-background-color-secondary p-1 text-wl-text-color-primary opacity-0 shadow-lg transition hover:block hover:opacity-100">
                {label}
            </div>
        </div>
    )
}
