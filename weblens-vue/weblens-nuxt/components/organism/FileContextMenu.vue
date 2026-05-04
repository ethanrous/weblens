<template>
    <div
        id="file-context-menu"
        ref="contextMenu"
        :class="{
            'file-context-menu animate-fade-in shadow-lg': true,
            'gone hidden': !menuStore.isOpen,
        }"
        :style="{ top: menuPosition.y + 'px', left: menuPosition.x + 'px' }"
    >
        <ContextMenuHeader
            :file="targetFile"
            :selected-files="selectedFiles"
            :naming-file="menuStore.menuMode"
        />

        <ContextMenuActions
            v-if="!menuStore.menuMode"
            :target-file="targetFile"
            :selected-files="selectedFiles"
            @create-folder="menuStore.setMenuMode('newName')"
            @rename-file="menuStore.setMenuMode('rename')"
            @share-file="menuStore.setSharing(true)"
            @tag-files="menuStore.setMenuMode('tags')"
            @set-cover="handleSetCoverClick"
        />

        <TagSelector
            v-else-if="menuStore.menuMode === 'tags'"
            :file-i-ds="selectedFiles"
        />

        <ContextNameFile
            v-else
            :value="menuStore.menuMode === 'rename' ? targetFile?.GetFilename() : ''"
            :disabled="!targetFile"
            @submit="handleNameFile"
        />
    </div>

    <ShareModal
        v-if="targetFile && menuStore.isSharing"
        :file="targetFile"
    />

    <MoveFilesModal v-if="selectedFiles && menuStore.menuMode === 'move'" />

    <FolderPickerModal
        v-if="showFolderPicker"
        :visible="showFolderPicker"
        :suggested-path="coverTargetFile ? (coverTargetFile.GetFilepath().parent ?? undefined) : undefined"
        @close="showFolderPicker = false"
        @select="handleSelectFolder"
    />
</template>

<script setup lang="ts">
import useFilesStore from '~/stores/files'
import ContextMenuHeader from '../molecule/ContextMenuHeader.vue'
import { onClickOutside, onKeyDown, useElementBounding, useElementSize } from '@vueuse/core'
import ContextMenuActions from '../molecule/ContextMenuActions.vue'
import ContextNameFile from '../molecule/ContextNameFile.vue'
import TagSelector from '../molecule/TagSelector.vue'
import useLocationStore from '~/stores/location'
import ShareModal from './ShareModal.vue'
import { useWeblensAPI } from '~/api/AllApi'
import type WeblensFile from '~/types/weblensFile'
import MoveFilesModal from './MoveFilesModal.vue'
import FolderPickerModal from './FolderPickerModal.vue'

const filesStore = useFilesStore()
const menuStore = useContextMenuStore()
const locationStore = useLocationStore()

const menuRef = useTemplateRef('contextMenu')
const menuSize = useElementSize(menuRef)
const container = shallowRef(document.getElementById('filebrowser-container'))
const containerBounds = useElementBounding(container)

onClickOutside(menuRef, (e) => {
    if (menuStore.isOpen) {
        e.stopPropagation()
        menuStore.setMenuOpen(false)
    }
})

onKeyDown(['Escape'], (e) => {
    e.stopPropagation()
    if (menuStore.isOpen) {
        menuStore.setMenuOpen(false)
    }
})

const menuPosition = computed(() => {
    const menuPos = {
        x: Math.min(
            menuStore.menuPosition.x,
            containerBounds.right.value - containerBounds.left.value - menuSize.width.value - 24,
        ),
        y: Math.min(
            menuStore.menuPosition.y,
            containerBounds.bottom.value - containerBounds.top.value - menuSize.height.value - 24,
        ),
    }

    return menuPos
})

const targetFile = ref<WeblensFile>()

watch([() => menuStore.directTargetID, () => filesStore.files, () => filesStore.activeFile], () => {
    targetFile.value = filesStore.getFileByID(menuStore.directTargetID)
})

const selectedFiles = computed(() => {
    const selectedFiles = filesStore.selectedFiles
    if (!selectedFiles.size) {
        return [menuStore.directTargetID]
    }

    if (!selectedFiles.has(menuStore.directTargetID)) {
        return [menuStore.directTargetID]
    }

    return [...selectedFiles]
})

async function handleNameFile(newName: string) {
    if (!targetFile.value) {
        console.warn('No active file to rename or create folder for.')
        return
    }

    if (menuStore.menuMode === 'rename') {
        const updateFileRequest = { newName }
        await useWeblensAPI().FilesAPI.updateFile(targetFile.value.id, updateFileRequest, locationStore.activeShareID)
    } else if (menuStore.menuMode === 'newName') {
        const newFolderRequest = { newFolderName: newName, parentFolderID: targetFile.value.ID() }
        await useWeblensAPI().FoldersAPI.createFolder(newFolderRequest)
    }

    menuStore.setMenuMode()
    menuStore.setMenuOpen(false)
}

onMounted(() => {
    container.value = document.getElementById('filebrowser-container')
})

const showFolderPicker = ref(false)
const coverTargetFile = ref<WeblensFile>()

function handleSetCoverClick() {
    coverTargetFile.value = targetFile.value
    menuStore.setMenuOpen(false)
    showFolderPicker.value = true
}

async function handleSelectFolder(folderID: string) {
    if (!coverTargetFile.value || !folderID) return

    try {
        await useWeblensAPI().FoldersAPI.setFolderCover(folderID, coverTargetFile.value.contentID)
    } catch (error) {
        console.error('Error setting cover:', error)
    } finally {
        showFolderPicker.value = false
        coverTargetFile.value = undefined
    }
}
</script>

<style>
@reference '../../assets/css/base.css';

.file-context-menu {
    background-color: var(--color-background-primary);
    position: absolute;
    height: max-content;
    width: max-content;
    border-radius: 0.375rem; /* rounded */
    border: 1px solid var(--color-border-primary);
    padding: calc(var(--spacing) * 2);
    transition: height 0.1s var(--ease-wl-default);
    transition-property: height, top, left, opacity;
    z-index: 90;
    isolation: isolate;

    @apply shadow-sm;
}
</style>
