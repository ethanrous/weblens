<template>
    <div
        id="file-history-sidebar"
        :class="{
            'relative min-w-0 overflow-hidden border-l': true,
            'transition-[width]': !isDragging,
            'h-full': locationStore.isHistoryOpen,
        }"
        :style="{
            width: locationStore.isHistoryOpen ? locationStore.historyWidth + 'px' : '0px',
            'min-width': locationStore.isHistoryOpen ? locationStore.historyWidth + 'px' : '0px',
        }"
    >
        <ResizeHandle
            v-if="locationStore.isHistoryOpen"
            @drag-start="onDragStart"
            @drag="onDrag"
            @drag-end="isDragging = false"
            @reset="locationStore.setHistoryWidth(640)"
        />
        <div :class="{ 'flex h-full flex-col p-2': true }">
            <div :class="{ 'mb-2 flex border-b pb-2': true }">
                <IconX
                    :class="{ 'hover:text-text-primary text-text-secondary absolute cursor-pointer rounded': true }"
                    @click="locationStore.setHistoryOpen(false)"
                />
                <div :class="{ 'mx-auto inline-flex': true }">
                    <h5 :class="{ 'mr-1': true }">History of</h5>
                    <FileIcon
                        :file="filesStore.activeFile"
                        with-name
                    />
                </div>
            </div>
            <div
                v-if="!error"
                :class="{ 'flex h-full w-full flex-col gap-2 overflow-y-scroll': true }"
                @click.stop="locationStore.setViewTimestamp(0)"
            >
                <FileEventGroup
                    v-for="group in groupedHistory"
                    :key="group.eventID"
                    :group="group"
                />
            </div>
            <ErrorCard
                v-else
                :error="error"
                message="Failed to load file history."
                :class="{ 'left-[unset] mx-auto block! translate-0!': true }"
            />
        </div>
    </div>
</template>

<script setup lang="ts">
import { IconX } from '@tabler/icons-vue'
import useLocationStore from '~/stores/location'
import FileIcon from '../atom/FileIcon.vue'
import ResizeHandle from '../atom/ResizeHandle.vue'
import useFilesStore from '~/stores/files'
import { useWeblensAPI } from '~/api/AllApi'
import FileEventGroup from '../molecule/FileEventGroup.vue'
import type { EventGroup } from '../molecule/FileEventGroup.vue'
import { PortablePath } from '~/types/portablePath'
import ErrorCard from '../molecule/ErrorCard.vue'
import { WLError } from '~/types/wlError'

const locationStore = useLocationStore()
const filesStore = useFilesStore()

const isDragging = ref(false)
const widthAtDragStart = ref(0)

function onDragStart() {
    isDragging.value = true
    widthAtDragStart.value = locationStore.historyWidth
}

function onDrag(deltaX: number) {
    locationStore.setHistoryWidth(widthAtDragStart.value - deltaX)
}

const { data: historyData, error } = useAsyncData(
    'file-history-' + locationStore.activeFolderID + '-' + locationStore.isHistoryOpen,
    async () => {
        if (!locationStore.isHistoryOpen) {
            return []
        }

        let history = await useWeblensAPI()
            .FoldersAPI.getFolderHistory(locationStore.activeFolderID)
            .then((res) => {
                return res.data
            })
            .catch((err) => {
                throw new WLError(err)
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
        watch: [() => locationStore.activeFolderID, () => locationStore.isHistoryOpen, () => filesStore.files.length],
    },
)

const groupedHistory = computed((): EventGroup[] => {
    if (!historyData.value) return []

    const groups = new Map<string, EventGroup>()

    for (const action of historyData.value) {
        const existing = groups.get(action.eventID)
        if (existing) {
            existing.actions.push(action)
        } else {
            groups.set(action.eventID, {
                eventID: action.eventID,
                actions: [action],
                timestamp: action.timestamp,
                actionType: action.actionType,
            })
        }
    }

    return Array.from(groups.values())
})
</script>
