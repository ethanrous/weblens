<template>
    <div
        id="filebrowser-container"
        :class="{ 'flex h-full min-h-0 w-full': true }"
    >
        <FileDragCounter />
        <FileScroller
            v-if="filesStore.files && !locationStore.isInTimeline"
            :files="filesStore.files"
        />
        <MediaTimeline v-else-if="locationStore.isInTimeline" />
    </div>
</template>

<script setup lang="ts">
import FileDragCounter from '~/components/organism/FileDragCounter.vue'
import FileScroller from '~/components/organism/FileScroller.vue'
import MediaTimeline from '~/components/organism/MediaTimeline.vue'
import useFilesStore from '~/stores/files'
import useLocationStore from '~/stores/location'

const filesStore = useFilesStore()
const locationStore = useLocationStore()
</script>
