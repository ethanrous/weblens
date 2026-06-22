<template>
    <Teleport to="body">
        <div
            v-if="task"
            :class="{
                'bg-background-primary pointer-events-none fixed z-100 flex max-w-100 min-w-72 flex-col gap-1 rounded border p-3 text-sm shadow-lg': true,
            }"
            :style="{ left: clampedX + 'px', top: clampedY + 'px' }"
        >
            <div :class="{ 'flex items-center justify-between gap-2': true }">
                <span :class="{ 'truncate font-medium': true }">{{ task.jobName }}</span>
                <span
                    :class="{
                        'rounded px-1.5 py-0.5 text-xs whitespace-nowrap text-white/90': true,
                        [colorClass]: true,
                    }"
                >
                    {{ task.State }}
                </span>
            </div>

            <div :class="{ 'grid grid-cols-[auto_1fr] gap-x-3 gap-y-0.5': true }">
                <span :class="{ 'text-text-tertiary': true }">Task ID</span>
                <span :class="{ 'font-mono text-xs break-all': true }">{{ task.taskID }}</span>

                <span :class="{ 'text-text-tertiary': true }">Status</span>
                <span>{{ task.status || '-' }}</span>

                <span :class="{ 'text-text-tertiary': true }">Worker</span>
                <span>{{ isQueuedTask ? 'unassigned' : task.workerID }}</span>

                <span :class="{ 'text-text-tertiary': true }">Start</span>
                <span>{{ formatTime(task.startTime) }}</span>

                <span :class="{ 'text-text-tertiary': true }">Finish</span>
                <span>{{ formatTime(task.finishTime) }}</span>

                <span :class="{ 'text-text-tertiary': true }">Duration</span>
                <span>{{ durationLabel }}</span>
            </div>

            <div
                v-if="task.error"
                :class="{ 'mt-1 flex flex-col gap-0.5': true }"
            >
                <span :class="{ 'text-danger text-xs font-medium': true }">Error</span>
                <pre
                    :class="{
                        'text-danger max-h-32 overflow-auto font-mono text-xs break-all whitespace-pre-wrap': true,
                    }"
                    >{{ task.error }}</pre
                >
            </div>

            <div
                v-if="(task.totalChildTasks ?? 0) > 0"
                :class="{ 'mt-1 flex items-center gap-2': true }"
            >
                <ProgressSquare
                    :class="{ 'h-1.5 w-32': true }"
                    :progress="childPercent"
                />
                <span :class="{ 'text-text-secondary text-xs whitespace-nowrap': true }">
                    {{ task.completedChildTasks ?? 0 }}/{{ task.totalChildTasks }}
                </span>
            </div>

            <div
                v-if="metaText"
                :class="{ 'mt-1 flex flex-col gap-0.5': true }"
            >
                <span :class="{ 'text-text-tertiary': true }">Metadata</span>
                <pre
                    :class="{
                        'text-text-secondary max-h-32 overflow-auto font-mono text-xs break-all whitespace-pre-wrap': true,
                    }"
                    >{{ metaText }}</pre
                >
            </div>

            <div
                v-if="resultText"
                :class="{ 'mt-1 flex flex-col gap-0.5': true }"
            >
                <span :class="{ 'text-text-tertiary': true }">Result</span>
                <pre
                    :class="{
                        'text-text-secondary max-h-32 overflow-auto font-mono text-xs break-all whitespace-pre-wrap': true,
                    }"
                >
                    {{ resultText }}
                    </pre
                >
            </div>
        </div>
    </Teleport>
</template>

<script setup lang="ts">
import type { TaskInfo } from '@ethanrous/weblens-api'
import ProgressSquare from '~/components/atom/ProgressSquare.vue'
import { isQueued, parseTimeMs, stateColorClass } from '~/util/gantt'
import { humanDuration } from '~/util/humanBytes'

const { task, now, x, y } = defineProps<{
    task: TaskInfo | null
    now: number
    x: number
    y: number
}>()

const TOOLTIP_W = 400
const TOOLTIP_H = 320
const OFFSET = 16

const clampedX = computed(() => {
    if (import.meta.client && x + TOOLTIP_W + OFFSET > window.innerWidth) {
        return Math.max(OFFSET, x - TOOLTIP_W - OFFSET)
    }
    return x + OFFSET
})

const clampedY = computed(() => {
    if (import.meta.client && y + TOOLTIP_H + OFFSET > window.innerHeight) {
        return Math.max(OFFSET, window.innerHeight - TOOLTIP_H - OFFSET)
    }
    return y + OFFSET
})

const colorClass = computed(() => (task ? stateColorClass(task) : 'bg-muted'))
const isQueuedTask = computed(() => (task ? isQueued(task) : false))

const childPercent = computed(() => {
    if (!task || !task.totalChildTasks) {
        return 0
    }
    return Math.round(((task.completedChildTasks ?? 0) / task.totalChildTasks) * 100)
})

const durationLabel = computed(() => {
    if (!task) {
        return '-'
    }
    const start = parseTimeMs(task.startTime)
    if (start === 0) {
        return '-'
    }
    const end = parseTimeMs(task.finishTime) || now
    return humanDuration(end - start)
})

const metaText = computed(() => {
    if (!task?.metadata || Object.keys(task.metadata).length === 0) {
        return ''
    }
    try {
        return JSON.stringify(task.metadata, null, 2)
    } catch {
        return String(task.metadata)
    }
})

const resultText = computed(() => {
    if (!task?.result || Object.keys(task.result).length === 0) {
        return ''
    }
    try {
        return JSON.stringify(task.result, null, 2)
    } catch {
        return String(task.result)
    }
})

function formatTime(t?: string): string {
    const ms = parseTimeMs(t)
    return ms === 0 ? '-' : new Date(ms).toLocaleTimeString()
}
</script>
