<template>
    <div v-if="selectedTower">
        <!-- Tower Header -->
        <div class="mb-6 flex items-start justify-between">
            <div class="flex items-center gap-4">
                <div
                    class="flex h-14 w-14 items-center justify-center rounded-xl"
                    :class="selectedTower.online ? 'bg-valid/20' : 'bg-danger/20'"
                >
                    <IconServer
                        :size="28"
                        :class="selectedTower.online ? 'text-valid' : 'text-danger'"
                    />
                </div>
                <div>
                    <h2 class="text-2xl font-semibold">{{ selectedTower.name }}</h2>
                    <div class="mt-1 flex items-center gap-2">
                        <WebsocketStatus
                            :ws-status="selectedTower.online ? 'OPEN' : 'CLOSED'"
                            show-as-text
                            :size="10"
                        />
                    </div>
                </div>
            </div>
            <WeblensButton
                label="Backup Now"
                :disabled="!selectedTower.online"
                @click="useWeblensAPI().TowersAPI.launchBackup(selectedTower.id)"
            >
                <IconCloudUpload :size="18" />
            </WeblensButton>
        </div>

        <!-- Stats Cards -->
        <div class="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <div class="bg-card-background-primary rounded-lg border p-4">
                <div class="text-text-secondary mb-1 flex items-center gap-2 text-sm">
                    <IconDatabase :size="16" />
                    Total Backup Size
                </div>
                <span class="text-2xl font-semibold">
                    {{ humanBytesStr(selectedTower.backupSize) }}
                </span>
            </div>
            <div class="bg-card-background-primary flex rounded-lg border p-4">
                <div :class="{ 'flex flex-col': true }">
                    <div class="text-text-secondary mb-1 flex items-center gap-2 text-sm">
                        <IconActivity :size="16" />
                        Status
                    </div>
                    <span
                        class="text-2xl font-semibold"
                        :class="selectedTower.online ? 'text-valid' : 'text-danger'"
                    >
                        {{ selectedTower.online ? 'Online' : 'Offline' }}
                    </span>
                </div>
                <WeblensButton
                    v-if="!selectedTower.online"
                    type="outline"
                    :class="{ 'ml-auto': true }"
                    :disabled="websocketStore.status !== 'OPEN'"
                    @click="
                        () => {
                            if (selectedTower) attemptReconnectTower(selectedTower.id)
                        }
                    "
                >
                    <IconRefresh />
                </WeblensButton>
            </div>
            <div class="bg-card-background-primary rounded-lg border p-4">
                <div class="text-text-secondary mb-1 flex items-center gap-2 text-sm">
                    <IconClock :size="16" />
                    Last Backup
                </div>
                <span class="text-2xl font-semibold">
                    {{ lastBackup }}
                </span>
                <span
                    v-if="selectedTower.lastBackup"
                    class="text-text-tertiary mt-1 block text-xs"
                >
                    {{ formatLastBackupDate(selectedTower.lastBackup) }}
                </span>
            </div>
        </div>

        <!-- Connection Details -->
        <div class="bg-card-background-primary mb-6 rounded-lg border">
            <div class="border-b px-4 py-3">
                <h4 class="font-medium">Connection Details</h4>
            </div>
            <div class="divide-y">
                <div class="flex items-center gap-2 px-4 py-3">
                    <IconId
                        :size="18"
                        class="text-text-secondary shrink-0"
                    />
                    <div class="min-w-0 flex-1">
                        <span class="text-text-secondary ml-1 block text-xs">Tower ID</span>
                        <CopyBox :text="selectedTower.id" />
                    </div>
                </div>
                <div class="flex items-center gap-2 px-4 py-3">
                    <IconWorldWww
                        :size="18"
                        class="text-text-secondary shrink-0"
                    />
                    <div class="min-w-0 flex-1">
                        <span class="text-text-secondary ml-1 block text-xs">Core Address</span>
                        <CopyBox :text="selectedTower.coreAddress" />
                    </div>
                </div>
            </div>
        </div>

        <!-- Backup Status + History -->
        <div class="bg-card-background-primary mb-6 rounded-lg border">
            <div class="border-b px-4 py-3">
                <h4 class="font-medium">Backup Status</h4>
            </div>
            <div class="p-4">
                <!-- No backup running state -->
                <div
                    v-if="!selectedTowerBackup"
                    class="text-text-secondary flex items-center gap-2 py-4"
                >
                    <IconCloudOff :size="20" />
                    <span>No Backup Running</span>
                </div>

                <!-- Backup in progress -->
                <div
                    v-else-if="!selectedTowerBackup.Completed"
                    class="space-y-4"
                >
                    <!-- Progress bar -->
                    <div>
                        <div class="mb-2 flex items-center justify-between">
                            <span class="text-text-secondary text-sm">Progress</span>
                            <span class="font-mono text-sm font-medium"> {{ backupProgress.toFixed(1) }}% </span>
                        </div>
                        <ProgressSquare
                            :progress="backupProgress"
                            class="h-3 w-full"
                        />
                    </div>

                    <!-- Stats grid -->
                    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
                        <div>
                            <span class="text-text-tertiary block text-xs">Copied</span>
                            <span class="font-mono text-sm font-medium">
                                {{ humanBytesStr(selectedTowerBackup.BytesSoFar) }}
                            </span>
                        </div>
                        <div>
                            <span class="text-text-tertiary block text-xs">Total</span>
                            <span class="font-mono text-sm font-medium">
                                {{ humanBytesStr(selectedTowerBackup.TotalBytes) }}
                            </span>
                        </div>
                        <div>
                            <span class="text-text-tertiary block text-xs">Elapsed</span>
                            <span class="font-mono text-sm font-medium">
                                {{ formatDuration(elapsedTime) }}
                            </span>
                        </div>
                        <div>
                            <span class="text-text-tertiary block text-xs">Remaining</span>
                            <span class="font-mono text-sm font-medium">
                                {{ estimatedTimeRemaining }}
                            </span>
                        </div>
                    </div>

                    <!-- Transfer rate -->
                    <div class="border-t pt-3">
                        <div class="flex items-center gap-2">
                            <IconArrowUp
                                :size="16"
                                class="text-valid"
                            />
                            <span class="text-text-secondary text-sm">Transfer Rate:</span>
                            <span class="font-mono text-sm font-medium"> {{ humanBytesStr(transferRate) }}/s </span>
                        </div>
                    </div>
                </div>

                <!-- Backup completed -->
                <div
                    v-else
                    class="space-y-4"
                >
                    <!-- Success banner -->
                    <div class="bg-valid/10 flex items-center gap-3 rounded-lg p-3">
                        <div class="bg-valid/20 flex h-10 w-10 items-center justify-center rounded-full">
                            <IconCheck
                                :size="20"
                                class="text-valid"
                            />
                        </div>
                        <div>
                            <span class="text-valid font-medium">Backup Complete</span>
                            <span class="text-text-secondary block text-sm">
                                Finished {{ formatLastBackup(selectedTowerBackup.EndTime) }}
                            </span>
                        </div>
                    </div>

                    <!-- Final stats -->
                    <div class="grid grid-cols-2 gap-4 lg:grid-cols-2">
                        <div class="bg-background-secondary rounded-lg p-3">
                            <span class="text-text-tertiary block text-xs">Total Data Copied</span>
                            <span class="font-mono text-lg font-medium">
                                {{ humanBytesStr(selectedTowerBackup.TotalBytes) }}
                            </span>
                        </div>
                        <div class="bg-background-secondary rounded-lg p-3">
                            <span class="text-text-tertiary block text-xs">Duration</span>
                            <span class="font-mono text-lg font-medium">
                                {{ formatDuration(selectedTowerBackup.EndTime - selectedTowerBackup.StartTime) }}
                            </span>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="bg-card-background-primary mb-6 rounded-lg border">
            <div class="border-b px-4 py-3">
                <h4 class="font-medium">File Tree</h4>
                <FileList :files="filesStore.files" />
            </div>
            <div></div>
        </div>
    </div>
</template>

<script setup lang="ts">
import {
    IconActivity,
    IconArrowUp,
    IconCheck,
    IconClock,
    IconCloudOff,
    IconCloudUpload,
    IconDatabase,
    IconId,
    IconRefresh,
    IconServer,
    IconWorldWww,
} from '@tabler/icons-vue'
import { useWeblensAPI } from '~/api/AllApi'
import ProgressSquare from '~/components/atom/ProgressSquare.vue'
import WeblensButton from '~/components/atom/WeblensButton.vue'
import WebsocketStatus from '~/components/atom/WebsocketStatus.vue'
import CopyBox from '~/components/molecule/CopyBox.vue'
import FileList from '~/components/organism/FileList.vue'
import useFilesStore from '~/stores/files'
import useLocationStore from '~/stores/location'
import useWebsocketStore from '~/stores/websocket'
import { WsAction } from '~/types/websocket'
import { humanBytesStr } from '~/util/humanBytes'

const remotesStore = useRemotesStore()
const backupStore = useBackupStore()
const websocketStore = useWebsocketStore()
const filesStore = useFilesStore()
const locationStore = useLocationStore()

// Reactive timestamp for real-time progress updates
const currentTime = ref(Date.now())

const selectedTower = computed(() => {
    return locationStore.activeTowerID ? remotesStore.remotes?.get(locationStore.activeTowerID) : null
})

const selectedTowerBackup = computed(() => {
    return backupStore.activeBackups.get(locationStore.activeTowerID || '') || null
})

// Backup progress calculations
const backupProgress = computed(() => {
    if (!selectedTowerBackup.value || selectedTowerBackup.value.TotalBytes === 0) {
        return 0
    }
    return (selectedTowerBackup.value.BytesSoFar / selectedTowerBackup.value.TotalBytes) * 100
})

const elapsedTime = computed(() => {
    if (!selectedTowerBackup.value) {
        return 0
    }
    if (selectedTowerBackup.value.Completed && selectedTowerBackup.value.EndTime) {
        return selectedTowerBackup.value.EndTime - selectedTowerBackup.value.StartTime
    }
    return currentTime.value - selectedTowerBackup.value.StartTime
})

const transferRate = computed(() => {
    if (!selectedTowerBackup.value || elapsedTime.value === 0) {
        return 0
    }
    return (selectedTowerBackup.value.BytesSoFar / elapsedTime.value) * 1000
})

const estimatedTimeRemaining = computed(() => {
    if (!selectedTowerBackup.value || transferRate.value === 0) {
        return 'Calculating...'
    }
    const bytesRemaining = selectedTowerBackup.value.TotalBytes - selectedTowerBackup.value.BytesSoFar
    const msRemaining = (bytesRemaining / transferRate.value) * 1000
    return formatDuration(msRemaining)
})

function formatDuration(ms: number): string {
    if (ms <= 0) {
        return '0s'
    }
    const seconds = Math.floor(ms / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)

    if (hours > 0) {
        const remainingMins = minutes % 60
        return `${hours}h ${remainingMins}m`
    }
    if (minutes > 0) {
        const remainingSecs = seconds % 60
        return `${minutes}m ${remainingSecs}s`
    }
    return `${seconds}s`
}

function attemptReconnectTower(towerID: string) {
    useWebsocketStore().send({
        action: WsAction.RefreshTower,
        content: { towerID: towerID },
    })
}

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

function updateTimers() {
    currentTime.value = Date.now()
    if (selectedTower.value) {
        lastBackup.value = formatLastBackup(selectedTower.value.lastBackup)
    } else {
        lastBackup.value = ''
    }
}

watch(
    selectedTower,
    () => {
        updateTimers()
    },
    { immediate: true },
)

function formatLastBackupDate(timestamp: number | undefined): string {
    if (!timestamp) {
        return ''
    }
    return new Date(timestamp).toLocaleString()
}

let timer: NodeJS.Timeout | null = null
onMounted(() => {
    timer = setInterval(updateTimers, 1000)
})

onUnmounted(() => {
    // Crucial: clear the interval when the component is unmounted
    if (timer) {
        clearInterval(timer)
    }
})
</script>
