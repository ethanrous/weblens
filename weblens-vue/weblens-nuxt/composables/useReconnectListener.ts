import { WsAction, WsSubscriptionType } from '~/types/websocket'

const folderSubs = ref(new Map<string, string>())

const taskSubs = ref(new Set<string>())

export function useReconnectListener() {
    function addFolderSub(folderID: string, shareID: string) {
        if (folderID) {
            folderSubs.value.set(folderID, shareID)
        }
    }

    function removeFolderSub(folderID: string) {
        folderSubs.value.delete(folderID)
    }

    function addTaskSub(taskID: string) {
        if (taskID) {
            taskSubs.value.add(taskID)
        }
    }

    function removeTaskSub(taskID: string) {
        taskSubs.value.delete(taskID)
    }

    // Takes the websocket send function as a parameter (rather than importing
    // the store or FileBrowserApi) to avoid a module import cycle.
    function replay(send: (data: object) => void) {
        for (const [folderID, shareID] of folderSubs.value) {
            send({
                action: WsAction.Subscribe,
                subscriptionType: WsSubscriptionType.Folder,
                subscribeKey: folderID,
                content: { shareID: shareID },
            })
        }

        for (const taskID of taskSubs.value) {
            send({
                action: WsAction.Subscribe,
                subscriptionType: WsSubscriptionType.Task,
                subscribeKey: taskID,
                content: {},
            })
        }
    }

    return { addFolderSub, removeFolderSub, addTaskSub, removeTaskSub, replay }
}
