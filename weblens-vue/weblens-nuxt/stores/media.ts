import { defineStore } from 'pinia'
import WeblensMedia from '~/types/weblensMedia'
import type { ShallowRef } from 'vue'
import { useWeblensAPI } from '~/api/AllApi'
import type { MediaInfo, MediaTypeInfo } from '@ethanrous/weblens-api'
import useLocationStore from './location'
import { useStorage } from '@vueuse/core'
import type { WLError } from '~/types/wlError'

export const TIMELINE_PAGE_SIZE = 200
export const TIMELINE_IMAGE_MIN_SIZE = 150
export const TIMELINE_IMAGE_MAX_SIZE = 450

type MediaSettings = {
    showRaw: boolean
    sortDirection: 1 | -1
}

const mediaSettingsDefaults: MediaSettings = {
    sortDirection: -1,
    showRaw: true,
}

export const useMediaStore = defineStore('media', () => {
    const locationStore = useLocationStore()

    const mediaMap: ShallowRef<Map<string, WeblensMedia>> = shallowRef(new Map())
    const timelineMedia = shallowRef<WeblensMedia[]>([])
    const mediaTypeMap: Record<string, MediaTypeInfo> = {}

    const totalMedias = ref<number>(-1)
    const canLoadMore = ref<boolean>(true)
    const mediaPageNum = ref<number>(0)

    const timelineLoading = ref<Promise<void> | null>(null)
    const timelineFetchError = ref<WLError | null>(null)

    const fetching: ShallowRef<Map<string, Promise<WeblensMedia | undefined>>> = shallowRef(new Map())

    const timelineSort = ref<'createDate'>('createDate')
    const timelineSortDirection = ref<1 | -1>(-1) // 1 for ascending, -1 for descending
    const timelineImageSize = ref<number>(200)

    const searchFilters = useStorage('wl-media-settings', {} as Record<string, MediaSettings>)

    const showRaw = ref(true)
    watch(
        () => locationStore.getQueryParam('raw'),
        (newVal) => {
            showRaw.value = newVal !== 'false'
        },
        { immediate: true },
    )
    watch(showRaw, () => {
        locationStore.setQueryParam('raw', showRaw.value ? null : 'false')
    })

    function initSearchFilters() {
        if (searchFilters.value[locationStore.activeFolderID]) {
            return
        }

        searchFilters.value[locationStore.activeFolderID] = { ...mediaSettingsDefaults }
    }

    function clearData() {
        // Reset media list
        mediaPageNum.value = 0
        timelineMedia.value = []
        canLoadMore.value = true
        mediaMap.value = new Map()
        timelineFetchError.value = null
    }

    function saveSearchFilters() {
        initSearchFilters()

        searchFilters.value[locationStore.activeFolderID] = {
            sortDirection: timelineSortDirection.value,
            showRaw: showRaw.value,
        }

        clearData()
    }

    async function fetchSingleMedia(contentID: string): Promise<WeblensMedia | undefined> {
        if (timelineLoading.value) {
            await timelineLoading.value
        }

        if (mediaMap.value.has(contentID)) {
            return mediaMap.value.get(contentID)
        }

        let mp = fetching.value.get(contentID)
        let hasPromise = true

        if (!mp) {
            hasPromise = false
            mp = useWeblensAPI()
                .MediaAPI.getMediaInfo(contentID)
                .then((res) => new WeblensMedia(res.data))
                .catch((err) => {
                    console.error('Error fetching single media:', err)
                    return undefined
                })
            fetching.value.set(contentID, mp)
        }

        const m = await mp

        if (!hasPromise) {
            if (m) {
                addMedia(m)
            }
            fetching.value.delete(contentID)
        }

        return m
    }

    async function fetchMoreMedia(): Promise<void> {
        if (!locationStore.isInTimeline) {
            return Promise.reject('not in timeline')
        }

        if (timelineFetchError.value) {
            console.debug('Previous timeline fetch error exists, not fetching more')
            return Promise.resolve()
        }

        if (timelineLoading.value || !canLoadMore.value) {
            console.debug('Already loading timeline or cannot load more')
            return Promise.resolve()
        }

        const timelinePromise = useWeblensAPI()
            .MediaAPI.getMedia(
                {
                    raw: showRaw.value,
                    hidden: false,
                    sort: timelineSort.value,
                    sortDirection: timelineSortDirection.value,
                    page: mediaPageNum.value++,
                    limit: TIMELINE_PAGE_SIZE,
                    folderIDs: [locationStore.activeFolderID],
                    search: locationStore.search,
                },
                locationStore.activeShareID,
            )
            .then((res) => {
                const medias =
                    res.data.Media?.map((mInfo, i) => {
                        const m = new WeblensMedia(mInfo)
                        m.index = (mediaPageNum.value - 1) * TIMELINE_PAGE_SIZE + i
                        return m
                    }) ?? []
                return {
                    medias,
                    totalMedias: res.data.mediaCount ?? 0,
                    canLoadMore: medias.length === TIMELINE_PAGE_SIZE,
                }
            })
            .then(({ medias: newMedias, totalMedias: totalMediasResponse, canLoadMore: canLoadMoreResponse }) => {
                totalMedias.value = totalMediasResponse
                canLoadMore.value = canLoadMoreResponse

                if (
                    timelineMedia.value.length > 0 &&
                    timelineMedia.value[timelineMedia.value.length - 1]?.index + 1 != newMedias[0]?.index
                ) {
                    console.warn('Media fetch returned overlapping media, skipping addition')
                    return
                } else if (timelineMedia.value.length === 0 && newMedias[0]?.index != 0) {
                    console.warn('Media fetch did not start at index 0, skipping addition')
                    return
                }

                timelineMedia.value = [...timelineMedia.value, ...newMedias]
                addMedia(...newMedias)
            })
            .catch((err) => {
                console.error('Error fetching more media:', err)
                timelineFetchError.value = err as WLError
            })
            .finally(() => {
                timelineLoading.value = null
            })

        timelineLoading.value = timelinePromise
        return timelinePromise
    }

    function addMedia(...mediaInfo: MediaInfo[]) {
        for (const m of mediaInfo) {
            if (!m.contentID) {
                console.warn('Media item missing contentID, skipping addition')
                continue
            }

            if (mediaMap.value.has(m.contentID)) {
                console.warn(`Media with contentID ${m.contentID} already exists, skipping addition`)
                continue
            }

            if (m instanceof WeblensMedia) {
                mediaMap.value.set(m.contentID, m)
            } else {
                mediaMap.value.set(m.contentID, new WeblensMedia(m))
            }
        }

        triggerRef(mediaMap)
    }

    function toggleSortDirection() {
        timelineSortDirection.value = timelineSortDirection.value === 1 ? -1 : 1

        saveSearchFilters()
    }

    function updateImageSize(direction: 'increase' | 'decrease' | number) {
        if (typeof direction === 'number') {
            timelineImageSize.value = Math.max(TIMELINE_IMAGE_MIN_SIZE, Math.min(TIMELINE_IMAGE_MAX_SIZE, direction))
            return
        }

        if (direction === 'increase') {
            timelineImageSize.value = Math.min(TIMELINE_IMAGE_MAX_SIZE, timelineImageSize.value + 50)
        } else if (direction === 'decrease') {
            timelineImageSize.value = Math.max(TIMELINE_IMAGE_MIN_SIZE, timelineImageSize.value - 50)
        }
    }

    function setShowRaw(raw: boolean) {
        showRaw.value = raw
        saveSearchFilters()
    }

    watch([() => locationStore.isInTimeline, () => locationStore.activeFolderID], () => {
        clearData()

        if (locationStore.isInTimeline) {
            initSearchFilters()

            timelineSortDirection.value =
                searchFilters.value[locationStore.activeFolderID]?.sortDirection ?? mediaSettingsDefaults.sortDirection

            showRaw.value = searchFilters.value[locationStore.activeFolderID]?.showRaw ?? mediaSettingsDefaults.showRaw
        } else {
            locationStore.setQueryParam('raw', null)
        }
    })

    function getNextMediaID(currentMediaID: string): string | null {
        const currentMedia = mediaMap.value.get(currentMediaID)
        if (!currentMedia) {
            return null
        }

        if (currentMedia.index >= timelineMedia.value.length - 1) {
            return null
        }

        return timelineMedia.value[currentMedia.index + 1]?.contentID ?? null
    }

    function getPreviousMediaID(currentMediaID: string): string | null {
        const currentMedia = mediaMap.value.get(currentMediaID)
        if (!currentMedia) {
            return null
        }

        if (currentMedia.index <= 0) {
            return null
        }

        return timelineMedia.value[currentMedia.index - 1]?.contentID ?? null
    }

    return {
        mediaMap,
        mediaTypeMap,
        timelineImageSize,
        timelineSortDirection,
        showRaw,

        timelineMedia,
        timelineLoading,
        canLoadMore,
        totalMedias,

        addMedia,
        clearData,
        fetchSingleMedia,
        fetchMoreMedia,
        toggleSortDirection,
        updateImageSize,
        setShowRaw,
        getNextMediaID,
        getPreviousMediaID,
    }
})
