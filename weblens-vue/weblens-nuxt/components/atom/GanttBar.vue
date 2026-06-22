<template>
    <div
        :class="{
            'border-abyss-900 absolute top-1 bottom-1 cursor-default overflow-hidden rounded-sm border px-1 text-xs leading-none text-white/90': true,
            'flex items-center': true,
            [colorClass]: true,
        }"
        :style="{ left: leftPx + 'px', width: widthPx + 'px' }"
        @mouseenter="(e) => emit('enter', bar, e)"
        @mousemove="(e) => emit('enter', bar, e)"
        @mouseleave="emit('leave')"
    >
        <span
            v-if="widthPx >= 28"
            :class="{ 'min-w-0 truncate': true }"
        >
            {{ bar.task.jobName }}
        </span>
    </div>
</template>

<script setup lang="ts">
import type { GanttBar } from '~/util/gantt'

const { bar, leftPx, widthPx, colorClass } = defineProps<{
    bar: GanttBar
    leftPx: number
    widthPx: number
    colorClass: string
}>()

const emit = defineEmits<{
    enter: [bar: GanttBar, event: MouseEvent]
    leave: []
}>()
</script>
