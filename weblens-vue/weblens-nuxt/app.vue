<template>
    <div>
        <NuxtRouteAnnouncer />
        <ConfirmModal />
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
import { useDark, useDocumentVisibility, useWindowSize } from '@vueuse/core'
import Loader from '~/components/atom/Loader.vue'
import FileSidebar from '~/components/organism/FileSidebar.vue'
import ConfirmModal from './components/molecule/ConfirmModal.vue'
import useLocationStore from './stores/location'
import useWebsocketStore from './stores/websocket'

const userStore = useUserStore()
const towerStore = useTowerStore()
const route = useRoute()

// Initialize stores that need to be active globally
useLocationStore()
const websocketStore = useWebsocketStore()

const isPageVisible = useDocumentVisibility()

const dark = useDark()
watchEffect(() => {
    document?.documentElement?.style?.setProperty('color-scheme', dark.value ? 'dark' : 'light')
})

const windowSize = useWindowSize()

const showSidebar = computed(() => {
    if (towerStore.towerInfo?.role !== TowerRole.CORE) {
        return false
    }

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
    return towerStore.towerInfo && userStore.user.isLoggedIn.isSet()
})

// If the page becomes visible and the websocket is not connected, reload the page. If the page goes idle or the server
// goes down while the user is away, the websocket will disconnect and won't reconnect until the page is reloaded.
watch(isPageVisible, (visible) => {
    if (visible === 'visible' && websocketStore.status === 'CLOSED' && userStore.user.isLoggedIn.get() === true) {
        console.debug(
            `Page is now visible and websocket is not connected (${websocketStore.status}), opening websocket...`,
        )

        reloadNuxtApp()
    }
})
</script>
