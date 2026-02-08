<template>
    <div
        :class="{
            'hidden h-10 w-max min-w-0 items-start justify-end gap-0.5 rounded transition-[background-color,width,height] sm:flex lg:w-full': true,
        }"
    >
        <WeblensOptions
            v-if="!locationStore.isInTimeline"
            v-model:value="fileShape as string"
            :options="shapeOptions"
            :class="{ 'mr-1': true }"
        />

        <WeblensOptions
            v-model:value="sortCondition as string"
            :options="sortOptions"
            merge="right"
        />

        <WeblensButton
            merge="row"
            :class="{ 'h-10 rounded-l-none!': true }"
            @click="toggleSortDirection"
        >
            <IconSortAscending v-if="filesStore.sortDirection === 'asc'" />
            <IconSortDescending v-if="filesStore.sortDirection === 'desc'" />
        </WeblensButton>
    </div>
</template>

<script setup lang="ts">
import {
    IconCalendar,
    IconFileAnalytics,
    IconLayoutColumns,
    IconLayoutGrid,
    IconLayoutRows,
    IconSortAscending,
    IconSortAZ,
    IconSortDescending,
    type Icon,
} from '@tabler/icons-vue'
import useFilesStore, { type FileShape, type SortCondition } from '~/stores/files'
import WeblensButton from '../atom/WeblensButton.vue'
import useLocationStore from '~/stores/location'
import WeblensOptions from '../atom/WeblensOptions.vue'

const filesStore = useFilesStore()
const locationStore = useLocationStore()

const fileShape = ref<FileShape | undefined>(filesStore.fileShape)
watch(fileShape, (newShape) => {
    if (newShape) filesStore.setFileShape(newShape)
})

const sortCondition = ref<SortCondition>(filesStore.sortCondition)
watch(sortCondition, (newSortCondition) => {
    filesStore.setSortCondition(newSortCondition)
})

const sortOptions: Record<SortCondition, { label: string; icon: Icon }> = {
    name: { label: 'Filename', icon: IconSortAZ },
    size: { label: 'Size', icon: IconFileAnalytics },
    updatedAt: { label: 'Date', icon: IconCalendar },
}

const shapeOptions = {
    square: { label: 'Grid', icon: IconLayoutGrid },
    row: { label: 'Rows', icon: IconLayoutRows },
    column: { label: 'Columns', icon: IconLayoutColumns, disabled: true },
}

function toggleSortDirection() {
    return filesStore.toggleSortDirection()
}
</script>
