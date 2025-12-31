import { defineStore } from 'pinia'
import { useWeblensAPI } from '~/api/AllApi'
import WeblensShare from '~/types/weblensShare'

export enum FbModeT {
    unset,
    default,
    share,
    external,
    stats,
    search,
}

const useLocationStore = defineStore('location', () => {
    const route = useRoute()
    const userStore = useUserStore()

    const user = computed(() => userStore.user)

    const isHistoryOpen = ref<boolean>(false)

    const activeFolderID = computed(() => {
        let fileID = route.params.fileID
        if (fileID === 'home') {
            fileID = user.value.homeID
        }

        if (fileID === 'trash') {
            fileID = user.value.trashID
        }

        return fileID as string
    })

    const activeShareID = computed(() => {
        return (route.params.shareID as string | undefined) ?? ''
    })

    const isInShare = computed(() => {
        return (route.name as string | undefined)?.startsWith('files-share') ?? false
    })

    const { data: activeShare } = useAsyncData('share-' + route.params.shareID, async () => {
        if (!route.params.shareID) {
            return
        }

        if (Array.isArray(route.params.shareID)) {
            console.warn('ShareID param is array')
            return
        }

        const shareInfo = (await useWeblensAPI().SharesAPI.getFileShare(route.params.shareID)).data

        return new WeblensShare(shareInfo)
    })

    const inShareRoot = computed(() => {
        return isInShare.value && !activeShare.value
    })

    const isInFiles = computed(() => {
        return (route.name as string | undefined)?.startsWith('files') ?? false
    })

    watchEffect(() => {
        if (!isInFiles.value) return

        const loggedIn = userStore.user.isLoggedIn
        if ((!isInShare.value || !activeShareID.value) && loggedIn.isSet() && !loggedIn.get()) {
            console.warn('User is not logged in and not in share, redirecting to login page')

            return navigateTo({ path: '/login' })
        }

        if (isInShare && activeShare.value && !route.params.fileID) {
            return navigateTo({
                path: `/files/share/${activeShareID.value}/${activeShare.value?.fileID}`,
                query: route.query,
            })
        }
    })

    const isInTimeline = computed(() => {
        return route.query['timeline'] === 'true'
    })

    const isInTrash = computed(() => {
        return activeFolderID.value === user.value.trashID
    })

    const operatingSystem = computed(() => {
        if (navigator.userAgent.indexOf('Win') != -1) {
            return 'windows'
        } else if (navigator.userAgent.indexOf('Mac') != -1) {
            return 'macos'
        }

        return ''
    })

    const viewTimestamp = computed(() => {
        const rewindTo = route.query['rewindTo']
        if (rewindTo) {
            const ts = new Date(rewindTo as string).getTime()
            if (!isNaN(ts)) {
                return ts
            }
        }

        return 0
    })

    const isViewingPast = computed(() => {
        return viewTimestamp.value > 0
    })

    async function setTimeline(timeline: boolean) {
        await navigateTo({
            query: {
                ...route.query,
                timeline: String(timeline),
            },
        })
    }

    function setHistoryOpen(opened: boolean) {
        isHistoryOpen.value = opened
    }

    async function setViewTimestamp(ts: number) {
        const tsString = ts > 0 ? new Date(ts).toISOString() : undefined

        await navigateTo({
            query: {
                ...route.query,
                rewindTo: tsString,
            },
        })
    }

    return {
        activeShareID,
        isInShare,
        activeShare,
        inShareRoot,

        activeFolderID,
        isInTimeline,
        isInTrash,

        operatingSystem,

        setTimeline,

        isHistoryOpen,
        setHistoryOpen,

        viewTimestamp,
        isViewingPast,
        setViewTimestamp,
    }
})

export default useLocationStore
