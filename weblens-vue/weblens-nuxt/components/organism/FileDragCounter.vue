<template>
    <div
        :class="{
            'absolute z-30': true,
            gone: !filesStore.dragging,
        }"
        :style="{
            top: mousePos.y.value - fbBound.top.value + 16 + 'px',
            left: mousePos.x.value - fbBound.left.value + 16 + 'px',
        }"
    >
        <span :class="{ 'text-lg font-bold': true }">{{ filesStore.selectedFiles.size }}</span>
    </div>
</template>

<script setup lang="ts">
import { useElementBounding, useMouse } from '@vueuse/core'
import useFilesStore from '~/stores/files'

const target = ref<HTMLElement>()
const fbBound = useElementBounding(target)

const filesStore = useFilesStore()
const mousePos = useMouse({
    type: 'page',
    target: target,
})

onMounted(() => {
    const tmpTarget = document.getElementById('filebrowser-container')
    if (!tmpTarget) {
        return
    }

    target.value = tmpTarget
})
</script>
