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

    const searchUpToDate = ref<boolean>(true)

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
        mediaPageNum.value = 0
        timelineMedia.value = []
        canLoadMore.value = true
        mediaMap.value = new Map()
        timelineFetchError.value = null
        timelineLoading.value = null
        searchUpToDate.value = true
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
                locationStore.activeShareID,
                showRaw.value,
                false,
                timelineSort.value,
                timelineSortDirection.value,
                locationStore.search,
                mediaPageNum.value++,
                TIMELINE_PAGE_SIZE,
                [locationStore.activeFolderID],
            )
            .then((res) => {
                if (timelineLoading.value !== timelinePromise) return

                const medias =
                    res.data.Media?.map((mInfo, i) => {
                        const m = new WeblensMedia(mInfo)
                        m.index = (mediaPageNum.value - 1) * TIMELINE_PAGE_SIZE + i
                        return m
                    }) ?? []

                totalMedias.value = res.data.mediaCount ?? 0
                canLoadMore.value = medias.length === TIMELINE_PAGE_SIZE

                if (
                    timelineMedia.value.length > 0 &&
                    timelineMedia.value[timelineMedia.value.length - 1]?.index + 1 != medias[0]?.index
                ) {
                    console.warn('Media fetch returned overlapping media, skipping addition')
                    return
                } else if (timelineMedia.value.length === 0 && medias[0]?.index !== 0) {
                    console.warn('Media fetch did not start at index 0, skipping addition', {
                        firstIndex: medias[0]?.index,
                        mediaPageNum: mediaPageNum.value,
                        mediasLength: medias.length,
                    })
                    return
                }

                timelineMedia.value = [...timelineMedia.value, ...medias]
                addMedia(...medias)
            })
            .catch((err) => {
                if (timelineLoading.value !== timelinePromise) return
                console.error('Error fetching more media:', err)
                timelineFetchError.value = { status: err.status, message: err.response?.data?.error } as WLError
                canLoadMore.value = false
            })
            .finally(() => {
                if (timelineLoading.value === timelinePromise) {
                    timelineLoading.value = null
                }
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

    watch(
        () => locationStore.search,
        () => {
            if (locationStore.search === '') {
                // If search was cleared, reset timeline to show all media
                clearData()
            } else if (locationStore.search !== '') {
                // If search query changed, but is not cleared, mark timeline media as outdated until next fetch
                searchUpToDate.value = false
            }
        },
    )

    return {
        mediaMap,
        mediaTypeMap,
        timelineImageSize,
        timelineSortDirection,
        showRaw,

        timelineMedia,
        timelineLoading,
        timelineFetchError,
        canLoadMore,
        totalMedias,
        searchUpToDate,

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
