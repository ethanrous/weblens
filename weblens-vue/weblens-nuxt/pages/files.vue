<template>
    <div :class="{ 'page-root relative': true }">
        <Presentation />
        <UploadProgress />

        <div :class="{ 'flex h-full w-full max-w-full min-w-0 flex-col': true }">
            <FileHeader />
            <NuxtPage />
            <PathCrumbs />
        </div>

        <WebsocketStatus
            :class="{ 'absolute right-4 bottom-4 hidden sm:block': true }"
            :status="wsStore.status"
        />
    </div>
</template>

<script setup lang="ts">
import WebsocketStatus from '~/components/molecule/WebsocketStatus.vue'
import FileHeader from '~/components/organism/FileHeader.vue'
import PathCrumbs from '~/components/organism/PathCrumbs.vue'
import Presentation from '~/components/organism/Presentation.vue'
import UploadProgress from '~/components/organism/UploadProgress.vue'

const wsStore = useWebsocketStore()
</script>
