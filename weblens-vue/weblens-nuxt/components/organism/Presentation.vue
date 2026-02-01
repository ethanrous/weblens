<template>
    <Teleport to="body">
        <div
            v-if="media"
            ref="presentation"
            :class="{
                'presentation fullscreen-modal flex-col justify-end sm:flex-row sm:justify-around': true,
            }"
            @click.stop="presentationStore.clearPresentation"
        >
            <div :class="{ 'relative flex h-full w-full': true }">
                <VideoPlayer
                    v-if="media.IsVideo()"
                    :media="media"
                />

                <PDF
                    v-else-if="media.IsPdf()"
                    :media="media"
                />

                <MediaImage
                    v-else-if="media"
                    :media="media"
                    :quality="PhotoQuality.HighRes"
                    :class="{ 'min-w-0 p-2 max-sm:my-auto lg:p-0': true }"
                    :style="{
                        height: `calc(${presentationSize.height.value}px - 1rem)`,
                        maxHeight: `calc(${presentationSize.height.value}px - 1rem)`,
                        maxWidth: presentationSize.width.value + 'px',
                    }"
                    :contain="true"
                    no-click
                />

                <div
                    v-else-if="presentingFile?.IsFolder()"
                    :class="{
                        'flex h-full w-max min-w-max items-center justify-center gap-4': true,
                    }"
                >
                    <IconFolder
                        size="10%"
                        stroke="1"
                    />
                    <h1>{{ presentingFile.GetFilename() }}</h1>
                </div>

                <div
                    :class="{
                        'absolute flex h-full max-w-full shrink-0 flex-col items-center justify-center overflow-hidden transition-[width,height,margin] duration-300 lg:relative lg:mb-0 lg:ml-4': true,
                        'w-full p-4 backdrop-blur-xs lg:w-1/3 lg:p-0 lg:backdrop-blur-none': fileInfoOpen,
                        'pointer-events-none opacity-0 lg:w-0 lg:opacity-100': !fileInfoOpen,
                    }"
                >
                    <h3 v-if="presentingFile">{{ presentingFile?.GetFilename() }}</h3>
                    <CopyBox
                        v-if="media"
                        :class="{
                            'relative mb-2 w-full min-w-0 overflow-x-auto': true,
                        }"
                        :text="media.MediaUrl()"
                    >
                        <IconPhoto
                            size="20"
                            :class="{ 'shrink-0': true }"
                        />
                    </CopyBox>

                    <Mapbox
                        v-if="media.location && media.location[0] !== 0"
                        :coords="media.location"
                        :class="{ 'h-98 w-full': true }"
                    />
                </div>

                <IconInfoCircle
                    :class="{
                        'absolute top-4 right-4 shrink-0 cursor-pointer rounded p-0.5 transition': true,
                        'bg-card-background-primary/50 text-text-primary': fileInfoOpen,
                        'text-text-secondary': !fileInfoOpen,
                    }"
                    size="20"
                    @click.stop="fileInfoOpen = !fileInfoOpen"
                />
                <IconArrowLeft
                    :class="{
                        'bg-card-background-primary/50 absolute bottom-10 left-10 m-2 rounded p-1 sm:hidden': true,
                    }"
                    size="32"
                    @click.stop="movePresentation(-1)"
                />
                <IconArrowRight
                    :class="{
                        'bg-card-background-primary/50 absolute right-10 bottom-10 m-2 rounded p-1 sm:hidden': true,
                    }"
                    size="32"
                    @click.stop="movePresentation(1)"
                />
            </div>
        </div>
    </Teleport>
</template>

<script setup lang="ts">
import useFilesStore from '~/stores/files'
import MediaImage from '../atom/MediaImage.vue'
import WeblensMedia, { PhotoQuality } from '~/types/weblensMedia'
import { onKeyStroke, onKeyUp, useElementSize } from '@vueuse/core'
import { IconArrowLeft, IconArrowRight, IconFolder, IconInfoCircle, IconPhoto } from '@tabler/icons-vue'
import CopyBox from '../molecule/CopyBox.vue'
import VideoPlayer from '../molecule/VideoPlayer.vue'
import PDF from '../atom/PDF.vue'
import Mapbox from '../atom/Mapbox.vue'

const presentationStore = usePresentationStore()
const filesStore = useFilesStore()
const mediaStore = useMediaStore()
const fileInfoOpen = ref<boolean>(false)

const presentation = ref<HTMLDivElement>()
const presentationSize = useElementSize(presentation)

const presentingFile = computed(() => {
    return filesStore.children?.find((f) => f.id === presentationStore.presentationFileID)
})

const media = computed(() => {
    const contentID = presentationStore.presentationMediaID
        ? presentationStore.presentationMediaID
        : presentingFile.value?.GetContentID()
    if (!contentID) {
        return
    }

    return mediaStore.media.get(contentID) ?? new WeblensMedia({ contentID: contentID, mimeType: 'application/pdf' })
})

onKeyUp(['Escape'], (e) => {
    e.stopPropagation()
    presentationStore.clearPresentation()
})

onKeyUp(['i'], (e) => {
    e.stopPropagation()
    fileInfoOpen.value = !fileInfoOpen.value
})

function movePresentation(direction: number) {
    if (presentationStore.onMovePresentation) {
        presentationStore.onMovePresentation(direction)
        return
    }

    if (!filesStore.children) {
        console.warn('No children to move to')
        return
    }

    const presentingIndex = filesStore.children.findIndex((f) => f.id === presentationStore.presentationFileID)
    if (presentingIndex === -1) {
        console.warn('No presentingIndex', presentationStore.presentationFileID, filesStore.children)
        return
    }

    if (
        (presentingIndex === 0 && direction === -1) ||
        (presentingIndex === filesStore.children.length - 1 && direction === 1)
    ) {
        return
    }

    const newID = filesStore.children[presentingIndex + direction]?.id
    if (!newID) {
        console.error('No newID', presentingIndex, direction, filesStore.children)
        return
    }

    presentationStore.setPresentationFileID(newID)
}

onKeyStroke(['ArrowRight'], (e) => {
    if (media.value) {
        e.preventDefault()
        e.stopPropagation()
        movePresentation(1)
    }
})

onKeyStroke(['ArrowLeft'], (e) => {
    if (media.value) {
        e.preventDefault()
        e.stopPropagation()
        movePresentation(-1)
    }
})
</script>
