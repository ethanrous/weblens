import { defineStore } from 'pinia'
import { Task, type TaskType, type TaskParams } from '~/types/task'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type TaskPromiseParams<T = any> = {
    resolve: (ti: T) => void
    taskID: string
}

export const useTasksStore = defineStore('tasks', () => {
    const tasks = shallowRef<Map<string, Task<TaskType>>>()
    const taskPromises = shallowRef<Map<string, TaskPromiseParams>>(new Map())

    const anyRunning = computed(() => {
        if (!tasks.value) return false

        return Array.from(tasks.value.values()).some((task) => task.isRunning)
    })

    function setTaskPromise<T>(params: TaskPromiseParams<T>) {
        taskPromises.value.set(params.taskID, params)
    }

    function removeTaskPromise(taskID: string) {
        taskPromises.value.delete(taskID)
    }

    function upsertTask(taskID: string, params: TaskParams) {
        if (!tasks.value) {
            tasks.value = new Map()
        }

        let task: Task
        if (!tasks.value.has(taskID)) {
            task = new Task(params)
            tasks.value.set(taskID, task)
        } else {
            task = tasks.value.get(taskID)!
            task.updateProgress(params)
        }

        // Trigger reactivity
        triggerRef(tasks)
    }

    function setTaskComplete<T>(taskID: string, content: T) {
        if (!tasks.value || !tasks.value.has(taskID)) return

        const task = tasks.value.get(taskID)!
        task.setComplete()

        const taskProm = taskPromises.value.get(taskID)
        if (taskProm) {
            taskProm.resolve(content)

            // Remove the promise from the map
            taskPromises.value.delete(taskID)
        }

        // Trigger reactivity
        triggerRef(tasks)
    }

    function cancelTask(taskID: string) {
        if (!tasks.value || !tasks.value.has(taskID)) return

        const task = tasks.value.get(taskID)!
        task.setCanceled()

        // Trigger reactivity
        triggerRef(tasks)
    }

    function failTask(taskID: string, opts?: { tasksFailed?: number }) {
        if (!tasks.value || !tasks.value.has(taskID)) {
            console.warn('Tried to fail a task that does not exist:', taskID)
            return
        }

        const task = tasks.value.get(taskID)!
        task.setFailed(opts)

        // Trigger reactivity
        triggerRef(tasks)
    }

    function removeTask(taskID: string) {
        if (!tasks.value) return

        tasks.value.delete(taskID)

        // Trigger reactivity
        triggerRef(tasks)
    }

    return {
        tasks,
        anyRunning,
        taskPromises,

        upsertTask,
        setTaskComplete,
        removeTask,
        cancelTask,
        failTask,

        setTaskPromise,
        removeTaskPromise,
    }
})
