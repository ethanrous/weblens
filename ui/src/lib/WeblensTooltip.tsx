import "./tooltipStyle.scss"

export default function WeblensTooltip({label, children}) {
    return (
        <div>
            {children}
            <div className="wl-tooltip">
                {label}
            </div>
        </div>
    )
}
