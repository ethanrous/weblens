import { useStorage } from '@vueuse/core'
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

type HistorySettings = {
    isOpen: boolean
    width: number
}

const useLocationStore = defineStore('location', () => {
    const router = useRouter()
    const route = computed(() => router.currentRoute.value)

    // Synchronous query state — always up-to-date, no async lag.
    // localQuery is the single source of truth for query params within the app.
    const localQuery = reactive<Record<string, string>>({})

    // Sync from route on external changes (page load, back/forward, link navigation)
    watch(
        () => route.value.query,
        (q) => {
            for (const key of Object.keys(localQuery)) {
                if (!(key in q)) Reflect.deleteProperty(localQuery, key)
            }
            for (const [k, v] of Object.entries(q)) {
                if (typeof v === 'string') localQuery[k] = v
            }
        },
        { immediate: true },
    )

    function setQueryParam(key: string, value: string | null | undefined) {
        if (value === null || value === undefined || value === '') {
            Reflect.deleteProperty(localQuery, key)
        } else {
            localQuery[key] = value
        }
        navigateTo({ query: { ...localQuery } })
    }

    function getQueryParam(key: string): string | undefined {
        return localQuery[key]
    }

    const userStore = useUserStore()
    const towerStore = useTowerStore()

    const user = computed(() => userStore.user)

    const returnTo = ref<string | null>(null)

    const historySettings = useStorage('wl-history-view', {} as HistorySettings)
    const isHistoryOpen = ref<boolean>(historySettings.value.isOpen === true)
    const historyWidth = ref<number>(historySettings.value.width ?? 640)

    watch(isHistoryOpen, (newVal) => {
        historySettings.value = {
            ...historySettings.value,
            isOpen: newVal,
        }
    })

    watch(historyWidth, (newVal) => {
        historySettings.value = {
            ...historySettings.value,
            width: newVal,
        }
    })

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
        return (route.value.params.shareID as string | undefined) ?? getQueryParam('shareID') ?? ''
    })

    const activeTagID = computed(() => {
        return (route.value.params.tagID as string | undefined) ?? ''
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

    const { data: activeShare } = useAsyncData(
        'active-share',
        async () => {
            const shareID = activeShareID.value
            if (!shareID) return null
            const shareReq = await useWeblensAPI()
                .SharesAPI.getFileShare(shareID)
                .catch((err) => {
                    if (err.status === 401) {
                        console.warn('Unauthorized access to share, redirecting to login page.')
                        navigateTo('/login?returnTo=' + encodeURIComponent(route.value.fullPath))

                        return null
                    }
                })

            if (!shareReq) return null

            const shareInfo = shareReq.data
            return new WeblensShare(shareInfo)
        },
        { watch: [activeShareID] },
    )

    const inShareRoot = computed(() => {
        return isInShare.value && !activeShareID.value
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
            return navigateTo('/files/home')
        }

        // If not logged in and on a settings page, redirect to login
        if (route.value.path.startsWith('/settings') && !isLoggedIn) {
            return navigateTo({ path: '/login' })
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
        if (navigator.userAgent.indexOf('Win') !== -1) {
            return 'windows'
        } else if (navigator.userAgent.indexOf('Mac') !== -1) {
            return 'macos'
        }

        return ''
    })

    const viewTimestamp = computed(() => {
        const rewindTo = getQueryParam('rewindTo')
        if (rewindTo) {
            const ts = new Date(rewindTo).getTime()
            if (!isNaN(ts)) return ts
        }
        return 0
    })

    const isViewingPast = computed(() => {
        return viewTimestamp.value > 0
    })

    const isInSettings = computed(() => {
        return (route.value.name as string | undefined)?.startsWith('settings') ?? false
    })

    function setHistoryOpen(opened: boolean) {
        isHistoryOpen.value = opened
    }

    function setHistoryWidth(w: number) {
        historyWidth.value = Math.round(Math.max(280, Math.min(w, window.innerWidth * 0.5)))
    }

    async function setViewTimestamp(ts: number) {
        const tsString = ts > 0 ? new Date(ts).toISOString() : undefined
        setQueryParam('rewindTo', tsString)
    }

    const activeTowerID = computed(() => {
        return (route.value.params.towerID as string | undefined) ?? ''
    })

    async function setActiveTowerID(towerID: string | null) {
        return navigateTo({ path: '/backup' + (towerID ? '/' + towerID : '') })
    }

    const search = ref('')
    watch(
        () => getQueryParam('search'),
        (newVal) => {
            search.value = newVal ?? ''
        },
        { immediate: true },
    )
    watch(search, () => {
        setQueryParam('search', search.value || null)
    })

    const isInTimeline = ref(false)
    watch(
        () => getQueryParam('timeline'),
        (newVal) => {
            isInTimeline.value = newVal === 'true'
        },
        { immediate: true },
    )
    watch(isInTimeline, () => {
        setQueryParam('timeline', isInTimeline.value ? 'true' : null)
        search.value = ''
        isHistoryOpen.value = false
    })

    return {
        activeShareID,
        isInShare,
        activeShare,
        inShareRoot,

        activeTagID,

        activeTowerID,
        setActiveTowerID,

        activeFolderID,
        isInTimeline,

        // isInTrash is true when the active file is, or is in, the users trash.
        isInTrash,

        isInSettings,
        returnTo,

        highlightFileID,

        operatingSystem,

        isHistoryOpen,
        setHistoryOpen,
        historyWidth,
        setHistoryWidth,

        viewTimestamp,
        isViewingPast,
        setViewTimestamp,

        search,

        setQueryParam,
        getQueryParam,
    }
})

export default useLocationStore
