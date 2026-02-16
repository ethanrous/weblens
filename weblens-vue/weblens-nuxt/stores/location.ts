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
    const router = useRouter()
    const route = computed(() => router.currentRoute.value)

    // Batched query param updates — merges all setQueryParam calls within the
    // same tick into a single navigateTo so rapid changes don't clobber each other.
    let pendingQueryUpdates: Record<string, string | null> = {}
    let isQueryUpdatePending = false

    function setQueryParam(key: string, value: string | null | undefined) {
        pendingQueryUpdates[key] = value ?? null
        if (!isQueryUpdatePending) {
            isQueryUpdatePending = true
            nextTick(() => {
                const newQuery: Record<string, string | undefined> = {}
                for (const [k, v] of Object.entries(route.value.query)) {
                    if (typeof v === 'string') {
                        newQuery[k] = v
                    }
                }
                let hasChanges = false
                for (const [k, v] of Object.entries(pendingQueryUpdates)) {
                    const newVal = v === null || v === '' ? undefined : v
                    if (newQuery[k] !== newVal) {
                        hasChanges = true
                        newQuery[k] = newVal
                    }
                }
                if (hasChanges) {
                    navigateTo({ query: newQuery })
                }
                pendingQueryUpdates = {}
                isQueryUpdatePending = false
            })
        }
    }

    function getQueryParam(key: string): string | undefined {
        const val = route.value.query[key]
        return typeof val === 'string' ? val : undefined
    }

    const userStore = useUserStore()
    const towerStore = useTowerStore()

    const user = computed(() => userStore.user)

    const isHistoryOpen = ref<boolean>(false)

    const activeFolderID = computed(() => {
        let fileID = route.value.params.fileID
        if (fileID === 'home') {
            fileID = user.value.homeID
        }

        if (fileID === 'trash') {
            fileID = user.value.trashID
        }

        return fileID as string
    })

    const activeShareID = computed(() => {
        return (route.value.params.shareID as string | undefined) ?? ''
    })

    const isInShare = computed(() => {
        return (route.value.name as string | undefined)?.startsWith('files-share') ?? false
    })

    const highlightFileID = computed(() => {
        const hash = route.value.hash
        if (hash.startsWith('#file-')) {
            return hash.slice('#file-'.length)
        }
        return hash
    })

    const { data: activeShare } = useAsyncData('share-' + route.value.params.shareID, async () => {
        if (!route.value.params.shareID) {
            return
        }

        if (Array.isArray(route.value.params.shareID)) {
            console.warn('ShareID param is array')
            return
        }

        const shareInfo = (await useWeblensAPI().SharesAPI.getFileShare(route.value.params.shareID as string)).data

        return new WeblensShare(shareInfo)
    })

    const inShareRoot = computed(() => {
        return isInShare.value && !activeShare.value
    })

    const isInFiles = computed(() => {
        return (route.value.name as string | undefined)?.startsWith('files') ?? false
    })

    watchEffect(() => {
        // Wait until user login state is known
        if (!userStore.user.isLoggedIn.isSet()) return

        const isLoggedIn = userStore.user.isLoggedIn.get({ default: false })
        const towerRole = towerStore.towerInfo?.role

        // If the tower is uninitialized, always redirect to setup page
        if (route.value.path !== '/setup' && towerRole === TowerRole.UNINITIALIZED) {
            return navigateTo('/setup')
        } else if (route.value.path === '/setup' && towerRole !== TowerRole.UNINITIALIZED) {
            // If the tower is initialized, redirect away from setup page
            if (isLoggedIn) return navigateTo('/files/home')

            return navigateTo('/login')
        }

        // Handle backup tower redirection
        if (towerRole === TowerRole.BACKUP) {
            // If not logged in, redirect to login page
            if (!isLoggedIn) return navigateTo('/login')

            // If logged in, redirect to backup page
            if (route.value.path.startsWith('/files')) return navigateTo('/backup')

            return
        }

        if (isLoggedIn && route.value.path === '/files') {
            return navigateTo('/files/home')
        }

        if (route.value.path === '/login' && isLoggedIn) {
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
        if (isInShare && activeShare.value && !route.value.params.fileID) {
            return navigateTo({
                path: `/files/share/${activeShareID.value}/${activeShare.value?.fileID}`,
                query: route.value.query,
            })
        }
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
        const rewindTo = route.value.query['rewindTo']
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

    function setHistoryOpen(opened: boolean) {
        isHistoryOpen.value = opened
    }

    async function setViewTimestamp(ts: number) {
        const tsString = ts > 0 ? new Date(ts).toISOString() : undefined

        await navigateTo({
            query: {
                ...route.value.query,
                rewindTo: tsString,
            },
        })
    }

    const activeTowerID = computed(() => {
        return (route.value.params.towerID as string | undefined) ?? ''
    })

    async function setActiveTowerID(towerID: string | null) {
        return navigateTo({ path: '/backup' + (towerID ? '/' + towerID : '') })
    }

    const search = ref('')
    watch(
        () => route.value.query.search,
        (newVal) => {
            search.value = (typeof newVal === 'string' ? newVal : '') || ''
        },
        { immediate: true },
    )
    watch(search, () => {
        setQueryParam('search', search.value || null)
    })

    const isInTimeline = ref(false)
    watch(
        () => route.value.query.timeline,
        (newVal) => {
            isInTimeline.value = newVal === 'true'
        },
        { immediate: true },
    )
    watch(isInTimeline, () => {
        setQueryParam('timeline', isInTimeline.value ? 'true' : null)
        search.value = '' // Clear search when changing timeline mode
    })

    return {
        activeShareID,
        isInShare,
        activeShare,
        inShareRoot,

        activeTowerID,
        setActiveTowerID,

        activeFolderID,
        isInTimeline,
        isInTrash,

        highlightFileID,

        operatingSystem,

        isHistoryOpen,
        setHistoryOpen,

        viewTimestamp,
        isViewingPast,
        setViewTimestamp,

        search,

        setQueryParam,
        getQueryParam,
    }
})

export default useLocationStore
