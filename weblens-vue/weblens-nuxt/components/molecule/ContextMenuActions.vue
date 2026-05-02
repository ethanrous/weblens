<template>
    <div :class="{ 'flex min-w-40 flex-col gap-1.5': true }">
        <WeblensButton
            v-if="targetIsActiveFolder"
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
            :disabled="!targetFile?.CanEdit() || protectedFile"
            @click.stop="emit('renameFile')"
        >
            <IconPencil />
        </WeblensButton>

        <WeblensButton
            v-if="!multipleSelected"
            label="Info"
            fill-width
            :disabled="!targetFile"
            @click.stop="
                () => {
                    usePresentationStore().setPresentationFileID(targetFile!.ID())
                    menuStore.setMenuOpen(false)
                }
            "
        >
            <IconInfoCircle />
        </WeblensButton>

        <WeblensButton
            v-if="filesStore.isSearching && targetFile && !multipleSelected"
            label="Jump to File"
            fill-width
            :disabled="!targetFile"
            @click.stop="goToFile"
        >
            <IconSearch />
        </WeblensButton>

        <WeblensButton
            label="Tags"
            fill-width
            :disabled="!targetFile || !canModifyTarget"
            @click.stop="emit('tagFiles')"
        >
            <IconTag />
        </WeblensButton>

        <WeblensButton
            v-if="!multipleSelected"
            :label="targetFile?.shareID ? 'Edit Share' : 'Share'"
            fill-width
            :disabled="!canModifyTarget || protectedFile"
            @click.stop="emit('shareFile')"
        >
            <IconUsersPlus />
        </WeblensButton>

        <WeblensButton
            v-if="canSetAsCover"
            label="Set as Cover"
            fill-width
            :disabled="!canSetAsCover"
            @click.stop="emit('setCover')"
        >
            <IconPhoto />
        </WeblensButton>

        <WeblensButton
            v-if="targetIsActiveFolder"
            label="Scan Folder"
            fill-width
            :disabled="!canModifyTarget || websocketStore.status !== 'OPEN'"
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
            @click.stop="doDownload"
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
            v-if="locationStore.isViewingPast"
            label="Restore"
            fill-width
            :disabled="!targetFile || isRestoring"
            @click.stop="handleRestore"
        >
            <IconRestore />
        </WeblensButton>

        <WeblensButton
            v-if="targetIsActiveFolder"
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
            v-if="canRemoveCover"
            label="Remove Cover"
            fill-width
            flavor="danger"
            :disabled="!canRemoveCover || isRemovingCover"
            @click.stop="handleRemoveCover"
        >
            <IconPhotoOff />
        </WeblensButton>

        <WeblensButton
            :label="deleteText"
            fill-width
            flavor="danger"
            :disabled="!targetFile?.CanDelete() && !locationStore.isInTrash"
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
    IconInfoCircle,
    IconPencil,
    IconPhoto,
    IconPhotoOff,
    IconPhotoScan,
    IconRestore,
    IconSearch,
    IconTag,
    IconTrash,
    IconUsersPlus,
} from '@tabler/icons-vue'
import WeblensButton from '../atom/WeblensButton.vue'
import type WeblensFile from '~/types/weblensFile'
import useFilesStore from '~/stores/files'
import { handleDownload, ScanDirectory } from '~/api/FileBrowserApi'
import { useWeblensAPI } from '~/api/AllApi'
import useLocationStore from '~/stores/location'
import ProgressSquare from '../atom/ProgressSquare.vue'
import useWebsocketStore from '~/stores/websocket'

const filesStore = useFilesStore()
const locationStore = useLocationStore()
const menuStore = useContextMenuStore()
const tasksStore = useTasksStore()
const userStore = useUserStore()
const websocketStore = useWebsocketStore()

const downloadTaskID = ref<string>()
const isRemovingCover = ref(false)

const emit = defineEmits<{
    (e: 'createFolder' | 'renameFile' | 'shareFile' | 'tagFiles' | 'setCover'): void
}>()

async function handleRemoveCover() {
    if (!props.targetFile) return

    isRemovingCover.value = true

    try {
        await useWeblensAPI().FoldersAPI.removeFolderCover(props.targetFile.id)
        menuStore.setMenuOpen(false)
    } catch (error) {
        console.error('Error removing cover:', error)
    } finally {
        isRemovingCover.value = false
    }
}

const props = defineProps<{
    targetFile?: WeblensFile
    selectedFiles?: string[]
}>()

const targetIsActiveFolder = computed(() => {
    return (
        filesStore.activeFile && filesStore.activeFile.IsFolder() && filesStore.activeFile.id === props.targetFile?.id
    )
})

const protectedFile = computed(() => {
    return props.targetFile?.IsHome() || props.targetFile?.IsTrash()
})

const canModifyTarget = computed(() => {
    return props.targetFile?.modifiable && !locationStore.isViewingPast && props.targetFile?.CanEdit()
})

const canModifyParent = computed(() => {
    return filesStore.activeFile?.modifiable && !locationStore.isViewingPast
})

const canSetAsCover = computed(() => {
    if (multipleSelected.value || !canModifyTarget.value) {
        return false
    }

    if (!props.targetFile || props.targetFile.IsFolder()) {
        return false
    }

    return props.targetFile.hasMedia && props.targetFile.contentID !== ''
})

const canRemoveCover = computed(() => {
    if (multipleSelected.value || !canModifyTarget.value) {
        return false
    }

    if (!props.targetFile || !props.targetFile.IsFolder() || !props.targetFile.contentID) {
        return false
    }

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

async function goToFile() {
    if (!props.targetFile) {
        console.warn('No target file to jump to')
        return
    }

    await props.targetFile.GoTo()
    filesStore.clearSearch()
    menuStore.setMenuOpen(false)
}

async function doDownload() {
    let files: WeblensFile[] = []
    if (props.selectedFiles && props.selectedFiles.length > 0) {
        files = props.selectedFiles.map((fID) => filesStore.getFileByID(fID)).filter((f) => !!f)
    } else if (props.targetFile) {
        files = [props.targetFile]
    } else {
        console.warn('No target file to download')

        return
    }

    const downloadInfo = await handleDownload(files)

    if (!downloadInfo) {
        return
    }

    downloadTaskID.value = downloadInfo.zipTaskID
    await downloadInfo.downloadPromise.finally(() => {
        menuStore.setMenuOpen(false)
        downloadTaskID.value = undefined
    })
}

watch([() => props.targetFile, () => props.selectedFiles], () => {
    downloadTaskID.value = undefined
})

const isRestoring = ref(false)

async function handleRestore(): Promise<void> {
    if (!props.targetFile || !props.selectedFiles) {
        return
    }

    isRestoring.value = true

    try {
        await useWeblensAPI().FilesAPI.restoreFiles({
            fileIDs: props.selectedFiles,
            newParentID: filesStore.activeFile?.id ?? '',
            timestamp: locationStore.viewTimestamp,
        })

        // Exit past view and navigate to the parent folder
        locationStore.setViewTimestamp(0)
        menuStore.setMenuOpen(false)
    } catch (error) {
        console.error('Error restoring files:', error)
    } finally {
        isRestoring.value = false
    }
}

async function handleDeleteFile(): Promise<void> {
    if (!props.targetFile) {
        console.warn('No target file to delete')
        return
    }

    if (!props.selectedFiles) {
        console.warn('No selected files to delete')
        return
    }

    filesStore.setMovedFile(...props.selectedFiles)
    if (props.targetFile?.IsTrash()) {
        await useWeblensAPI().FilesAPI.deleteFiles({ fileIDs: [userStore.user.trashID] }, false, true)
    } else if (locationStore.isInTrash) {
        await useWeblensAPI().FilesAPI.deleteFiles({ fileIDs: props.selectedFiles })
    } else {
        await useWeblensAPI().FilesAPI.moveFiles({ fileIDs: props.selectedFiles, newParentID: userStore.user.trashID })
    }

    menuStore.setMenuOpen(false)
}
</script>
