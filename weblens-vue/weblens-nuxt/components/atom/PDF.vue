<template>
    <div :class="{ 'h-full w-full': true }">
        <iframe
            :src="pdfUrl"
            style="width: 100%; height: 100%"
            @load="onLoadSuccess"
            @error="onLoadError"
        ></iframe>
        <Loader v-if="loadingState === 'loading'" />
        <div v-else-if="loadingState === 'error'">Failed to load PDF</div>
    </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type WeblensMedia from '~/types/weblensMedia'
import { PhotoQuality } from '~/types/weblensMedia'
import Loader from './Loader.vue'

const props = defineProps<{
    media: WeblensMedia
}>()

const pdfUrl = computed(() => {
    return props.media.ImgUrl(PhotoQuality.HighRes)
})
const loadingState = ref<'idle' | 'loading' | 'error' | 'success'>('idle')

async function onLoadSuccess() {
    loadingState.value = 'success'
}

function onLoadError(payload: Event) {
    console.error('PDF load error:', payload)
    loadingState.value = 'error'
}
</script>
