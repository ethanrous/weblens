<template>
    <div
        v-if="file"
        class="mb-1.5 flex items-center gap-1 border-b px-2 pb-1.5 select-none"
    >
        <FileIcon
            v-if="!namingFile"
            :file="file"
        />
        <h5>{{ label }}</h5>
        <FileIcon
            v-if="namingFile === 'rename'"
            :file="file"
            with-name
        />
    </div>
</template>

<script setup lang="ts">
import type WeblensFile from '~/types/weblensFile'
import FileIcon from '../atom/FileIcon.vue'

const props = defineProps<{
    file?: WeblensFile
    selectedFiles?: string[]
    namingFile?: 'rename' | 'newName'
}>()

const label = computed(() => {
    if (props.namingFile === 'rename') {
        return 'Rename '
    }

    if (props.namingFile === 'newName') {
        return 'New folder'
    }

    if (props.selectedFiles && props.selectedFiles.length > 1) {
        return `Selected ${props.selectedFiles.length} file${props.selectedFiles.length > 1 ? 's' : ''}`
    }

    return props.file?.GetFilename() ?? ''
})
</script>
