<template>
    <div :class="{ 'flex min-w-40 flex-col gap-1.5': true }">
        <WeblensButton
            v-if="targetIsFolder"
            label="New Folder"
            fill-width
            :disabled="!canModifyParent"
            @click.stop="emit('createFolder')"
        >
            <IconFolderPlus />
        </WeblensButton>

        <WeblensButton
            v-if="!multipleSelected"
            label="Rename"
            fill-width
            :disabled="!targetFile?.CanEdit()"
            @click.stop="emit('renameFile')"
        >
            <IconPencil />
        </WeblensButton>

        <WeblensButton
            v-if="!multipleSelected"
            label="Share"
            fill-width
            :disabled="!canModifyTarget || protectedFile"
            @click.stop="emit('shareFile')"
        >
            <IconUsersPlus />
        </WeblensButton>

        <WeblensButton
            v-if="targetIsFolder"
            label="Scan Folder"
            fill-width
            :disabled="!canModifyTarget"
            @click.stop="handleScan"
        >
            <IconPhotoScan />
        </WeblensButton>

        <WeblensButton
            v-if="!locationStore.isInTrash"
            :key="targetFile?.ID()"
            :label="downloadTaskPercentComplete ? `Zipping (${downloadTaskPercentComplete.toFixed(0)}%)` : 'Download'"
            fill-width
            :class="{
                'relative overflow-hidden': true,
                'rounded-b-xs': downloadTaskPercentComplete !== undefined,
            }"
            :disabled="!targetFile?.CanDownload()"
            @click.stop="handleDownload"
        >
            <IconDownload />
            <ProgressSquare
                v-if="downloadTaskPercentComplete"
                :class="{ 'bg-background-secondary! absolute! bottom-0 left-[-1%] z-40 h-1 w-[102%]': true }"
                :show-outline="false"
                :progress="downloadTaskPercentComplete"
            />
        </WeblensButton>

        <WeblensButton
            label="Folder History"
            fill-width
            @click.stop="
                () => {
                    locationStore.setHistoryOpen(true)
                    menuStore.setMenuOpen(false)
                }
            "
        >
            <IconHistoryToggle />
        </WeblensButton>

        <WeblensButton
            :label="deleteText"
            fill-width
            flavor="danger"
            :disabled="!targetFile?.CanDelete()"
            @click.stop="handleDeleteFile"
        >
            <IconTrash />
        </WeblensButton>
    </div>
</template>

<script setup lang="ts">
import {
    IconDownload,
    IconFolderPlus,
    IconHistoryToggle,
    IconPencil,
    IconPhotoScan,
    IconTrash,
    IconUsersPlus,
} from '@tabler/icons-vue'
import WeblensButton from '../atom/WeblensButton.vue'
import type WeblensFile from '~/types/weblensFile'
import useFilesStore from '~/stores/files'
import { downloadManyFiles, downloadSingleFile, ScanDirectory } from '~/api/FileBrowserApi'
import { useWeblensAPI } from '~/api/AllApi'
import useLocationStore from '~/stores/location'
import ProgressSquare from '../atom/ProgressSquare.vue'

const filesStore = useFilesStore()
const locationStore = useLocationStore()
const menuStore = useContextMenuStore()
const tasksStore = useTasksStore()
const userStore = useUserStore()

const downloadTaskID = ref<string>()

const emit = defineEmits<{
    (e: 'createFolder' | 'renameFile' | 'shareFile'): void
}>()

const props = defineProps<{
    targetFile?: WeblensFile
    selectedFiles?: string[]
}>()

const targetIsFolder = computed(() => {
    return (
        filesStore.activeFile && filesStore.activeFile.IsFolder() && filesStore.activeFile.id === props.targetFile?.id
    )
})

const protectedFile = computed(() => {
    return props.targetFile?.IsHome() || props.targetFile?.IsTrash()
})

const canModifyTarget = computed(() => {
    return props.targetFile?.modifiable && !locationStore.isViewingPast
})

const canModifyParent = computed(() => {
    return filesStore.activeFile?.modifiable && !locationStore.isViewingPast
})

const canDelete = computed(() => {
    if (props.targetFile?.IsTrash()) return true

    if (locationStore.isInTrash) return true

    if (!canModifyTarget.value) return false

    if (targetIsFolder.value) return false

    if (locationStore.activeShare && !locationStore.activeShare.checkPermission('canDelete')) return false

    if (protectedFile.value) return false

    return true
})

const multipleSelected = computed(() => {
    return props.selectedFiles && props.selectedFiles.length > 1
})

const downloadTaskPercentComplete = computed(() => {
    if (!downloadTaskID.value) {
        return undefined
    }

    return tasksStore.tasks?.get(downloadTaskID.value)?.percentComplete
})

const deleteText = computed(() => {
    if (props.targetFile?.IsTrash()) {
        return 'Empty Trash'
    } else if (locationStore.isInTrash) {
        return 'Delete'
    }

    return 'Trash'
})

function handleScan() {
    if (!props.targetFile) {
        console.warn('No target file to scan')
        return
    }

    ScanDirectory(props.targetFile)
}

async function handleDownload() {
    if (!props.targetFile) {
        console.warn('No target file to scan')
        return
    }

    if (!multipleSelected.value && !props.targetFile.IsFolder()) {
        await downloadSingleFile(props.targetFile?.ID(), props.targetFile?.GetFilename())
            .then(() => {
                menuStore.setMenuOpen(false)
            })
            .catch((error) => {
                console.error('Error downloading file:', error)
            })
        return
    } else if (props.selectedFiles) {
        const takeoutRes = await downloadManyFiles(props.selectedFiles)

        if (takeoutRes.taskID) {
            downloadTaskID.value = takeoutRes.taskID
        }

        const takeoutInfo = await takeoutRes.takeoutInfo

        if (!takeoutInfo.takeoutID || !takeoutInfo.filename) {
            console.warn('Missing takeoutID or filename returned from downloadManyFiles', takeoutInfo)
            return
        }

        await downloadSingleFile(takeoutInfo.takeoutID, takeoutInfo.filename, 'zip')
            .then(() => {
                menuStore.setMenuOpen(false)
            })
            .catch((error) => {
                console.error('Error downloading file:', error)
            })
            .finally(() => {
                downloadTaskID.value = undefined
            })
    }
}

watch([() => props.targetFile, () => props.selectedFiles], () => {
    downloadTaskID.value = undefined
})

async function handleDeleteFile(): Promise<void> {
    if (!props.targetFile) {
        console.warn('No target file to delete')
        return
    }

    if (!props.selectedFiles) {
        console.warn('No selected files to delete')
        return
    }

    filesStore.setMovedFile(props.selectedFiles, true)
    if (props.targetFile?.IsTrash()) {
        await useWeblensAPI().FilesAPI.deleteFiles({ fileIDs: [userStore.user.trashID] }, false, true)
    } else if (locationStore.activeFolderID === userStore.user.trashID) {
        await useWeblensAPI().FilesAPI.deleteFiles({ fileIDs: props.selectedFiles })
    } else {
        await useWeblensAPI().FilesAPI.moveFiles({ fileIDs: props.selectedFiles, newParentID: userStore.user.trashID })
    }

    menuStore.setMenuOpen(false)
}
</script>
