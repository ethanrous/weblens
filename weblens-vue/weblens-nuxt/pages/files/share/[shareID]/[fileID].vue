<template>
    <div
        v-if="!locationStore.activeShare"
        :class="{ 'flex h-full w-full items-center justify-center': true }"
    >
        <Loader />
    </div>
    <div
        v-else-if="!locationStore.activeShare.isDir"
        ref="presentationContainer"
        :class="{ 'flex w-full items-center': true }"
    >
        <PresentationMediaContent
            :media-id="mediaID"
            :presentation-size="presentationSize"
        />
        <div :class="{ 'mr-2 ml-auto': true }">
            <PresentationFileInfo
                :class="{ 'mx-8 w-max': true }"
                :file-id="locationStore.activeShare.fileID"
                :media-id="mediaID"
            />
        </div>
    </div>
    <FileBrowser v-else />
</template>

<script setup lang="ts">
import { useElementSize } from '@vueuse/core'
import Loader from '~/components/atom/Loader.vue'
import PresentationFileInfo from '~/components/molecule/PresentationFileInfo.vue'
import PresentationMediaContent from '~/components/molecule/PresentationMediaContent.vue'

import FileBrowser from '~/components/organism/FileBrowser.vue'
import useFilesStore from '~/stores/files'
import useLocationStore from '~/stores/location'

const presentationContainer = ref<HTMLElement | null>(null)
const presentationSize = useElementSize(presentationContainer)

const locationStore = useLocationStore()
const filesStore = useFilesStore()

const mediaID = computed(() => {
    return filesStore.activeFile?.contentID ?? ''
})
</script>
