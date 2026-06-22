<template>
    <div :class="{ 'mb-40 flex flex-col gap-1': true }">
        <h4>Actions</h4>
        <div :class="{ 'mb-4 flex flex-col gap-4 rounded border p-4': true }">
            <div :class="{ 'flex flex-wrap gap-4': true }">
                <WeblensButton
                    label="Reindex All Files"
                    center-content
                    :class="{ 'mt-auto': true }"
                    @click="scanAllMedia"
                />

                <WeblensButton
                    label="Enable Trace Logging"
                    center-content
                    :disabled="towerStore.towerInfo?.logLevel === 'trace'"
                    @click="enableTraceLogging()"
                />

                <div class="flex items-center gap-2">
                    <WeblensButton
                        :label="featureFlags?.['embed.processing_enabled'] ? 'Disable Embedding' : 'Enable Embedding'"
                        center-content
                        @click="enableEmbed(!featureFlags?.['embed.processing_enabled'])"
                    />

                    <div
                        :class="{
                            'flex shrink-0 items-center gap-1.5 rounded-md border px-2.5 py-1.5 text-xs': true,
                            'border-green-700 bg-green-900/40 text-green-300': embedAvailable,
                            'border-red-700 bg-red-900/40 text-red-300': !embedAvailable,
                            'opacity-50 select-none': !featureFlags?.['embed.processing_enabled'],
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
            </div>

            <div :class="{ 'flex flex-wrap gap-4': true }">
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
        </div>

        <div :class="{ 'flex items-center gap-2': true }">
            <h4>Tasks</h4>
            <div :class="{ 'ml-auto flex': true }">
                <WeblensButton
                    label="Gantt"
                    type="light"
                    merge="row"
                    :selected="taskView === 'gantt'"
                    :on-click="() => (taskView = 'gantt')"
                />
                <WeblensButton
                    label="Table"
                    type="light"
                    merge="row"
                    :selected="taskView === 'table'"
                    :on-click="() => (taskView = 'table')"
                />
            </div>
        </div>

        <TaskGanttChart
            v-if="taskView === 'gantt'"
            :class="{ 'my-4': true }"
            :tasks="allTasks"
            empty-text="No tasks observed yet"
        />
        <TaskTreeTable
            v-else
            :class="{ 'my-4 max-h-[60vh]': true }"
            empty-text="No running tasks"
            :tasks="runningTasks"
            @cancel="onCancel"
        />
        <span v-if="error">{{ error }}</span>

        <h4>Debug</h4>

        <WeblensOptions
            :options="{
                file: { label: 'File', value: 'file', default: true },
                media: { label: 'Media', value: 'media' },
            }"
            @update:value="
                (val) => {
                    if (!val) return
                    selectedDebugOption = val
                }
            "
        />

        <WeblensInput
            :placeholder="selectedDebugOption === 'file' ? 'FileID' : 'MediaID'"
            show-submit
            @submit="handleDebugSubmit"
        />
        <pre>
            {{ debugReturn }}
        </pre>

        <div
            v-if="debugMedia"
            :class="{ 'h-250': true }"
        >
            <MediaImage
                :media="debugMedia"
                :class="{ 'min-h-40 w-full': true }"
                :quality="PhotoQuality.HighRes"
                contain
            />
        </div>
    </div>
</template>

<script setup lang="ts">
import { useIntervalFn } from '@vueuse/core'
import { useWeblensAPI } from '~/api/AllApi'
import { CancelTask } from '~/api/FileBrowserApi'
import TaskTreeTable from '~/components/molecule/TaskTreeTable.vue'
import TaskGanttChart from '~/components/molecule/TaskGanttChart.vue'
import WeblensButton from '~/components/atom/WeblensButton.vue'
import WeblensInput from '~/components/atom/WeblensInput.vue'
import WeblensOptions from '~/components/atom/WeblensOptions.vue'
import MediaImage from '~/components/atom/MediaImage.vue'
import WeblensMedia, { PhotoQuality } from '~/types/weblensMedia'

const towerStore = useTowerStore()
const userStore = useUserStore()

const selectedDebugOption = ref<string>('file')
const taskView = ref<'gantt' | 'table'>('gantt')

async function scanAllMedia() {
    await useWeblensAPI().FoldersAPI.scanFolder('USERS', undefined, true)
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

const { allTasks, runningTasks, error, refresh } = useTaskHistory()

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

const debugReturn = ref<string>('')
const debugMedia = ref<WeblensMedia>()

async function getMediaInfo(mediaID: string) {
    return useWeblensAPI()
        .MediaAPI.getMediaInfo(mediaID)
        .then((res) => {
            return res
        })
}

function handleDebugSubmit(val: string) {
    debugReturn.value = ''
    debugMedia.value = undefined

    if (selectedDebugOption.value === 'file') {
        useWeblensAPI()
            .FilesAPI.getFile(val)
            .then((res) => {
                debugReturn.value = JSON.stringify(res.data, null, 2)
                if (res.data.contentID) {
                    getMediaInfo(res.data.contentID)
                        .then((mediaRes) => {
                            debugMedia.value = new WeblensMedia(mediaRes.data)
                        })
                        .catch((err) => {
                            console.error('Error fetching media info:', err)
                        })
                }
            })
            .catch((err) => {
                debugReturn.value = 'Error fetching file info: ' + err.message
                console.error('Error fetching file info:', err)
            })
    } else if (selectedDebugOption.value === 'media') {
        getMediaInfo(val)
            .then((res) => {
                debugReturn.value = JSON.stringify(res.data, null, 2)
                debugMedia.value = new WeblensMedia(res.data)
            })
            .catch((err) => {
                debugReturn.value = 'Error fetching media info: ' + err.message
                console.error('Error fetching media info:', err)
            })
    }
}
</script>
