export function startupWebsocketHandler(setSetupProgress, setSetupMostRecent) {
    return (msgData) => {
        switch (msgData.eventTag) {
            case 'weblens_loaded': {
                setTimeout(() => location.reload(), 500)
                break
            }
            case 'task_complete': {
                setSetupProgress(
                    (msgData.content.tasksComplete /
                        msgData.content.tasksTotal) *
                        100
                )
                setSetupMostRecent(
                    `${msgData.taskType}: ${msgData.content.fileName}`
                )
                break
            }
            default: {
                setSetupMostRecent(`${msgData.taskType}`)
            }
        }
    }
}
