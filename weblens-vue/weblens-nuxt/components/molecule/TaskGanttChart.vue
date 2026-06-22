<template>
    <div
        v-bind="{ 'data-testid': 'task-gantt' }"
        :class="{ 'flex min-w-0 flex-col gap-2': true }"
    >
        <div :class="{ 'flex items-center justify-between gap-2': true }">
            <span :class="{ 'text-text-secondary text-xs': true }">
                {{ tasks.length }} task{{ tasks.length === 1 ? '' : 's' }} - {{ model.queuedCount }} queued
            </span>
            <div :class="{ 'flex items-center gap-1': true }">
                <WeblensButton
                    type="light"
                    :square-size="28"
                    :on-click="() => zoom(0.5)"
                >
                    <IconZoomOut :size="18" />
                </WeblensButton>
                <WeblensButton
                    type="light"
                    :square-size="28"
                    :on-click="() => zoom(2)"
                >
                    <IconZoomIn :size="18" />
                </WeblensButton>
            </div>
        </div>

        <div
            v-if="model.lanes.length === 0"
            :class="{ 'text-text-tertiary rounded border p-4 text-center font-medium': true }"
        >
            {{ emptyText }}
        </div>

        <div
            v-else
            :class="{ 'flex w-full min-w-0 overflow-hidden rounded border': true }"
        >
            <!-- Sticky lane-label gutter -->
            <div :class="{ 'bg-background-primary z-10 shrink-0 border-r': true }">
                <div :class="{ 'h-6 border-b': true }" />
                <div
                    v-for="lane in model.lanes"
                    :key="lane.key"
                    :class="{
                        'text-text-secondary flex h-9 items-center px-3 text-xs font-medium whitespace-nowrap': true,
                    }"
                >
                    {{ lane.label }}
                </div>
            </div>

            <!-- Scrollable timeline -->
            <div
                ref="scroller"
                :class="{ 'relative min-w-0 flex-1 overflow-x-auto': true }"
                @scroll="onScroll"
            >
                <GanttTimeAxis
                    :domain="model.domain"
                    :px-per-ms="pxPerMs"
                    :width-px="widthPx"
                />

                <div
                    v-for="lane in visibleLanes"
                    :key="lane.key"
                    :class="{ 'even:bg-background-secondary/40 relative h-9 border-b last:border-b-0': true }"
                    :style="{ width: widthPx + 'px' }"
                >
                    <GanttBar
                        v-for="bar in lane.bars"
                        :key="bar.task.taskID"
                        :bar="bar"
                        :left-px="barLeftPx(bar.startMs, model.domain, pxPerMs)"
                        :width-px="barWidthPx(bar.startMs, bar.endMs, pxPerMs)"
                        :color-class="stateColorClass(bar.task)"
                        @enter="onBarEnter"
                        @leave="hoveredId = null"
                    />
                </div>

                <!-- now line -->
                <div
                    :class="{ 'bg-danger/70 pointer-events-none absolute top-0 bottom-0 w-px': true }"
                    :style="{ left: nowLineLeft + 'px' }"
                />
            </div>
        </div>

        <TaskGanttTooltip
            :task="hoveredTask"
            :now="nowMs"
            :x="mouseX"
            :y="mouseY"
        />
    </div>
</template>

<script setup lang="ts">
import type { TaskInfo } from '@ethanrous/weblens-api'
import { useElementSize, useIntervalFn } from '@vueuse/core'
import { IconZoomIn, IconZoomOut } from '@tabler/icons-vue'
import GanttBar from '~/components/atom/GanttBar.vue'
import GanttTimeAxis from '~/components/atom/GanttTimeAxis.vue'
import WeblensButton from '~/components/atom/WeblensButton.vue'
import TaskGanttTooltip from '~/components/molecule/TaskGanttTooltip.vue'
import type { GanttBar as GanttBarType } from '~/util/gantt'
import { barLeftPx, barWidthPx, buildGantt, stateColorClass, totalWidthPx } from '~/util/gantt'

const { tasks, emptyText = 'No tasks observed yet' } = defineProps<{
    tasks: TaskInfo[]
    emptyText?: string
}>()

const MIN_PX_PER_MS = 0.0005 // ~1px / 2s
const MAX_PX_PER_MS = 0.2 // ~200px / s
const DEFAULT_PX_PER_MS = 0.05 // ~50px / s
const OVERSCAN_PX = 300 // render a little beyond the viewport so scrolling doesn't reveal gaps

const pxPerMs = ref(DEFAULT_PX_PER_MS)
const nowMs = ref(Date.now()) // live tick, drives the now-line only
const modelNow = ref(Date.now()) // refreshed per poll, drives the (potentially large) model
const hoveredId = ref<string | null>(null)
const mouseX = ref(0)
const mouseY = ref(0)
const scroller = ref<HTMLElement>()
const scrollLeft = ref(0)
const { width: scrollerWidth } = useElementSize(scroller)
const pinnedRight = ref(true)

useIntervalFn(() => {
    nowMs.value = Date.now()
}, 1000)

// Rebuild the model when the task set changes, not on every now tick — otherwise a large
// retained history is re-grouped once a second.
watch(
    () => tasks,
    () => {
        modelNow.value = Date.now()
    },
    { immediate: true },
)

// Span at least the viewport's worth of time so "now" sits at the right edge and the chart fills the width.
const minSpanMs = computed(() => (scrollerWidth.value > 0 ? scrollerWidth.value / pxPerMs.value : 0))
const model = computed(() => buildGantt(tasks, modelNow.value, minSpanMs.value))
const widthPx = computed(() => totalWidthPx(model.value.domain, pxPerMs.value))

// Only keep bars intersecting the scroll viewport (plus overscan) in the DOM, so a lane with
// thousands of historical tasks doesn't render them all at once.
const visibleLanes = computed(() => {
    const xMin = scrollLeft.value - OVERSCAN_PX
    const xMax = scrollLeft.value + scrollerWidth.value + OVERSCAN_PX
    const domain = model.value.domain
    const ppm = pxPerMs.value

    return model.value.lanes.map((lane) => ({
        key: lane.key,
        bars: lane.bars.filter((bar) => {
            const left = barLeftPx(bar.startMs, domain, ppm)
            return left + barWidthPx(bar.startMs, bar.endMs, ppm) >= xMin && left <= xMax
        }),
    }))
})

// Clamp to the content edge so the now-line doesn't drift past the rendered timeline between polls.
const nowLineLeft = computed(() => Math.min(widthPx.value, barLeftPx(nowMs.value, model.value.domain, pxPerMs.value)))

// Resolve the hovered task from the live list each update so the tooltip reflects fresh data without re-hovering.
const hoveredTask = computed(() => (hoveredId.value ? (tasks.find((t) => t.taskID === hoveredId.value) ?? null) : null))

let scrollRaf = 0
let autoScrolling = false // true while we drive a smooth scroll to the right edge
let autoScrollTimer = 0

function onScroll() {
    const el = scroller.value
    if (!el) {
        return
    }

    const atRight = el.scrollLeft + el.clientWidth >= el.scrollWidth - 4
    if (atRight) {
        pinnedRight.value = true
        autoScrolling = false
    } else if (!autoScrolling) {
        // A real user scroll away from the edge unpins; the intermediate frames of our own
        // smooth auto-scroll must not.
        pinnedRight.value = false
    }

    // Coalesce scroll events to one viewport recompute per frame.
    if (scrollRaf) {
        return
    }
    scrollRaf = requestAnimationFrame(() => {
        scrollRaf = 0
        if (scroller.value) {
            scrollLeft.value = scroller.value.scrollLeft
        }
    })
}

onBeforeUnmount(() => {
    if (scrollRaf) {
        cancelAnimationFrame(scrollRaf)
    }
    if (autoScrollTimer) {
        clearTimeout(autoScrollTimer)
    }
})

function zoom(factor: number) {
    pxPerMs.value = Math.min(MAX_PX_PER_MS, Math.max(MIN_PX_PER_MS, pxPerMs.value * factor))
}

function onBarEnter(bar: GanttBarType, e: MouseEvent) {
    hoveredId.value = bar.task.taskID
    mouseX.value = e.clientX
    mouseY.value = e.clientY
}

// Keep the latest activity in view as new tasks arrive, unless the user scrolled back in time.
watch(widthPx, async () => {
    await nextTick()
    if (scroller.value && pinnedRight.value) {
        autoScrolling = true
        scroller.value.scrollTo({ left: scroller.value.scrollWidth, behavior: 'smooth' })

        // Safety reset in case the smooth scroll is interrupted before reaching the edge,
        // so the user can still scroll away afterwards.
        if (autoScrollTimer) {
            clearTimeout(autoScrollTimer)
        }
        autoScrollTimer = window.setTimeout(() => {
            autoScrolling = false
        }, 800)
    }
})
</script>
