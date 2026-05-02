<template>
    <div class="animate-fade-in relative m-auto h-full w-full">
        <div
            v-if="file.contentID && media"
            class="folder-wrapper h-full w-full rounded-xs"
        >
            <svg
                class="folder-border pointer-events-none absolute inset-0 h-full w-full"
                viewBox="2 3 20 17"
            >
                <path
                    d="M5 4h4l3 3h7a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2z"
                    fill="none"
                    stroke="white"
                    stroke-width="0.5"
                    stroke-linejoin="round"
                />
            </svg>
            <MediaImage
                :media="media"
                :quality="quality ?? PhotoQuality.LowRes"
                no-click
                class="folder-content"
            />
        </div>
        <IconFolder
            v-if="!media"
            stroke="0.5"
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
.folder-wrapper {
    position: relative;
    isolation: isolate;
}

.folder-border {
    z-index: 0;
}

.folder-content {
    position: relative;
    z-index: 1;
    width: 100%;
    height: 100%;

    -webkit-mask-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='2 3 20 17'%3E%3Cpath fill='white' d='M5 4h4l3 3h7a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2z'/%3E%3C/svg%3E");
    mask-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='2 3 20 17'%3E%3Cpath fill='white' d='M5 4h4l3 3h7a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2z'/%3E%3C/svg%3E");
    -webkit-mask-size: contain;
    mask-size: contain;
    -webkit-mask-repeat: no-repeat;
    mask-repeat: no-repeat;
    -webkit-mask-position: center;
    mask-position: center;
}
</style>
