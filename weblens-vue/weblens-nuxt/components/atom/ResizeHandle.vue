<template>
    <div
        :class="{
            'absolute top-0 left-0 z-10 h-full w-1.5 cursor-col-resize': true,
            'hover:bg-border-primary': !dragging,
            'bg-text-tertiary': dragging,
        }"
        @mousedown.stop.prevent="onMouseDown"
        @dblclick="emit('reset')"
    />
</template>

<script setup lang="ts">
import { useEventListener, useMouse } from '@vueuse/core'

const emit = defineEmits<{
    (e: 'drag', deltaX: number): void
    (e: 'dragStart' | 'dragEnd' | 'reset'): void
}>()

const dragging = ref(false)
const startX = ref(0)

const mousePos = useMouse()

function onMouseDown() {
    dragging.value = true
    startX.value = mousePos.x.value
    document.body.style.cursor = 'col-resize'
    document.body.style.userSelect = 'none'
    emit('dragStart')
}

useEventListener(window, 'mouseup', () => {
    if (!dragging.value) return
    dragging.value = false
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
    emit('dragEnd')
})

watchEffect(() => {
    if (dragging.value) {
        emit('drag', mousePos.x.value - startX.value)
    }
})
</script>
