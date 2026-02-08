<template>
    <div
        id="filebrowser-container"
        :class="{ 'filebrowser-container relative flex h-full min-h-0 w-full min-w-60 items-center': true }"
    >
        <FileDragCounter />
        <span
            v-if="error"
            :class="{ 'mx-auto rounded-sm border border-amber-700 bg-amber-800/25 p-2 text-center text-xl': true }"
        >
            Fetching files failed
            <span :class="{ 'text-text-secondary mt-1 text-xs': true }">
                {{ error }}
            </span>
        </span>
        <div
            v-else-if="filesStore.loading"
            :class="{ 'm-auto': true }"
        >
            <Loader :class="{ 'h-8 w-8': true }" />
        </div>

        <div
            v-else-if="filesStore.files && !locationStore.isInTimeline"
            :class="{ 'flex h-full w-full': true }"
        >
            <FileScroller
                :files="filesStore.files"
                :no-require-parent-match="filesStore.isSearching"
            />
            <FileHistory />
        </div>

        <MediaTimeline v-else-if="locationStore.isInTimeline" />
    </div>
</template>

<script setup lang="ts">
import Loader from '~/components/atom/Loader.vue'
import FileDragCounter from '~/components/organism/FileDragCounter.vue'
import FileHistory from '~/components/organism/FileHistory.vue'
import FileScroller from '~/components/organism/FileScroller.vue'
import MediaTimeline from '~/components/organism/MediaTimeline.vue'
import useFilesStore from '~/stores/files'
import useLocationStore from '~/stores/location'

const filesStore = useFilesStore()
const locationStore = useLocationStore()

const error = computed(() => {
    return filesStore.fileFetchError
})
</script>
