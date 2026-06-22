import type { TaskInfo } from '@ethanrous/weblens-api'
import { useIntervalFn } from '@vueuse/core'
import { useWeblensAPI } from '~/api/AllApi'
import { mergeTaskPoll, parseTimeMs } from '~/util/gantt'

// useTaskHistory polls the tower task endpoint (including exited tasks) and accumulates
// every observed task into a session-scoped map, so the gantt can show recent history
// even after the backend evicts a finished task from memory.
export function useTaskHistory(pollMs = 3000) {
    const history = ref<Map<string, TaskInfo>>(new Map())
    const error = ref<unknown>(null)

    // Server-domain cursor: the latest finish time seen. Each poll only asks for tasks that
    // finished after it, so we don't re-fetch the whole retained history every time.
    let finishCursorMs = 0

    async function refresh() {
        try {
            const res = await useWeblensAPI().TowersAPI.getRunningTasks(true, finishCursorMs)
            history.value = mergeTaskPoll(history.value, res.data, Date.now())

            for (const task of res.data) {
                const finishedMs = parseTimeMs(task.finishTime)
                if (finishedMs > finishCursorMs) {
                    finishCursorMs = finishedMs
                }
            }

            error.value = null
        } catch (e) {
            error.value = e
        }
    }

    const allTasks = computed(() => Array.from(history.value.values()))

    const runningTasks = computed(() =>
        allTasks.value
            .filter((t) => !t.Completed && t.State !== 'Exited')
            .sort((a, b) => parseTimeMs(b.startTime) - parseTimeMs(a.startTime)),
    )

    function clear() {
        history.value = new Map()
        finishCursorMs = 0
    }

    onMounted(() => {
        void refresh()
    })
    useIntervalFn(() => {
        void refresh()
    }, pollMs)

    return { history, allTasks, runningTasks, error, refresh, clear }
}
