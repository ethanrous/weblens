import { defineStore } from 'pinia'
import { SubToFolder, UnsubFromFolder } from '~/api/FileBrowserApi'
import WeblensFile from '~/types/weblensFile'
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
    const children = shallowRef<WeblensFile[]>()

    const selectedFiles = ref<Set<string>>(new Set())
    const movedFiles = ref<Set<string>>(new Set())

    const lastSelected = ref<string | null>(null)
    const nextSelectedIndex = ref<number | null>(null) // This is used to track the next file to be selected when using shift-click
    const shiftPressed = ref<boolean>(false) // This is used to track the next file to be selected when using shift-click

    const sortCondition = ref<SortCondition>('name')
    const sortDirection = ref<SortDirection>('asc')

    const fileShape = ref<FileShape>('square')

    const dragging = ref<boolean>(false)

    const foldersSettings = useStorage('wl-folders-settings', {} as Record<string, FolderSettings>)

    const searchRecurively = ref<boolean>(false)
    const searchWithRegex = ref<boolean>(false)

    const searchUpToDate = ref<boolean>(true)

    const loading = ref<boolean>(false)

    const searchResults = shallowRef<WeblensFile[] | undefined>()

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

            searchResults.value = undefined
            searchUpToDate.value = false

            initFolderSettings()

            sortCondition.value = foldersSettings.value[locationStore.activeFolderID]?.sortCondition ?? 'updatedAt'
            sortDirection.value = foldersSettings.value[locationStore.activeFolderID]?.sortDirection ?? 'asc'
            fileShape.value = foldersSettings.value[locationStore.activeFolderID]?.fileShape ?? 'square'
        },
        { immediate: true },
    )

    const { data, error, status } = useAsyncData(
        'files-' + locationStore.activeFolderID,
        async () => {
            if (!user.value.isLoggedIn.isSet() || (!locationStore.activeFolderID && !locationStore.isInShare)) {
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

            children.value = newChildren

            const parents = res.data.parents?.map((fInfo) => new WeblensFile(fInfo))
            const activeFile = new WeblensFile(res.data.self)
            return { activeFile: activeFile, children: newChildren, parents }
        },
        {
            watch: [
                user,
                () => locationStore.activeFolderID,
                () => locationStore.viewTimestamp,
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
            lastSelected.value = fileID
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
        if (!children.value) {
            return
        }

        selectedFiles.value = new Set(children.value.map((file) => file.ID()))
    }

    function clearSelected() {
        selectedFiles.value = new Set()
    }

    function getFileByID(id: string): WeblensFile | undefined {
        if (!children.value) {
            return undefined
        }

        if (id === locationStore.activeFolderID) {
            return activeFile.value
        }

        return children.value.find((file) => file.ID() === id)
    }

    function addFile(file: FileInfo) {
        if (!children.value) {
            return
        }

        if (file.parentID !== locationStore.activeFolderID) {
            return
        }

        const newFile = new WeblensFile(file)

        const index = children.value.findIndex((file) => file.ID() === newFile.ID())
        if (index !== -1) {
            children.value.splice(index, 1)
        }

        children.value.push(newFile)

        children.value = [...children.value]
    }

    function removeFiles(...fileIDs: string[]) {
        if (!children.value) {
            return
        }

        const newChildren = children.value

        for (const fileID of fileIDs) {
            if (fileID === locationStore.activeFolderID) {
                console.warn('Cannot remove the active folder')
                continue
            }

            const index = newChildren.findIndex((file) => file.ID() === fileID)
            if (index !== -1) {
                newChildren.splice(index, 1)
            } else {
                console.warn(`File with ID ${fileID} not found in children`)
            }
        }

        setMovedFile(fileIDs, false)

        // Trigger reactivity
        triggerRef(children)
    }

    function setMovedFile(fileIDs: string[], moved: boolean) {
        for (const fileID of fileIDs) {
            if (moved) {
                movedFiles.value.add(fileID)
            } else {
                movedFiles.value.delete(fileID)
            }
        }

        // Trigger reactivity
        triggerRef(movedFiles)
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

        if (isSearching.value) doSearch()
    }

    function setSortCondition(newSortCondition: SortCondition) {
        sortCondition.value = newSortCondition

        saveFoldersSettings()

        if (isSearching.value) doSearch()
    }

    function setFileShape(newFileShape: FileShape) {
        fileShape.value = newFileShape

        saveFoldersSettings()
    }

    function setDragging(newDragging: boolean) {
        dragging.value = newDragging
    }

    function setSearchRecurively(recursive: boolean) {
        searchRecurively.value = recursive

        if (isSearching.value) doSearch()
    }

    function setSearchWithRegex(useRegex: boolean) {
        searchWithRegex.value = useRegex

        if (isSearching.value) doSearch()
    }

    function setLoading(load: boolean) {
        loading.value = load
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

    async function doSearch() {
        if (!locationStore.search || (locationStore.search as string).trim() === '') {
            searchResults.value = undefined
            return
        }

        loading.value = true

        return (async () => {
            const res = await useWeblensAPI().FilesAPI.searchByFilename(
                locationStore.search as string,
                locationStore.activeFolderID,
                sortCondition.value,
                sortDirection.value,
                searchRecurively.value,
                searchWithRegex.value,
            )
            const results = res.data.map((f) => {
                return new WeblensFile(f)
            })

            searchResults.value = results

            searchUpToDate.value = true
        })().finally(() => {
            loading.value = false
        })
    }

    // Computed Properties //
    const activeFile = computed(() => {
        return data.value?.activeFile
    })

    const parents = computed(() => {
        return data.value?.parents
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

    const lastSelectedIndex = computed(() => {
        if (!lastSelected.value || !children.value) {
            return -1
        }

        return files.value.findIndex((f) => f.ID() === lastSelected.value)
    })

    const files = computed(() => {
        if (isSearching.value) {
            return searchResults.value ?? []
        }

        if (!children.value) {
            return []
        }

        let files = children.value
        if (locationStore.search !== '' && !searchRecurively.value) {
            const search = (locationStore.search as string).toLowerCase()

            files = files.filter((f) => {
                return f.GetFilename().toLowerCase().includes(search)
            })
        }

        files = files.filter((f) => !f.IsTrash())

        return files
    })

    watch(
        () => locationStore.search,
        () => {
            searchUpToDate.value = false
        },
    )

    return {
        files,

        activeFile,

        dragging,

        children,
        parents,
        status,
        fileFetchError,

        loading,
        setLoading,

        movedFiles,
        selectedFiles,

        lastSelected,
        lastSelectedIndex,
        nextSelectedIndex,
        setNextSelectedIndex,
        clearNextSelectedIndex,

        getNextFileID,
        getPreviousFileID,

        shiftPressed,
        setShiftPressed,

        addFile,
        removeFiles,
        getFileByID,
        setMovedFile,

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
        searchRecurively,
        setSearchRecurively,
        searchWithRegex,
        setSearchWithRegex,
        searchResults,
        searchUpToDate,
        doSearch,
    }
})

export default useFilesStore
