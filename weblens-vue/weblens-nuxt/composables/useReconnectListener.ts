import { SubToFolder, SubToTask } from '~/api/FileBrowserApi'

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

    function replay() {
        for (const [folderID, shareID] of folderSubs.value) {
            SubToFolder(folderID, shareID)
        }

        for (const taskID of taskSubs.value) {
            SubToTask(taskID)
        }
    }

    return { addFolderSub, removeFolderSub, addTaskSub, removeTaskSub, replay }
}
