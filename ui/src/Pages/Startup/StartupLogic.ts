export function startupWebsocketHandler(setSetupProgress, setSetupMostRecent) {
    return msgData => {
        switch (msgData.eventTag) {
            case 'weblens_loaded': {
                location.reload();
            }
            case 'task_complete': {
                setSetupProgress((msgData.content.queue_remaining / msgData.content.queue_total) * 100);
                setSetupMostRecent(`${msgData.taskType}: ${msgData.content.task_id}`);
                break;
            }
            default: {
                setSetupMostRecent(`${msgData.taskType}`);
            }
        }
    };
}