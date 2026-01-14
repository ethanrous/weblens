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
    const towerStore = useTowerStore()

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
        // Wait until user login state is known
        if (!userStore.user.isLoggedIn.isSet()) return

        const isLoggedIn = userStore.user.isLoggedIn.get({ default: false })
        const towerRole = towerStore.towerInfo?.role

        // If the tower is uninitialized, always redirect to setup page
        if (route.path !== '/setup' && towerRole === TowerRole.UNINITIALIZED) {
            return navigateTo('/setup')
        } else if (route.path === '/setup' && towerRole !== TowerRole.UNINITIALIZED) {
            // If the tower is initialized, redirect away from setup page
            if (isLoggedIn) return navigateTo('/files/home')

            return navigateTo('/login')
        }

        // Handle backup tower redirection
        if (towerRole === TowerRole.BACKUP) {
            // If not logged in, redirect to login page
            if (!isLoggedIn) return navigateTo('/login')

            // If logged in, redirect to backup page
            if (route.path.startsWith('/files')) return navigateTo('/backup')

            return
        }

        if (route.path === '/login' && isLoggedIn) {
            console.warn('User is already logged in, redirecting to /files/home')

            return navigateTo('/files/home')
        }

        if (!isInFiles.value) return

        // If not logged in and not in share, redirect to login
        if ((!isInShare.value || !activeShareID.value) && !isLoggedIn) {
            console.warn('User is not logged in and not in share, redirecting to login page')

            return navigateTo({ path: '/login' })
        }

        // If in share but no fileID, redirect to share root. This allows sharing links to be shorter
        // e.g. /files/share/:shareID instead of /files/share/:shareID/:fileID
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
