import { defineStore } from 'pinia'
import { SubToFolder, UnsubFromFolder } from '~/api/FileBrowserApi'
import WeblensFile, { SelectedState } from '~/types/weblensFile'
import useLocationStore from './location'
import { onWatcherCleanup } from 'vue'
import type { AxiosResponse } from 'axios'
import WeblensMedia from '~/types/weblensMedia'
import { useWeblensAPI } from '~/api/AllApi'
import { useStorage } from '@vueuse/core'
import type {
    FileInfo,
    FolderInfo,
    GetFolderSortOrderEnum,
    GetFolderSortPropEnum,
    SearchByFilenameSortOrderEnum,
    SearchByFilenameSortPropEnum,
} from '@ethanrous/weblens-api'
import { WLError } from '~/types/wlError'

export type FileShape = 'square' | 'row' | 'column'
export type SortCondition = SearchByFilenameSortPropEnum & GetFolderSortPropEnum
type SortDirection = SearchByFilenameSortOrderEnum & GetFolderSortOrderEnum

type FolderSettings = {
    sortCondition: SortCondition
    sortDirection: SortDirection
    fileShape: FileShape
}

const folderSettingsDefault: FolderSettings = {
    sortCondition: 'updatedAt',
    sortDirection: 'asc',
    fileShape: 'square',
}

const useFilesStore = defineStore('files', () => {
    // External //
    const userStore = useUserStore()
    const locationStore = useLocationStore()
    const mediaStore = useMediaStore()
    const user = computed(() => userStore.user)

    // Local State //
    const files = shallowRef<WeblensFile[]>([])

    const selectedFiles = ref<Set<string>>(new Set())

    const lastSelectedID = ref<string | null>(null)
    const nextSelectedIndex = ref<number | null>(null) // This is used to track the next file to be selected when using shift-click
    const shiftPressed = ref<boolean>(false) // This is used to track the next file to be selected when using shift-click

    const sortCondition = ref<SortCondition>('name')
    const sortDirection = ref<SortDirection>('asc')

    const fileShape = ref<FileShape>('square')

    const dragging = ref<boolean>(false)

    const foldersSettings = useStorage('wl-folders-settings', {} as Record<string, FolderSettings>)

    const searchRecursively = ref<boolean>(false)
    const searchWithRegex = ref<boolean>(false)

    const searchUpToDate = ref<boolean>(true)

    const isSearching = computed(() => {
        return locationStore.search !== ''
    })

    watch(
        () => locationStore.activeFolderID,
        (_, prev) => {
            selectedFiles.value = new Set()

            if (prev) {
                locationStore.search = ''
            }

            initFolderSettings()

            sortCondition.value = foldersSettings.value[locationStore.activeFolderID]?.sortCondition ?? 'updatedAt'
            sortDirection.value = foldersSettings.value[locationStore.activeFolderID]?.sortDirection ?? 'asc'
            fileShape.value = foldersSettings.value[locationStore.activeFolderID]?.fileShape ?? 'square'
        },
        { immediate: true },
    )

    const {
        data: filesResponse,
        error,
        status: folderStatus,
    } = useAsyncData(
        'files-' + locationStore.activeFolderID,
        async () => {
            if (
                !user.value.isLoggedIn.isSet() ||
                locationStore.isInTimeline ||
                (!locationStore.activeFolderID && !locationStore.isInShare)
            ) {
                return {}
            }

            let res: AxiosResponse<FolderInfo, FolderInfo>
            if (locationStore.isInShare && !locationStore.activeShareID) {
                res = await useWeblensAPI().FilesAPI.getSharedFiles()
            } else {
                res = await useWeblensAPI().FoldersAPI.getFolder(
                    locationStore.activeFolderID,
                    locationStore.activeShareID,
                    locationStore.viewTimestamp,
                    sortCondition.value,
                    sortDirection.value,
                )
            }

            if (!res.data.self || !res.data.children) {
                return {}
            }

            const newChildren = res.data.children
                ?.map((fInfo) => {
                    const f = new WeblensFile(fInfo)
                    f.displayable =
                        (f.contentID !== '' &&
                            res.data.medias?.findIndex((mediaInfo) => mediaInfo.contentID === f.contentID) !== -1) ??
                        false
                    if (locationStore.highlightFileID !== '' && locationStore.highlightFileID === f.ID()) {
                        setSelected(f.ID(), true)
                    }
                    return f
                })
                .filter((file) => !file.IsTrash())

            const mediaMap = new Map<string, WeblensMedia>()
            res.data.medias?.forEach((mInfo) => {
                const m = new WeblensMedia(mInfo)
                mediaMap.set(m.contentID, m)
            })

            mediaStore.addMedia(...(res.data.medias ?? []))

            newChildren.forEach((f) => {
                const m = mediaMap.get(f.contentID)
                if (!m) {
                    return
                }

                f.contentCreationDate = new Date(m.createDate)
            })

            // children.value = newChildren

            const parents = res.data.parents?.map((fInfo) => new WeblensFile(fInfo))
            const activeFile = new WeblensFile(res.data.self)
            return { activeFile: activeFile, children: newChildren, parents }
        },
        {
            watch: [
                user,
                () => locationStore.activeFolderID,
                () => locationStore.viewTimestamp,
                () => locationStore.isInTimeline,
                isSearching,
                sortCondition,
                sortDirection,
            ],
            lazy: true,
        },
    )

    const fileFetchError = computed(() => {
        if (!error.value) {
            return null
        }

        return new WLError(error.value)
    })

    const { data: searchResults, status: searchStatus } = useAsyncData(
        'file-search',
        async () => {
            if (
                !locationStore.search ||
                !searchUpToDate.value ||
                locationStore.isInTimeline ||
                (locationStore.search as string).trim() === ''
            ) {
                return
            }

            const res = await useWeblensAPI().FilesAPI.searchByFilename(
                locationStore.search as string,
                locationStore.activeFolderID,
                sortCondition.value,
                sortDirection.value,
                searchRecursively.value,
                searchWithRegex.value,
            )

            const results = res.data.map((f) => {
                return new WeblensFile(f)
            })

            return results
        },
        {
            watch: [
                () => locationStore.search,
                searchRecursively,
                searchWithRegex,
                sortCondition,
                sortDirection,
                searchUpToDate,
            ],
            lazy: true,
        },
    )

    // Computed Properties //
    const activeFile = computed(() => {
        return filesResponse.value?.activeFile
    })

    const parents = computed(() => {
        return filesResponse.value?.parents
    })

    // Watchers //

    // When the active folder changes, clear selected files and reinitialize folder settings
    watchEffect(() => {
        const _activeFolderID = locationStore.activeFolderID
        if (!_activeFolderID) {
            return
        }

        SubToFolder(_activeFolderID, locationStore.activeShareID)

        onWatcherCleanup(() => {
            UnsubFromFolder(_activeFolderID)
        })
    })

    watchEffect(() => {
        if (isSearching.value) {
            files.value = searchResults.value ?? []
            return
        }

        if (!filesResponse.value?.children) {
            files.value = []

            return
        }

        files.value = filesResponse.value?.children.filter((f) => !f.IsTrash())
    })

    const lastSelectedIndex = computed(() => {
        if (!lastSelectedID.value) {
            return -1
        }

        return files.value.findIndex((f) => f.ID() === lastSelectedID.value)
    })

    watch(
        () => locationStore.search,
        () => {
            searchUpToDate.value = !searchRecursively.value || locationStore.search.trim() === ''
        },
    )

    const loading = computed(() => {
        return searchStatus.value === 'pending' || folderStatus.value === 'pending'
    })

    // Funcs //
    function setSelected(fileID: string, selected: boolean, doShiftSelect = false) {
        if (selected) {
            if (doShiftSelect && lastSelectedIndex.value !== -1 && nextSelectedIndex.value !== null) {
                const startIndex = Math.min(lastSelectedIndex.value, nextSelectedIndex.value)
                const endIndex = Math.max(lastSelectedIndex.value, nextSelectedIndex.value)

                for (let i = startIndex; i <= endIndex; i++) {
                    const file = files.value[i]
                    if (file) {
                        selectedFiles.value.add(file.ID())
                    }
                }
            }

            selectedFiles.value.add(fileID)
            lastSelectedID.value = fileID
        } else {
            selectedFiles.value.delete(fileID)
        }

        selectedFiles.value = new Set(selectedFiles.value)
    }

    function setNextSelectedIndex(index: number) {
        nextSelectedIndex.value = index
    }

    function clearNextSelectedIndex() {
        nextSelectedIndex.value = null
    }

    function setShiftPressed(pressed?: boolean) {
        if (pressed === undefined) {
            return
        }

        shiftPressed.value = pressed
    }

    function selectAll() {
        if (!files.value) {
            return
        }

        selectedFiles.value = new Set(files.value.map((file) => file.ID()))
    }

    function clearSelected() {
        selectedFiles.value = new Set()
    }

    function getFileByID(id: string): WeblensFile | undefined {
        if (!files.value) {
            return undefined
        }

        if (id === locationStore.activeFolderID) {
            return activeFile.value
        }

        return files.value.find((file) => file.ID() === id)
    }

    function addFile(file: FileInfo) {
        if (!files.value) {
            return
        }

        if (file.parentID !== locationStore.activeFolderID) {
            return
        }

        const newFile = new WeblensFile(file)

        const newFiles = files.value
        const index = newFiles.findIndex((file) => file.ID() === newFile.ID())
        if (index !== -1) {
            newFiles.splice(index, 1)
        }

        if (!newFile.permissions && newFile.ParentID() === locationStore.activeFolderID) {
            newFile.permissions = activeFile.value?.permissions
        }

        newFiles.push(newFile)

        files.value = newFiles

        triggerRef(files)
    }

    function setMovedFile(...fileIDs: string[]) {
        const filesMap = files.value.reduce<Record<string, WeblensFile | undefined>>((acc, f) => {
            acc[f.ID()] = f
            return acc
        }, {})

        for (const fileID of fileIDs) {
            if (fileID === locationStore.activeFolderID) {
                continue
            }

            filesMap[fileID]?.SetSelected(SelectedState.Moved)
        }

        // Trigger reactivity
        triggerRef(files)
    }

    function removeFiles(...fileIDs: string[]) {
        const filesMap = files.value.reduce<Record<string, WeblensFile | undefined>>((acc, f) => {
            acc[f.ID()] = f
            return acc
        }, {})

        for (const fileID of fileIDs) {
            if (fileID === locationStore.activeFolderID) {
                console.warn('Cannot remove the active folder')
                continue
            }

            filesMap[fileID]?.SetSelected(SelectedState.Moved)
            filesMap[fileID] = undefined
        }

        files.value = Object.values(filesMap).filter((f): f is WeblensFile => f !== undefined)

        // Trigger reactivity
        triggerRef(files)
    }

    function initFolderSettings() {
        if (foldersSettings.value[locationStore.activeFolderID]) {
            return
        }

        foldersSettings.value[locationStore.activeFolderID] = { ...folderSettingsDefault }
    }

    function saveFoldersSettings() {
        initFolderSettings()

        foldersSettings.value[locationStore.activeFolderID] = {
            sortCondition: sortCondition.value,
            sortDirection: sortDirection.value,
            fileShape: fileShape.value,
        }
    }

    function toggleSortDirection() {
        sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'

        saveFoldersSettings()
    }

    function setSortCondition(newSortCondition: SortCondition) {
        sortCondition.value = newSortCondition

        saveFoldersSettings()
    }

    function setFileShape(newFileShape: FileShape) {
        fileShape.value = newFileShape

        saveFoldersSettings()
    }

    function setDragging(newDragging: boolean) {
        dragging.value = newDragging
    }

    function setSearchRecurively(recursive: boolean) {
        searchRecursively.value = recursive
    }

    function setSearchWithRegex(useRegex: boolean) {
        searchWithRegex.value = useRegex
    }

    function getNextFileID(currentFileID: string): string | null {
        const index = files.value.findIndex((f) => f.ID() === currentFileID)
        if (index === -1 || index + 1 >= files.value.length) {
            return null
        }

        return files.value[index + 1].ID()
    }

    function getPreviousFileID(currentFileID: string): string | null {
        const index = files.value.findIndex((f) => f.ID() === currentFileID)
        if (index === -1 || index - 1 < 0) {
            return null
        }

        return files.value[index - 1].ID()
    }

    return {
        files,

        activeFile,

        dragging,

        parents,
        fileFetchError,

        loading,

        selectedFiles,

        lastSelected: lastSelectedID,
        lastSelectedIndex,
        nextSelectedIndex,
        setNextSelectedIndex,
        clearNextSelectedIndex,

        getNextFileID,
        getPreviousFileID,

        shiftPressed,
        setShiftPressed,

        addFile,
        setMovedFile,
        removeFiles,
        getFileByID,

        setSelected,
        selectAll,
        clearSelected,

        fileShape,
        setFileShape,

        sortDirection,
        sortCondition,
        toggleSortDirection,
        setSortCondition,

        setDragging,

        isSearching,
        searchRecursively,
        setSearchRecurively,
        searchWithRegex,
        setSearchWithRegex,
        searchResults,
        searchUpToDate,
    }
})

export default useFilesStore
