import { defineStore } from 'pinia'
import WeblensMedia from '~/types/weblensMedia'
import type { ShallowRef } from 'vue'
import { useWeblensAPI } from '~/api/AllApi'
import type { MediaInfo, MediaTypeInfo } from '@ethanrous/weblens-api'
import useLocationStore from './location'
import { useStorage } from '@vueuse/core'

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
    const route = useRoute()

    const media: ShallowRef<Map<string, WeblensMedia>> = shallowRef(new Map())
    const mediaTypeMap: Record<string, MediaTypeInfo> = {}

    const fetching: ShallowRef<Map<string, Promise<WeblensMedia | undefined>>> = shallowRef(new Map())

    const timelineSort = ref<'createDate'>('createDate')
    const timelineSortDirection = ref<1 | -1>(-1) // 1 for ascending, -1 for descending
    const timelineImageSize = ref<number>(200)

    const locationStore = useLocationStore()

    const mediaSettings = useStorage('wl-media-settings', {} as Record<string, MediaSettings>)

    const imageSearch = ref<string>('')

    const showRaw = computed(() => {
        return route.query['raw'] !== 'false'
    })

    function initMediaSettings() {
        if (mediaSettings.value[locationStore.activeFolderID]) {
            return
        }

        mediaSettings.value[locationStore.activeFolderID] = { ...mediaSettingsDefaults }
    }

    function saveMediaSettings() {
        initMediaSettings()

        mediaSettings.value[locationStore.activeFolderID] = {
            sortDirection: timelineSortDirection.value,
            showRaw: showRaw.value,
        }
    }

    async function fetchSingleMedia(contentID: string): Promise<WeblensMedia | undefined> {
        if (media.value.has(contentID)) {
            return media.value.get(contentID)
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

    async function fetchMoreMedia(
        pageNum: number,
    ): Promise<{ medias: WeblensMedia[]; totalMedias: number; canLoadMore: boolean }> {
        if (!locationStore.isInTimeline) {
            return Promise.reject('not in timeline')
        }

        return useWeblensAPI()
            .MediaAPI.getMedia(
                {
                    raw: showRaw.value,
                    hidden: false,
                    sort: timelineSort.value,
                    sortDirection: timelineSortDirection.value,
                    page: pageNum,
                    limit: TIMELINE_PAGE_SIZE,
                    folderIDs: [locationStore.activeFolderID],
                    search: imageSearch.value,
                },
                locationStore.activeShareID,
            )
            .then((res) => {
                const medias = res.data.Media?.map((m) => new WeblensMedia(m)) ?? []
                medias.forEach((m) => media.value.set(m.contentID, m))
                return {
                    medias,
                    totalMedias: res.data.mediaCount ?? 0,
                    canLoadMore: medias.length === TIMELINE_PAGE_SIZE,
                }
            })
    }

    function addMedia(...mediaInfo: MediaInfo[]) {
        for (const m of mediaInfo) {
            if (!m.contentID) {
                console.warn('Media item missing contentID, skipping addition')
                continue
            }

            if (media.value.has(m.contentID)) {
                console.warn(`Media with contentID ${m.contentID} already exists, skipping addition`)
                continue
            }

            if (m instanceof WeblensMedia) {
                media.value.set(m.contentID, m)
            } else {
                media.value.set(m.contentID, new WeblensMedia(m))
            }
        }

        media.value = new Map(media.value) // Trigger reactivity
    }

    function toggleSortDirection() {
        timelineSortDirection.value = timelineSortDirection.value === 1 ? -1 : 1

        saveMediaSettings()
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

    async function setShowRaw(raw: boolean) {
        await navigateTo({
            query: {
                ...route.query,
                raw: String(raw),
            },
        })

        saveMediaSettings()
    }

    function setImageSearch(search: string) {
        imageSearch.value = search
    }

    watch([() => locationStore.isInTimeline, () => locationStore.activeFolderID], async () => {
        if (locationStore.isInTimeline) {
            initMediaSettings()

            timelineSortDirection.value =
                mediaSettings.value[locationStore.activeFolderID]?.sortDirection ?? mediaSettingsDefaults.sortDirection

            await setShowRaw(
                mediaSettings.value[locationStore.activeFolderID]?.showRaw ?? mediaSettingsDefaults.showRaw,
            )
        } else {
            await navigateTo({
                query: {
                    ...route.query,
                    raw: undefined,
                },
            })
        }
    })

    return {
        media,
        mediaTypeMap,
        timelineImageSize,
        timelineSortDirection,
        showRaw,
        imageSearch,
        addMedia,
        fetchSingleMedia,
        fetchMoreMedia,
        toggleSortDirection,
        updateImageSize,
        setShowRaw,
        setImageSearch,
    }
})
