<template>
    <div :class="{ 'flex h-full flex-col gap-1': true }">
        <WeblensButton
            label="Refresh"
            @click="refresh()"
        >
            <IconRefresh />
        </WeblensButton>
        <Table
            v-if="runningTasks !== null"
            :class="{ 'my-4': true }"
            empty-text="No running tasks"
            :columns="['taskID', 'jobName', 'status', 'worker', 'startTime', 'cancel', 'result']"
            :rows="runningTasks"
        />
        <span v-else-if="error">{{ error }}</span>

        <WeblensButton
            label="Scan All Media"
            center-content
            :class="{ 'mt-auto': true }"
            @click="scanAllMedia"
        />

        <div class="flex flex-col gap-2 py-2">
            <WeblensCheckbox
                label="Trace logging"
                :checked="towerStore.towerInfo?.logLevel === 'trace'"
                @checked:changed="(v) => v && enableTraceLogging()"
            />
            <WeblensCheckbox
                label="Allow registrations"
                :checked="featureFlags?.['auth.allow_registrations'] ?? false"
                @checked:changed="(v) => setFlag('auth.allow_registrations', v)"
            />
            <WeblensCheckbox
                label="HDIR image processing"
                :checked="featureFlags?.['media.hdir_processing_enabled'] ?? false"
                @checked:changed="(v) => setFlag('media.hdir_processing_enabled', v)"
            />
            <WeblensCheckbox
                label="WebDAV file access"
                :checked="featureFlags?.['webdav.enabled'] ?? false"
                @checked:changed="(v) => setFlag('webdav.enabled', v)"
            />
        </div>

        <Divider />

        <WeblensButton
            label="Clear Media HDIR Data"
            flavor="danger"
            center-content
            @click="handleClearHDIRs"
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
import Table from '~/components/atom/Table.vue'
import WeblensButton from '~/components/atom/WeblensButton.vue'
import WeblensCheckbox from '~/components/atom/WeblensCheckbox.vue'
import { TableType, type TableColumn } from '~/types/table'

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

async function handleClearHDIRs() {
    await useWeblensAPI().MediaAPI.dropHDIRs()
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

    let taskInfos = res.data
    taskInfos = taskInfos.filter((t) => t.status === '')

    let tasks = taskInfos.map((task) => ({
        taskID: task.taskID,
        jobName: task.jobName,
        status: task.status,
        worker: task.workerID,
        startTime: task.startTime,
        cancel: {
            tableType: TableType.Button,
            label: 'Cancel',
            flavor: 'danger',
            onclick: async () => {
                CancelTask(task.taskID)
                await refresh()
            },
        } as TableColumn<TableType.Button>,
        result: { tableType: TableType.JSON, value: task.result } as TableColumn<TableType.JSON>,
    }))

    tasks = tasks.sort((a, b) => {
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
    return tasks
})

const { data: featureFlags, refresh: refreshFeatureFlags } = useAsyncData('feature-flags', async () => {
    const res = await useWeblensAPI().FeatureFlagsAPI.getFlags()

    return res.data
})

async function setFlag(key: string, value: boolean) {
    return useWeblensAPI()
        .FeatureFlagsAPI.setFlags([
            {
                configKey: key,
                configValue: value as unknown as object,
            },
        ])
        .then(() => {
            return refreshFeatureFlags()
        })
}
</script>
