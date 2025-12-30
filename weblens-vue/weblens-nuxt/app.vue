<template>
    <div>
        <NuxtRouteAnnouncer />
    </div>
    <NuxtLayout>
        <div
            v-if="loaded"
            :class="{ 'relative flex h-screen w-screen': true }"
        >
            <FileSidebar
                v-if="showSidebar"
                :collapsed="sidebarClosed"
            />
            <NuxtPage />
        </div>
        <div
            v-else
            :class="{ 'page-root items-center justify-center': true }"
        >
            <Loader />
        </div>
    </NuxtLayout>
</template>

<script setup lang="ts">
import { useDark, useWindowSize } from '@vueuse/core'
import Loader from '~/components/atom/Loader.vue'
import FileSidebar from '~/components/organism/FileSidebar.vue'

const userStore = useUserStore()
const towerStore = useTowerStore()
const route = useRoute()

useWebsocketStore()

const dark = useDark()
watchEffect(() => {
    document?.documentElement?.style?.setProperty('color-scheme', dark.value ? 'dark' : 'light')
})

const windowSize = useWindowSize()

const showSidebar = computed(() => {
    const routeName = route.name as string
    if (!routeName) {
        return false
    }

    if (routeName.startsWith('files') || routeName.startsWith('settings')) {
        return true
    }

    return false
})

const sidebarClosed = computed(() => {
    if (windowSize.width.value < 768 || windowSize.height.value < 500) {
        return true
    }

    const routeName = route.name as string
    if (routeName === 'media-contentID' || routeName.startsWith('settings')) {
        return true
    }

    return false
})

const loaded = computed(() => {
    return towerStore.towerInfo && userStore.user
})
</script>
