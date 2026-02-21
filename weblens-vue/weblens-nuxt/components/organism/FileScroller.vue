<template>
    <div
        id="file-scroller"
        :class="{ 'page-root relative': true }"
    >
        <FileContextMenu />

        <span
            v-if="!filesStore.searchUpToDate && filesStore.searchRecurively && locationStore.search !== ''"
            :class="{ 'text-text-secondary m-auto inline-flex items-center gap-1': true }"
        >
            Press
            <span :class="{ 'rounded border p-0.5': true }"> ENTER </span>
            to search
        </span>
        <div
            v-else
            ref="scrollerContainer"
            :class="{
                'relative flex h-full w-full flex-col': true,
                'after:bg-theme-secondary/50 after:border-theme-primary after:pointer-events-none after:absolute after:inset-0 after:h-full after:w-full after:border':
                    hovering,
            }"
            @mousemove="
                () => {
                    if (hovering == true) {
                        hovering = false
                    }
                }
            "
            @drop.stop.prevent="handleDrop"
            @dragover.stop="handleDragOver"
            @dragleave.stop="handleDragLeave"
            @click.stop="handleClick"
            @contextmenu.stop.prevent="handleContextMenu"
        >
            <Loader
                v-if="filesStore.loading || filesStore.status === 'pending'"
                :class="{ 'm-auto': true }"
            />

            <ErrorCard
                v-else-if="filesStore.fileFetchError"
                :error="filesStore.fileFetchError"
            />
            <NoResults v-else-if="files.length === 0" />

            <div
                v-else
                id="file-scroller"
                :class="{
                    'file-scroller': true,
                    'grid-cols-1!': filesStore.fileShape !== 'square',
                }"
                :style="{ maxWidth: maxW }"
            >
                <template
                    v-for="(file, index) of files"
                    :key="file.id"
                >
                    <FileCard
                        v-if="noRequireParentMatch || file.parentID === locationStore.activeFolderID"
                        :file="file"
                        :file-index="index"
                        :file-shape="filesStore.fileShape"
                    />
                </template>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import type WeblensFile from '~/types/weblensFile'
import FileCard from '../molecule/FileCard.vue'
import { isParent } from '~/util/domHelpers'
import { HandleDrop } from '~/api/uploadApi'
import useFilesStore from '~/stores/files'
import { onKeyDown, useActiveElement, useMagicKeys } from '@vueuse/core'
import FileContextMenu from './FileContextMenu.vue'
import Loader from '../atom/Loader.vue'
import useLocationStore from '~/stores/location'
import NoResults from '../molecule/NoResults.vue'
import ErrorCard from '../molecule/ErrorCard.vue'

const filesStore = useFilesStore()
const locationStore = useLocationStore()
const menuStore = useContextMenuStore()
const presentationStore = usePresentationStore()

const scrollerContainer = useTemplateRef('scrollerContainer')
const hovering = ref(false)

const { noRequireParentMatch = true } = defineProps<{
    files: WeblensFile[]
    noRequireParentMatch?: boolean
}>()

const activeElement = useActiveElement()
const notUsingInput = computed(
    () => activeElement.value?.tagName !== 'INPUT' && activeElement.value?.tagName !== 'TEXTAREA',
)

const maxW = computed(() => {
    if (filesStore.fileShape === 'square') {
        return `min(100%, calc(var(--spacing) * 80 * ${filesStore.files.length > 0 ? filesStore.files.length : 1} - var(--spacing) * 2))`
    }

    return ''
})

onKeyDown(
    ['Escape'],
    (e) => {
        if (menuStore.isSharing) {
            menuStore.setSharing(false)
            return
        }

        if (presentationStore.presentationFileID) {
            return
        }

        e.stopPropagation()
        filesStore.clearSelected()
    },
    { dedupe: true },
)

const { Ctrl_A, Cmd_A, space, shift } = useMagicKeys({
    passive: false,
    onEventFired: (e) => {
        if ((((e.ctrlKey || e.metaKey) && e.key === 'a') || /* Spacebar */ e.key === ' ') && notUsingInput.value) {
            e.preventDefault()
        }
    },
})

watch([Ctrl_A, Cmd_A], () => {
    if ((Ctrl_A?.value || Cmd_A?.value) && notUsingInput.value) {
        filesStore.selectAll()
    }
})

watch(
    () => space?.value,
    (isPressed) => {
        if (isPressed && filesStore.lastSelected && presentationStore.presentationFileID === '') {
            presentationStore.setPresentationFileID(filesStore.lastSelected)
        } else if (isPressed && filesStore.lastSelected && presentationStore.presentationFileID) {
            presentationStore.clearPresentation()
        }
    },
)

watch(shift!, filesStore.setShiftPressed)

onKeyDown(
    ['Escape'],
    (e) => {
        if (usePresentationStore().presentationFileID) {
            return
        }

        e.stopPropagation()
        filesStore.clearSelected()
    },
    { dedupe: true },
)

function handleClick() {
    filesStore.clearSelected()
}

function handleContextMenu(e: MouseEvent) {
    menuStore.setTarget(locationStore.activeFolderID)
    menuStore.setMenuOpen(true)
    menuStore.setMenuPosition({ x: e.offsetX, y: e.offsetY })
}

function handleDragOver(event: DragEvent) {
    event.preventDefault()
    event.stopPropagation()

    hovering.value = true
}

function handleDragLeave(event: DragEvent) {
    event.preventDefault()
    event.stopPropagation()
    if (!hovering.value) {
        return
    }

    if (!event.relatedTarget) {
        hovering.value = false
        return
    }

    if (!scrollerContainer.value || !event.relatedTarget) {
        return
    }

    if (
        scrollerContainer.value === event.relatedTarget ||
        isParent(scrollerContainer.value, event.relatedTarget as Element)
    ) {
        return
    }

    hovering.value = false
}

function handleDrop(event: DragEvent) {
    event.preventDefault()
    event.stopPropagation()
    hovering.value = false

    if (!event.dataTransfer) {
        console.warn('No files to drop')
        return
    }

    HandleDrop(event.dataTransfer.items, locationStore.activeFolderID, false, '')
}
</script>

<style scoped>
.file-scroller {
    grid-template-columns: repeat(auto-fit, minmax(calc(var(--spacing) * 60), 1fr));

    @media (max-width: 1250px) {
        grid-template-columns: repeat(auto-fit, minmax(calc(var(--spacing) * 40), 1fr));
    }

    position: relative;
    display: grid;
    height: max-content;
    width: 100%;

    gap: 0.5rem;
    overflow-y: auto;
    padding: 0.5rem;
}
</style>
