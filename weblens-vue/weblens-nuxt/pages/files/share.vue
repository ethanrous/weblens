<template>
    <div
        id="filebrowser-container"
        :class="{ 'page-root': true }"
    >
        <FileScroller
            v-if="locationStore.inShareRoot"
            :files="filesStore.children ?? []"
        />
        <NuxtPage v-if="!locationStore.inShareRoot" />
    </div>
</template>

<script setup lang="ts">
import FileScroller from '~/components/organism/FileScroller.vue'
import useFilesStore from '~/stores/files'
import useLocationStore from '~/stores/location'

const locationStore = useLocationStore()
const filesStore = useFilesStore()
</script>
