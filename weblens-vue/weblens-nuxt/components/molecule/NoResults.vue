<template>
    <div
        v-if="locationStore.inShareRoot"
        :class="{ 'text-text-tertiary absolute flex h-full w-full items-center justify-center gap-1': true }"
    >
        <IconFileSad />
        <span :class="{ 'select-none': true }">No files shared with you</span>
    </div>

    <div
        v-else-if="locationStore.search !== ''"
        :class="{ 'text-text-tertiary absolute flex h-full w-full items-center justify-center gap-1': true }"
    >
        <IconSearch size="20" />
        <span :class="{ 'inline-flex justify-center gap-1 text-[16px] select-none': true }">
            No files found in
            <FileIcon
                :file="filesStore.activeFile"
                with-name
            />
            {{ filesStore.searchRecurively ? 'or its subfolders' : '' }}
            matching
            {{ filesStore.searchWithRegex ? 'regex' : '' }}
            "<strong> {{ locationStore.search }} </strong>"
        </span>
    </div>

    <div
        v-else-if="filesStore.activeFile?.IsTrash()"
        :class="{ 'text-text-tertiary absolute flex h-full w-full items-center justify-center gap-1': true }"
    >
        <IconTrashX />
        <span :class="{ 'select-none': true }">Someone already took the trash out</span>
    </div>

    <div
        v-else
        :class="{ 'text-text-tertiary absolute flex h-full w-full items-center justify-center gap-1': true }"
    >
        <IconFileSad />
        <span
            v-if="locationStore.isViewingPast"
            :class="{ 'select-none': true }"
        >
            This folder was empty
        </span>
        <span
            v-else
            :class="{ 'select-none': true }"
        >
            This folder is empty
        </span>
    </div>
</template>

<script setup lang="ts">
import { IconFileSad, IconSearch, IconTrashX } from '@tabler/icons-vue'
import useFilesStore from '~/stores/files'
import FileIcon from '../atom/FileIcon.vue'
import useLocationStore from '~/stores/location'

const filesStore = useFilesStore()
const locationStore = useLocationStore()
</script>
