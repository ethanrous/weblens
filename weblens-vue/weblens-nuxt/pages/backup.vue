<template>
    <div class="page-root flex-row!">
        <!-- Left Sidebar - Tower List -->
        <div class="bg-background-primary flex h-full w-72 shrink-0 flex-col border-r">
            <!-- Header -->
            <div class="border-b p-4">
                <div class="flex items-center gap-3">
                    <div class="bg-theme-primary/20 flex h-10 w-10 items-center justify-center rounded-lg">
                        <IconServer
                            :size="22"
                            class="text-theme-primary"
                        />
                    </div>
                    <div class="min-w-0 flex-1">
                        <h4 class="truncate leading-tight font-semibold">
                            {{ towerStore.towerInfo?.name }}
                        </h4>
                        <div class="flex items-center gap-2">
                            <span class="text-text-secondary text-xs">
                                {{ capitalizeFirstLetter(towerStore.towerInfo?.role ?? '') }}
                            </span>
                            <WebsocketStatus
                                :ws-status="websocketStore.status"
                                :size="6"
                            />
                        </div>
                    </div>
                </div>
            </div>

            <!-- Tower List -->
            <div class="flex flex-1 flex-col overflow-hidden">
                <div class="flex items-center justify-between px-4 pt-4 pb-2">
                    <span class="text-text-secondary text-xs font-medium tracking-wider uppercase">
                        Linked Towers
                    </span>
                    <span class="bg-card-background-primary text-text-secondary rounded px-2 py-0.5 text-xs">
                        {{ remotesStore.remotes?.size ?? 0 }}
                    </span>
                </div>

                <div class="flex-1 space-y-2 overflow-y-auto px-4 pt-1 pb-4">
                    <div
                        v-if="remotesStore.remotes?.size === 0"
                        class="text-text-tertiary flex flex-col items-center justify-center py-8 text-center"
                    >
                        <IconServerOff
                            :size="40"
                            class="mb-2 opacity-50"
                        />
                        <span class="text-sm">No linked towers</span>
                    </div>

                    <template v-else-if="remotesStore.remotes">
                        <WeblensButton
                            v-for="tower in Array.from(remotesStore.remotes.values())"
                            :key="tower.id"
                            :label="tower.name"
                            :type="tower.id === locationStore.activeTowerID ? 'default' : 'light'"
                            :selected="tower.id === locationStore.activeTowerID"
                            fill-width
                            @click.stop="
                                async () =>
                                    await locationStore.setActiveTowerID(
                                        locationStore.activeTowerID === tower.id ? null : tower.id,
                                    )
                            "
                        >
                            <IconServer :size="18" />
                            <template #rightIcon>
                                <div
                                    class="shrink-0 rounded"
                                    :class="tower.online ? 'bg-valid' : 'bg-danger'"
                                    style="width: 8px; height: 8px"
                                />
                            </template> </WeblensButton
                    ></template>
                </div>
            </div>

            <!-- Settings Button -->
            <div class="border-t p-4">
                <WeblensButton
                    label="Settings"
                    type="outline"
                    fill-width
                    @click.stop="navigateTo('/settings')"
                >
                    <IconSettings :size="18" />
                </WeblensButton>
            </div>
        </div>

        <!-- Main Content - Tower Details -->
        <div class="flex flex-1 flex-col overflow-hidden">
            <!-- Selected Tower Details -->
            <div
                v-if="selectedTower"
                class="animate-fade-in flex flex-1 flex-col overflow-y-auto p-6"
            >
                <NuxtPage />
            </div>

            <!-- Empty State -->
            <div
                v-else-if="selectedTower === null"
                class="flex flex-1 flex-col items-center justify-center p-8"
            >
                <div class="bg-card-background-primary mb-4 rounded-xl p-6">
                    <IconServerCog
                        :size="48"
                        class="text-text-tertiary"
                    />
                </div>
                <h3 class="text-text-secondary mb-2 text-lg font-medium">No Tower Selected</h3>
                <p class="text-text-tertiary max-w-md text-center text-sm">
                    Select a tower from the list to view its details and manage backups.
                </p>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { IconServer, IconServerCog, IconServerOff, IconSettings } from '@tabler/icons-vue'
import WeblensButton from '~/components/atom/WeblensButton.vue'
import WebsocketStatus from '~/components/atom/WebsocketStatus.vue'
import useLocationStore from '~/stores/location'
import useWebsocketStore from '~/stores/websocket'

import { capitalizeFirstLetter } from '~/util/string'
const towerStore = useTowerStore()
const remotesStore = useRemotesStore()
const websocketStore = useWebsocketStore()
const locationStore = useLocationStore()

const selectedTower = computed(() => {
    return locationStore.activeTowerID ? remotesStore.remotes?.get(locationStore.activeTowerID) : null
})

const lastBackup = ref<string>('')

function formatLastBackup(timestamp: number | undefined): string {
    if (!timestamp) {
        return 'Never'
    }

    const now = Date.now()
    const diff = now - timestamp
    const seconds = Math.floor(diff / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)
    const days = Math.floor(hours / 24)

    if (days > 0) {
        return days === 1 ? '1 day ago' : `${days} days ago`
    }
    if (hours > 0) {
        return hours === 1 ? '1 hour ago' : `${hours} hours ago`
    }
    if (minutes > 0) {
        return minutes === 1 ? '1 min ago' : `${minutes} mins ago`
    }
    return 'Just now'
}

function setLastBackupText() {
    if (selectedTower.value) {
        lastBackup.value = formatLastBackup(selectedTower.value.lastBackup)
    } else {
        lastBackup.value = ''
    }
}

watch(selectedTower, () => {
    setLastBackupText()
})
</script>
