import WeblensTooltip from '@weblens/lib/WeblensTooltip'

function getWsState(ready: number): { color: string; status: string } {
    let color: string
    let status: string

    switch (ready) {
        case 1:
            color = '#00ff0055'
            status = 'Online'
            break
        case 2:
        case 3:
            color = 'orange'
            status = 'Connecting'
            break
        case -1:
            color = 'red'
            status = 'Disconnected'
    }

    return { color, status }
}

function WebsocketStatusDot({ ready }: { ready: number }) {
    const { color, status } = getWsState(ready)

    return (
        <WeblensTooltip label={status}>
            <svg width="24" height="24" fill={color}>
                <path d="M12 12m-9 0a9 9 0 1 0 18 0a9 9 0 1 0 -18 0" />
            </svg>
        </WeblensTooltip>
    )
}

export function WebsocketStatusCard({ ready }: { ready: number }) {
    const { color, status } = getWsState(ready)

    return (
        <div
            className="flex items-center gap-2 p-1 rounded w-max"
            style={{ backgroundColor: color }}
        >
            <span className='text-xs'>{status}</span>
        </div>
    )
}

export default WebsocketStatusDot
