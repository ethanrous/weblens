<template>
    <div
        ref="timelineContainer"
        :class="{
            'timelineContainer page-root relative flex max-w-full flex-col flex-nowrap overflow-x-hidden overflow-y-auto pt-2': true,
        }"
    >
        <ErrorCard
            v-if="mediaStore.timelineFetchError"
            :error="mediaStore.timelineFetchError"
        />

        <div
            v-else-if="rows.rows.length === 0 && !loading && mediaStore.searchUpToDate"
            :class="{ 'm-auto flex flex-col items-center gap-3': true }"
        >
            <h3 :class="{ 'text-text-secondary text-lg': true }">No media found</h3>
            <p :class="{ 'text-text-tertiary text-sm': true }">Try adjusting your filters</p>
            <WeblensButton
                label="Return to Files"
                @click="useLocationStore().isInTimeline = false"
            />
        </div>

        <span
            v-else-if="!mediaStore.searchUpToDate"
            :class="{ 'text-text-secondary m-auto inline-flex items-center gap-1': true }"
        >
            Press
            <span :class="{ 'rounded border p-0.5': true }"> ENTER </span>
            to search
        </span>

        <template v-else>
            <div
                v-for="(row, rowIndex) of rows.rows"
                :key="String(rowIndex) + row.items.length"
                :class="{
                    'mx-2 flex flex-row': true,
                }"
                :style="{
                    marginTop: MARGIN_SIZE / 2 + 'px',
                    marginBottom: MARGIN_SIZE / 2 + 'px',
                    height: row.rowHeight + 'px',
                    maxHeight: row.rowHeight + 'px',
                    width: row.rowWidth + 'px',
                    flexShrink: 0,
                }"
            >
                <UseElementVisibility
                    v-slot="{ isVisible }"
                    :class="{ 'flex h-full w-full': true }"
                >
                    <div
                        v-for="(media, colIndex) of row.items"
                        :id="media.m.contentID"
                        :key="media.m.contentID + rowIndex + isVisible"
                        :class="{ 'flex items-center justify-center': true }"
                        :style="{
                            marginLeft: MARGIN_SIZE / 2 + 'px',
                            marginRight: MARGIN_SIZE / 2 + 'px',
                            width: media.w + 'px',
                            height: row.rowHeight + 'px',
                        }"
                        @click.stop="startPresenting(rowIndex, colIndex)"
                    >
                        <MediaImage
                            :media="media.m"
                            placeholder
                            :class="{
                                'hover:border-text-primary border-text-primary/0 h-full max-h-full w-full max-w-full shrink-0 cursor-pointer overflow-hidden rounded-lg border transition-[scale,border,shadow] hover:shadow': true,
                            }"
                        />
                    </div>
                </UseElementVisibility>
            </div>
        </template>
        <Loader
            v-if="mediaStore.canLoadMore"
            :class="{ 'm-auto': true }"
        />
        <div
            ref="bottomSpacer"
            :class="{ 'w-full shrink-0': true }"
            :style="{ height: `${rows.remainingGap}px` }"
        />
    </div>
</template>

<script setup lang="ts">
import { onKeyPressed, useDebounce, useElementSize, useElementVisibility } from '@vueuse/core'
import { UseElementVisibility } from '@vueuse/components'

import MediaImage from '../atom/MediaImage.vue'
import { GetMediaRows } from '~/types/weblensMedia'
import ErrorCard from '../molecule/ErrorCard.vue'
import Loader from '../atom/Loader.vue'
import WeblensButton from '../atom/WeblensButton.vue'
import useLocationStore from '~/stores/location'

const mediaStore = useMediaStore()

const timelineContainer = ref<HTMLDivElement>()
const bottomSpacer = ref<HTMLDivElement>()
const timelineSize = useElementSize(timelineContainer)
const presentationStore = usePresentationStore()
const presentationIndex = ref<number>(-1)

const timelineWidthBounced = useDebounce(timelineSize.width, 100)

const MARGIN_SIZE = 4

const computedSizesLoading = ref<boolean>(true)

const rows = computed(() => {
    if (timelineWidthBounced.value <= 0) {
        return { rows: [], remainingGap: 0 }
    }

    const rows = GetMediaRows(
        mediaStore.timelineMedia,
        mediaStore.timelineImageSize,
        timelineWidthBounced.value - 8,
        MARGIN_SIZE,
        mediaStore.canLoadMore ? mediaStore.totalMedias : mediaStore.timelineMedia.length,
    )

    computedSizesLoading.value = false
    return rows
})

const loading = computed(() => {
    return mediaStore.timelineLoading || computedSizesLoading.value
})

const visible = useElementVisibility(bottomSpacer, {
    scrollTarget: timelineContainer,
    rootMargin: '0px 0px 1000px 0px',
})

watchEffect(async () => {
    if (visible.value) {
        mediaStore.fetchMoreMedia()
    }
})

function startPresenting(rowIndex: number, colIndex: number) {
    const absIndex = rows.value.rows.slice(0, rowIndex).reduce((acc, row) => acc + row.items.length, 0) + colIndex
    presentationIndex.value = absIndex
    presentationStore.setPresentationMediaID(mediaStore.timelineMedia[absIndex]?.contentID ?? '')
}

onKeyPressed(['=', '-'], (e) => {
    mediaStore.updateImageSize(e.key === '=' ? 'increase' : 'decrease')
})

onMounted(() => {
    presentationStore.setOnMovePresentation((direction: number) => {
        if (direction === 1 && presentationIndex.value < mediaStore.timelineMedia.length - 1) {
            presentationIndex.value++
        } else if (direction === -1 && presentationIndex.value > 0) {
            presentationIndex.value--
        }

        if (!mediaStore.timelineMedia[presentationIndex.value]) {
            console.warn('No media found at index', presentationIndex.value)
            return
        }

        const newContentID = mediaStore.timelineMedia[presentationIndex.value]?.contentID ?? ''
        presentationStore.setPresentationMediaID(newContentID)

        // if (medias.value.length - presentationIndex.value < 10 && !loading.value) {
        //     // If we are near the end, fetch more media
        //     loading.value = true
        //     fetchMore().finally(() => {
        //         page.value++
        //         loading.value = false
        //     })
        // }
    })
})
</script>
