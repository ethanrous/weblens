<template>
    <div v-if="!media"></div>
    <VideoPlayer
        v-else-if="media.IsVideo()"
        :media="media"
    />

    <PDF
        v-else-if="media.IsPDF()"
        :media="media"
    />

    <MediaImage
        v-else-if="media"
        :media="media"
        :quality="PhotoQuality.HighRes"
        :class="{ 'min-w-0 p-2 max-sm:my-auto lg:p-0': true }"
        :style="{
            height: `calc(${presentationSize.height.value}px - 1rem)`,
            maxHeight: `calc(${presentationSize.height.value}px - 1rem)`,
            maxWidth: presentationSize.width.value + 'px',
        }"
        :contain="true"
        no-click
    />
</template>

<script setup lang="ts">
import { PhotoQuality } from '~/types/weblensMedia'
import MediaImage from '../atom/MediaImage.vue'
import PDF from '../atom/PDF.vue'
import VideoPlayer from './VideoPlayer.vue'

const mediaStore = useMediaStore()

const props = defineProps<{
    mediaId: string
    presentationSize: {
        width: { value: number }
        height: { value: number }
    }
}>()

const media = computed(() => {
    return mediaStore.mediaMap.get(props.mediaId)
})
</script>
