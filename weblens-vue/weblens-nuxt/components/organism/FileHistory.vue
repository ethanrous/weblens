<template>
    <div
        id="file-history-sidebar"
        :class="{
            'min-w-0 border-l transition-[width]': true,
            'h-full w-160 min-w-160': locationStore.isHistoryOpen,
            'w-0': !locationStore.isHistoryOpen,
        }"
    >
        <div :class="{ 'flex h-full flex-col p-2': true }">
            <div :class="{ 'mb-2 flex border-b pb-2': true }">
                <IconX
                    :class="{ 'hover:text-text-primary text-text-secondary absolute cursor-pointer rounded': true }"
                    @click="locationStore.setHistoryOpen(false)"
                />
                <h5 :class="{ 'mr-1 ml-auto': true }">History of</h5>
                <FileIcon
                    :file="filesStore.activeFile"
                    with-name
                    :class="{ 'mr-auto': true }"
                />
            </div>
            <div :class="{ 'flex h-full w-full flex-col gap-1 overflow-y-scroll': true }">
                <div
                    v-for="historyItem in historyData"
                    :key="historyItem.eventID"
                >
                    <FileAction :action="historyItem" />
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { IconX } from '@tabler/icons-vue'
import useLocationStore from '~/stores/location'
import FileIcon from '../atom/FileIcon.vue'
import useFilesStore from '~/stores/files'
import { useWeblensAPI } from '~/api/AllApi'
import FileAction from '../molecule/FileAction.vue'
import { PortablePath } from '~/types/portablePath'

const locationStore = useLocationStore()
const filesStore = useFilesStore()

const { data: historyData } = useAsyncData(
    'file-history-' +
        locationStore.activeFolderID +
        '-' +
        locationStore.isHistoryOpen +
        '-' +
        locationStore.viewTimestamp,
    async () => {
        if (!locationStore.isHistoryOpen) {
            return []
        }

        let history = await useWeblensAPI()
            .FoldersAPI.getFolderHistory(locationStore.activeFolderID)
            .then((res) => {
                return res.data
            })

        history = history.filter((item) => {
            if (!item.filepath) {
                return true
            }

            const path = new PortablePath(item.filepath)
            if (path.filename === 'Trash') {
                return false
            }

            return true
        })

        return history
    },
    {
        watch: [
            () => locationStore.viewTimestamp,
            () => locationStore.activeFolderID,
            () => locationStore.isHistoryOpen,
        ],
    },
)
</script>
