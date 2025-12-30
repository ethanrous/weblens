<template>
    <div :class="{ 'page-root items-center justify-center p-5': true }">
        <IconX
            :class="{
                'hover:bg-card-background-hover bg-background-primary absolute top-2 left-2 z-10 cursor-pointer rounded p-1': true,
            }"
            @click="goHome"
        />

        <div
            v-if="error"
            :class="{ 'flex h-full w-full items-center justify-center gap-2': true }"
        >
            <IconExclamationCircle color="red" />
        </div>

        <VideoPlayer
            v-else-if="media && media.IsVideo()"
            :media="media"
        />

        <MediaImage
            v-else-if="media"
            :quality="PhotoQuality.HighRes"
            :media="media"
            contain
        />
    </div>
</template>

<script setup lang="ts">
import { IconExclamationCircle, IconX } from '@tabler/icons-vue'
import MediaImage from '~/components/atom/MediaImage.vue'
import VideoPlayer from '~/components/molecule/VideoPlayer.vue'
import WeblensFile from '~/types/weblensFile'
import WeblensMedia, { PhotoQuality } from '~/types/weblensMedia'

const route = useRoute()

useHead({
    meta: [
        {
            name: 'og:image',
            content: `${location.origin}/api/v1/media/${route.params.contentID}/thumbnail?quality=${PhotoQuality.HighRes}`,
        },
    ],
})

const contentID = computed(() => {
    const contentID = route.params.contentID
    if (Array.isArray(contentID)) {
        return contentID[0]
    }

    return contentID
})

async function goHome() {
    await WeblensFile.Home().GoTo()
}

const { data: media, error } = useAsyncData(
    'mediaInfo-' + contentID.value,
    async () => {
        const m = await new WeblensMedia({ contentID: contentID.value }).LoadInfo()
        return m
    },
    {
        immediate: true,
        lazy: true,
        watch: [contentID],
    },
)

watchEffect(() => {
    if (error.value) {
        console.error('Error loading media:', error.value)
    }
})
</script>
