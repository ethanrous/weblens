export function friendlyActionName(action: string) {
    switch (action) {
        case 'fileCreate':
            return 'Created'
        case 'fileMove':
            return 'Moved'
        case 'fileDelete':
            return 'Deleted'
        default:
            return action
    }
}
