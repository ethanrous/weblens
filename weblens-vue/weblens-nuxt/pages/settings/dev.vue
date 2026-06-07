<template>
    <div :class="{ 'flex h-full flex-col gap-1': true }">
        <WeblensButton
            label="Refresh"
            @click="refresh()"
        >
            <IconRefresh />
        </WeblensButton>
        <TaskTreeTable
            v-if="runningTasks !== null"
            :class="{ 'my-4 max-h-[60vh]': true }"
            empty-text="No running tasks"
            :tasks="runningTasks"
            @cancel="onCancel"
        />
        <span v-else-if="error">{{ error }}</span>

        <WeblensButton
            label="Scan All Media"
            center-content
            :class="{ 'mt-auto': true }"
            @click="scanAllMedia"
        />

        <WeblensButton
            label="Enable trace logging"
            center-content
            :disabled="towerStore.towerInfo?.logLevel === 'trace'"
            @click="enableTraceLogging()"
        />

        <div class="flex items-center gap-2">
            <WeblensButton
                :label="
                    featureFlags?.['embed.processing_enabled'] ? 'Disable embed processing' : 'Enable embed processing'
                "
                center-content
                @click="enableEmbed(!featureFlags?.['embed.processing_enabled'])"
            />

            <div
                :class="{
                    'flex shrink-0 items-center gap-1.5 rounded-md border px-2.5 py-1.5 text-xs': true,
                    'border-green-700 bg-green-900/40 text-green-300': embedAvailable,
                    'border-red-700 bg-red-900/40 text-red-300': !embedAvailable,
                }"
                :title="embedAvailable ? 'Embed service reachable' : 'Embed service unavailable'"
            >
                <span
                    :class="{
                        'h-2 w-2 rounded-full': true,
                        'bg-green-400': embedAvailable,
                        'bg-red-400': !embedAvailable,
                    }"
                />
                <span>{{ embedAvailable ? 'Embed online' : 'Embed offline' }}</span>
            </div>
        </div>

        <Divider />

        <WeblensButton
            label="Drop All Embeddings"
            flavor="danger"
            center-content
            @click="handleDropEmbeddings"
        />

        <WeblensButton
            label="Clean Media"
            flavor="danger"
            center-content
            @click="handleCleanMedia"
        />

        <WeblensButton
            label="Flush Cache"
            flavor="danger"
            center-content
            @click="flushCache"
        />

        <WeblensButton
            label="Clear Zip Cache"
            flavor="danger"
            center-content
            @click="clearZips()"
        />
    </div>
</template>

<script setup lang="ts">
import { IconRefresh } from '@tabler/icons-vue'
import { useIntervalFn } from '@vueuse/core'
import { useWeblensAPI } from '~/api/AllApi'
import { CancelTask } from '~/api/FileBrowserApi'
import Divider from '~/components/atom/Divider.vue'
import TaskTreeTable from '~/components/molecule/TaskTreeTable.vue'
import WeblensButton from '~/components/atom/WeblensButton.vue'

const towerStore = useTowerStore()
const userStore = useUserStore()

useIntervalFn(() => {
    refresh()
}, 5000)

async function scanAllMedia() {
    await useWeblensAPI().FoldersAPI.scanFolder('USERS')
}

async function flushCache() {
    await useWeblensAPI().TowersAPI.flushCache()
}

async function handleCleanMedia() {
    await useWeblensAPI().MediaAPI.dropMedia(userStore.user.username)
}

async function handleDropEmbeddings() {
    await useWeblensAPI().MediaAPI.dropEmbeddings()
}

async function clearZips() {
    await useWeblensAPI().FilesAPI.clearZipCache()
}

async function enableTraceLogging() {
    await useWeblensAPI().TowersAPI.enableTraceLogging()
    await towerStore.refreshTowerInfo()
}

const {
    data: runningTasks,
    refresh,
    error,
} = useAsyncData('running-tasks', async () => {
    const res = await useWeblensAPI().TowersAPI.getRunningTasks()

    // Copy before sorting (don't mutate the response); keep every task so parent subtask counts stay accurate.
    let taskInfos = [...res.data]
    taskInfos = taskInfos.filter((t) => Date.parse(t.startTime!) > 0)

    taskInfos.sort((a, b) => {
        const aMs = new Date(a.startTime ?? '').getTime()
        const bMs = new Date(b.startTime ?? '').getTime()
        if (isNaN(aMs) || isNaN(bMs) || aMs === bMs) {
            return 0
        }

        // Treat tasks with no start time as newest
        if (aMs <= 0) {
            return 1
        } else if (bMs <= 0) {
            return -1
        }

        if (aMs < bMs) {
            return -1
        } else if (aMs > bMs) {
            return 1
        }

        return 0
    })

    return taskInfos
})

async function onCancel(taskID: string) {
    CancelTask(taskID)
    await refresh()
}

const { data: featureFlags, refresh: refreshFeatureFlags } = useAsyncData('feature-flags', async () => {
    const res = await useWeblensAPI().FeatureFlagsAPI.getFlags()

    return res.data
})

async function enableEmbed(enable: boolean) {
    return useWeblensAPI()
        .FeatureFlagsAPI.setFlags([
            {
                configKey: 'embed.processing_enabled',
                configValue: enable as unknown as object,
            },
        ])
        .then(() => {
            return refreshFeatureFlags()
        })
}

const embedAvailable = computed(() => towerStore.towerInfo?.embedAvailable ?? false)

useIntervalFn(() => {
    void towerStore.refreshTowerInfo()
}, 5000)
</script>
