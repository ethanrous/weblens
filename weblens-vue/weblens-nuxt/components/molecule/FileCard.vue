<template>
    <div
        :id="'file-card-' + file.ID()"
        ref="fileRef"
        :class="{
            'border-card-background-primary flex max-h-full rounded border transition select-none': true,
            'hover:bg-card-background-selected/90 bg-card-background-selected border-theme-primary': fileState.Has(
                SelectedState.Selected,
            ),
            'hover:bg-card-background-selected/50 hover:border-theme-primary hover:border': fileState.Has(
                SelectedState.Hovering,
            ),
            'bg-card-background-primary hover:bg-card-background-hover': !fileState.Any(
                SelectedState.Selected,
                SelectedState.InRange,
            ),
            'bg-card-background-hover': fileState.Has(SelectedState.InRange),
            'bg-card-background-disabled text-text-tertiary! pointer-events-none': fileState.Has(SelectedState.Moved),
            'aspect-square h-max w-full max-w-full': fileShape === 'square',
            'h-20 w-full max-w-full': fileShape === 'row',
        }"
        @dblclick.stop="navigateToFile"
        @contextmenu.stop.prevent="handleContextMenu"
        @click.stop="handleSelect"
        @mouseover.stop="filesStore.setNextSelectedIndex(fileIndex)"
        @mouseleave.stop="filesStore.clearNextSelectedIndex()"
        @mouseup="handleDrop"
        @mousedown="(e) => (downPos = { x: e.clientX, y: e.clientY })"
        @mousemove="
            (e) => {
                if (
                    (e.clientX - downPos.x > 5 || e.clientY - downPos.y > 5) &&
                    mousePressed.pressed.value &&
                    !filesStore.dragging
                ) {
                    filesStore.setDragging(true)
                    filesStore.setSelected(file.id, true)
                }
            }
        "
    >
        <component
            :is="fileComponent"
            :file="file"
            :file-state="fileState"
            @context-menu="handleContextMenu"
        >
            <template #file-visual>
                <MediaImage
                    v-if="file.contentID"
                    :class="{ 'animate-fade-in': true }"
                    :media="media"
                    :quality="PhotoQuality.LowRes"
                    no-click
                />
                <IconFolder
                    v-else-if="file.IsFolder()"
                    size="80%"
                    stroke="1"
                    :class="{ 'm-auto h-full': true }"
                />
            </template>
        </component>
    </div>
</template>

<script setup lang="ts">
import type WeblensFile from '@/types/weblensFile'
import FileSquare from '@/components/molecule/FileSquare.vue'
import MediaImage from '../atom/MediaImage.vue'
import { PhotoQuality } from '~/types/weblensMedia'
import type WeblensMedia from '~/types/weblensMedia'
import useFilesStore, { type FileShape } from '~/stores/files'
import { IconFolder } from '@tabler/icons-vue'
import { SelectedState } from '@/types/weblensFile'
import { useElementVisibility, useMousePressed } from '@vueuse/core'
import { moveFiles } from '~/api/FileBrowserApi'
import type { coordinates } from '~/types/style'
import FileRow from './FileRow.vue'
import useLocationStore from '~/stores/location'

const filesStore = useFilesStore()
const presentationStore = usePresentationStore()
const menuStore = useContextMenuStore()
const mediaStore = useMediaStore()
const locationStore = useLocationStore()

const downPos = ref<coordinates>({ x: 0, y: 0 })

const fileRef = ref<HTMLDivElement>()
const visible = useElementVisibility(fileRef, { rootMargin: '250px' })

const mousePressed = useMousePressed({ target: fileRef })
watchEffect(() => {
    if (!mousePressed.pressed.value) {
        filesStore.setDragging(false)
    }
})

const fileComponent = computed<typeof FileSquare | typeof FileRow>(() => {
    if (fileShape === 'square') {
        return FileSquare
    } else if (fileShape === 'row') {
        return FileRow
    }

    return FileSquare
})

const { file, fileIndex, fileShape } = defineProps<{
    file: WeblensFile
    fileIndex: number
    fileShape: FileShape
}>()

function navigateToFile() {
    if (!file.IsFolder()) {
        return presentationStore.setPresentationFileID(file.id)
    }

    return file.GoTo()
}

const isSelected = computed(() => {
    return filesStore.selectedFiles.has(file.id)
})

const isInSelectionRange = computed(() => {
    if (!filesStore.shiftPressed || filesStore.nextSelectedIndex === null || filesStore.lastSelectedIndex === -1) {
        return false
    }

    return filesStore.lastSelectedIndex < fileIndex !== filesStore.nextSelectedIndex < fileIndex
})

const media = ref<WeblensMedia>()

watchEffect(() => {
    if (media.value || !file.GetContentID()) {
        return
    }

    const m = mediaStore.mediaMap.get(file.GetContentID())
    if (!m && visible.value) {
        mediaStore.fetchSingleMedia(file.GetContentID())
    }

    media.value = m
})

const fileState = computed(() => {
    if (filesStore.movedFiles.has(file.id)) {
        return SelectedState.Moved
    } else if (isSelected.value) {
        return SelectedState.Selected
    } else if (filesStore.dragging && file.IsFolder() && !filesStore.selectedFiles.has(file.id)) {
        return SelectedState.Hovering
    } else if (isInSelectionRange.value) {
        return SelectedState.InRange
    }

    return SelectedState.NotSelected
})

function handleSelect(e: MouseEvent) {
    filesStore.setSelected(file.id, !isSelected.value, e.shiftKey)
}

function handleContextMenu(e: MouseEvent) {
    e.stopPropagation()
    menuStore.setTarget(file.ID())
    menuStore.setMenuOpen(true)

    const fbBox = document.getElementById('filebrowser-container')
    if (!fbBox) {
        console.error('File browser container not found')
        return
    }

    const rect = fbBox.getBoundingClientRect()

    menuStore.setMenuPosition({ x: e.pageX - rect.left, y: e.pageY - rect.top })
}

async function handleDrop(e: MouseEvent) {
    e.stopPropagation()
    if (filesStore.dragging && file.IsFolder() && !filesStore.selectedFiles.has(file.id)) {
        await moveFiles(file)
    }
    mousePressed.pressed.value = false
}

onMounted(() => {
    if (locationStore.highlightFileID === file.id) {
        fileRef.value?.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }
})
</script>
