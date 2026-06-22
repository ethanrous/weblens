<template>
    <div
        :class="{ 'bg-background-primary text-text-tertiary relative h-6 border-b text-xs select-none': true }"
        :style="{ width: widthPx + 'px' }"
    >
        <div
            v-for="tick in ticks"
            :key="tick"
            :class="{ 'absolute top-0 bottom-0 border-l pl-1 whitespace-nowrap': true }"
            :style="{ left: barLeftPx(tick, domain, pxPerMs) + 'px' }"
        >
            {{ formatClock(tick) }}
        </div>
    </div>
</template>

<script setup lang="ts">
import type { TimeDomain } from '~/util/gantt'
import { axisTicks, barLeftPx, chooseTickStepMs, formatClock } from '~/util/gantt'

const { domain, pxPerMs, widthPx } = defineProps<{
    domain: TimeDomain
    pxPerMs: number
    widthPx: number
}>()

const ticks = computed(() => axisTicks(domain, chooseTickStepMs(pxPerMs)))
</script>
