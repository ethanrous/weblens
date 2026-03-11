export function friendlyActionName(action: string) {
    switch (action) {
        case 'fileCreate':
            return 'Created'
        case 'fileMove':
            return 'Moved'
        case 'fileDelete':
            return 'Deleted'
        case 'fileRestore':
            return 'Restored'
        default:
            return action
    }
}
