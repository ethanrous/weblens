<template>
    <div
        ref="seeker"
        :class="{
            'group/seeker relative flex h-3 w-full cursor-pointer items-center rounded-[2px] py-1': true,
        }"
        @mousedown.stop="dragging = true"
        @dragover="handleClick"
        @touchstart.stop="handleTouch"
    >
        <div
            :class="{
                'bg-card-background-primary absolute h-2 w-full rounded-[1px] rounded-l-[2px] transition-[height] duration-300 group-hover/seeker:h-3': true,
            }"
        />
        <div
            :class="{
                'bg-text-primary absolute h-2 rounded-[1px] rounded-l-[2px] transition-[height] duration-300 group-hover/seeker:h-3': true,
            }"
            :style="{
                width: internalPercent + '%',
            }"
        />

        <div
            :class="{
                'bg-text-primary absolute z-20 h-2 w-1 rounded-none opacity-100 transition-[opacity,height,border-radius] duration-300 group-hover/seeker:h-4 group-hover/seeker:rounded-sm group-hover/seeker:opacity-100': true,
            }"
            :style="{
                left: internalPercent + '%',
                top: '50%',
                transform: 'translateX(-50%) translateY(-50%)',
            }"
        />
    </div>
</template>

<script setup lang="ts">
import { useMouse, useMousePressed, useThrottleFn } from '@vueuse/core'

const seeker = ref<HTMLDivElement | null>(null)

const props = defineProps<{
    percent: number
}>()

const emit = defineEmits<{
    (e: 'seek', percent: number): void
}>()

const internalPercent = ref<number>(0)
const dragging = ref<boolean>(false)

const mousePos = useMouse()
const mousePressed = useMousePressed({ target: seeker })

watchEffect(() => {
    internalPercent.value = Math.min(Math.max(0, props.percent), 100)
})

watchEffect(() => {
    if (!mousePressed.pressed.value) {
        dragging.value = false
    }
})

const doEmit = useThrottleFn(
    (percent) => {
        emit('seek', percent)
    },
    500,
    true,
)

function handleDragOffset(target: HTMLDivElement, absX: number) {
    const rect = target.getBoundingClientRect()
    const offsetX = absX - rect.left
    const percent = Math.min(Math.max(0, (offsetX / rect.width) * 100), 100)

    internalPercent.value = percent
    doEmit(percent)
}

watchEffect(() => {
    if (dragging.value && seeker.value) {
        handleDragOffset(seeker.value, mousePos.x.value)
    }
})

function handleClick(event: MouseEvent) {
    handleDragOffset(event.currentTarget as HTMLDivElement, event.clientX)
}

function handleTouch(event: TouchEvent) {
    if (event.touches.length === 0 || !seeker.value) return
    handleDragOffset(seeker.value, event.touches[0].clientX)
}
</script>
