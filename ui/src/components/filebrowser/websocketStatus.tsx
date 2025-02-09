import WeblensTooltip from '@weblens/lib/WeblensTooltip'

function WebsocketStatus({ ready }: { ready: number }) {
    let color: string
    let status: string

    switch (ready) {
        case 1:
            color = '#00ff0055'
            status = 'Connected'
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

    return (
        <WeblensTooltip label={status}>
            <svg width="24" height="24" fill={color}>
                <path d="M12 12m-9 0a9 9 0 1 0 18 0a9 9 0 1 0 -18 0" />
            </svg>
        </WeblensTooltip>
    )
}

export default WebsocketStatus
