<template>
    <div
        id="filebrowser-container"
        :class="{ 'filebrowser-container relative flex h-full min-h-0 w-full min-w-60 items-center': true }"
    >
        <FileDragCounter />
        <RewindIndicator />

        <div :class="{ 'relative flex h-full w-full min-w-0': true }">
            <NoResults
                v-if="
                    filesStore.fileFetchError && filesStore.fileFetchError.status === 404 && locationStore.isViewingPast
                "
            />

            <ErrorCard
                v-else-if="filesStore.fileFetchError"
                :error="filesStore.fileFetchError"
            />

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
            </div>

            <MediaTimeline v-else-if="locationStore.isInTimeline" />
        </div>

        <FileHistory />
    </div>
</template>

<script setup lang="ts">
import Loader from '~/components/atom/Loader.vue'
import ErrorCard from '~/components/molecule/ErrorCard.vue'
import NoResults from '~/components/molecule/NoResults.vue'
import RewindIndicator from '~/components/molecule/RewindIndicator.vue'
import FileDragCounter from '~/components/organism/FileDragCounter.vue'
import FileHistory from '~/components/organism/FileHistory.vue'
import FileScroller from '~/components/organism/FileScroller.vue'
import MediaTimeline from '~/components/organism/MediaTimeline.vue'
import useFilesStore from '~/stores/files'
import useLocationStore from '~/stores/location'

const filesStore = useFilesStore()
const locationStore = useLocationStore()
</script>
