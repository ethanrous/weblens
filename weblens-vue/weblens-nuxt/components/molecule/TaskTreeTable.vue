<template>
    <div :class="{ 'flex shrink-0 overflow-y-auto rounded border': true }">
        <table
            :class="{
                'w-full caption-bottom border-separate border-spacing-0 text-sm': true,
            }"
        >
            <thead :class="{ 'bg-background-primary sticky top-0 z-50 [&_tr]:border-b': true }">
                <tr :class="{ 'border-b': true }">
                    <th
                        v-for="header in HEADERS"
                        :key="header"
                        :class="{
                            'text-text-secondary bg-background h-10 border-r border-b px-2 text-left align-middle font-medium whitespace-nowrap last:border-r-0': true,
                        }"
                    >
                        {{ header }}
                    </th>
                </tr>
            </thead>
            <tbody :class="{ '[&_tr:last-child]:border-0': true }">
                <tr v-if="rows.length === 0">
                    <td
                        :colspan="HEADERS.length"
                        :class="{ 'text-text-tertiary p-4 text-center font-medium': true }"
                    >
                        {{ emptyText }}
                    </td>
                </tr>
                <tr
                    v-for="row in rows"
                    :key="row.task.taskID"
                    :class="{
                        'hover:bg-muted/50 even:bg-accent/25 border-b transition-colors': true,
                    }"
                >
                    <td :class="{ 'border-r p-2 text-left align-middle': true }">
                        <div
                            :class="{ 'flex flex-col gap-1': true }"
                            :style="{ paddingLeft: row.depth * 1.25 + 'rem' }"
                        >
                            <div :class="{ 'flex items-center gap-1': true }">
                                <span
                                    v-if="row.depth > 0"
                                    :class="{ 'text-text-tertiary select-none': true }"
                                >
                                    └
                                </span>
                                <WeblensButton
                                    v-if="row.hasChildren"
                                    type="light"
                                    :square-size="28"
                                    :class="{ 'mr-1': true }"
                                    @click="toggle(row.task.taskID)"
                                >
                                    <IconChevronRight
                                        :size="20"
                                        :class="{
                                            'transition-transform': true,
                                            'rotate-90': expanded.has(row.task.taskID),
                                        }"
                                    />
                                </WeblensButton>
                                <span
                                    v-else
                                    :class="{ 'inline-block w-7 shrink-0': true }"
                                />
                                <span :class="{ truncate: true }">{{ row.task.jobName }}</span>
                            </div>
                            <div
                                v-if="row.hasChildren"
                                :class="{ 'flex items-center gap-2 pl-5': true }"
                            >
                                <ProgressSquare
                                    :class="{ 'h-1.5 w-32': true }"
                                    :progress="row.percent"
                                />
                                <span :class="{ 'text-text-secondary text-xs whitespace-nowrap': true }">
                                    {{ row.completedCount }}/{{ row.totalCount }} · {{ row.percent }}%
                                </span>
                            </div>
                        </div>
                    </td>
                    <td :class="{ 'text-text-tertiary border-r p-2 text-center align-middle font-mono text-xs': true }">
                        <span
                            :class="{ 'block max-w-40 truncate': true }"
                            :title="row.task.taskID"
                        >
                            {{ row.task.taskID }}
                        </span>
                    </td>
                    <td :class="{ 'border-r p-2 text-center align-middle whitespace-nowrap': true }">
                        {{ row.task.State }}
                    </td>
                    <td :class="{ 'border-r p-2 text-center align-middle whitespace-nowrap': true }">
                        {{ row.task.workerID }}
                    </td>
                    <td :class="{ 'border-r p-2 text-center align-middle whitespace-nowrap': true }">
                        {{ formatStartTime(row.task.startTime) }}
                    </td>
                    <td
                        :class="{
                            'text-text-secondary border-r p-2 text-left align-middle font-mono text-xs break-all': true,
                        }"
                    >
                        {{ formatResult(row.task.result) }}
                    </td>
                    <td :class="{ 'p-2 text-center align-middle': true }">
                        <WeblensButton
                            label="Cancel"
                            flavor="danger"
                            :on-click="() => emit('cancel', row.task.taskID)"
                        />
                    </td>
                </tr>
            </tbody>
        </table>
    </div>
</template>

<script setup lang="ts">
import type { TaskInfo } from '@ethanrous/weblens-api'
import { IconChevronRight } from '@tabler/icons-vue'
import ProgressSquare from '~/components/atom/ProgressSquare.vue'
import WeblensButton from '~/components/atom/WeblensButton.vue'
import { buildTaskTree, flattenVisible } from '~/util/tasks'

const { tasks, emptyText = '' } = defineProps<{
    tasks: TaskInfo[]
    emptyText?: string
}>()

const emit = defineEmits<{
    cancel: [taskID: string]
}>()

const HEADERS = ['Job Name', 'Task ID', 'Status', 'Worker', 'Start Time', 'Result', 'Cancel']

const expanded = ref<Set<string>>(new Set())

const rows = computed(() => flattenVisible(buildTaskTree(tasks), expanded.value))

function toggle(taskID: string) {
    const next = new Set(expanded.value)
    if (next.has(taskID)) {
        next.delete(taskID)
    } else {
        next.add(taskID)
    }
    expanded.value = next
}

function formatStartTime(startTime?: string): string {
    if (!startTime) {
        return ''
    }
    const ms = Date.parse(startTime)
    if (isNaN(ms) || ms <= 0) {
        return ''
    }
    return new Date(ms).toLocaleTimeString()
}

function formatResult(result?: object): string {
    if (result === undefined || result === null) {
        return ''
    }
    try {
        return JSON.stringify(result)
    } catch {
        return String(result)
    }
}
</script>
