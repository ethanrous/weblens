<template>
    <div class="animate-fade-in relative m-auto h-full w-full">
        <div
            v-if="file.contentID && media"
            class="folder-cover-mask h-full w-full"
        >
            <MediaImage
                :media="media"
                :quality="quality ?? PhotoQuality.LowRes"
                no-click
            />
        </div>
        <IconFolder
            size="100%"
            stroke="1"
            class="absolute inset-0 h-full w-full"
        />
    </div>
</template>

<script setup lang="ts">
import { IconFolder } from '@tabler/icons-vue'
import type WeblensMedia from '~/types/weblensMedia'
import MediaImage from './MediaImage.vue'
import type WeblensFile from '~/types/weblensFile'
import { PhotoQuality } from '~/types/weblensMedia'

defineProps<{
    file: WeblensFile
    media?: WeblensMedia
    quality?: PhotoQuality
}>()
</script>

<style scoped>
.folder-cover-mask {
    mask-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'%3E%3Cpath fill='white' d='M5 4h4l3 3h7a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2z'/%3E%3C/svg%3E");
    -webkit-mask-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'%3E%3Cpath fill='white' d='M5 4h4l3 3h7a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2z'/%3E%3C/svg%3E");
    mask-size: contain;
    -webkit-mask-size: contain;
    mask-repeat: no-repeat;
    -webkit-mask-repeat: no-repeat;
    mask-position: center;
    -webkit-mask-position: center;
}
</style>
