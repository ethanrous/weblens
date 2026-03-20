import { defineStore } from 'pinia'
import { SubToFolder, UnsubFromFolder } from '~/api/FileBrowserApi'
import WeblensFile, { SelectedState } from '~/types/weblensFile'
import useLocationStore from './location'
import useTagsStore from './tags'
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
    SearchFilesSortOrderEnum,
    SearchFilesSortPropEnum,
} from '@ethanrous/weblens-api'
import { WLError } from '~/types/wlError'

export type FileShape = 'square' | 'row' | 'column'
export type SortCondition = SearchFilesSortPropEnum & GetFolderSortPropEnum
type SortDirection = SearchFilesSortOrderEnum & GetFolderSortOrderEnum

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
    const filterTagIDs = ref<Set<string>>(new Set())
    const filterTagMode = ref<'and' | 'or'>('and')

    const searchUpToDate = ref<boolean>(true)

    const isSearching = computed(() => {
        return locationStore.search !== '' || filterTagIDs.value.size > 0
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

    // Sync searchRecursively with ?recursive query param
    watch(
        () => locationStore.getQueryParam('recursive'),
        (newVal) => {
            searchRecursively.value = newVal === 'true'
        },
        { immediate: true },
    )
    watch(searchRecursively, () => {
        locationStore.setQueryParam('recursive', searchRecursively.value ? 'true' : null)
    })

    // Sync searchWithRegex with ?regex query param
    watch(
        () => locationStore.getQueryParam('regex'),
        (newVal) => {
            searchWithRegex.value = newVal === 'true'
        },
        { immediate: true },
    )
    watch(searchWithRegex, () => {
        locationStore.setQueryParam('regex', searchWithRegex.value ? 'true' : null)
    })

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
                (!locationStore.activeFolderID && !locationStore.isInShare) ||
                (locationStore.isInShare && locationStore.activeShareID && !locationStore.activeShare)
            ) {
                return {}
            }

            let res: { r: AxiosResponse<FolderInfo>; t: 'folder' } | { r: AxiosResponse<FileInfo>; t: 'file' }
            if (locationStore.isInShare && !locationStore.activeShareID) {
                res = { r: await useWeblensAPI().FilesAPI.getSharedFiles(), t: 'folder' }
            } else if (locationStore.isInShare && locationStore.activeShare && !locationStore.activeShare.isDir) {
                res = {
                    r: await useWeblensAPI().FilesAPI.getFile(
                        locationStore.activeShare.fileID,
                        locationStore.activeShareID,
                    ),
                    t: 'file',
                }
            } else {
                res = {
                    r: await useWeblensAPI().FoldersAPI.getFolder(
                        locationStore.activeFolderID,
                        locationStore.activeShareID,
                        locationStore.viewTimestamp,
                        sortCondition.value,
                        sortDirection.value,
                    ),
                    t: 'folder',
                }
            }

            if (res.t === 'file') {
                const activeFile = new WeblensFile(res.r.data)
                return { activeFile }
            }

            if (!res.r.data.self || !res.r.data.children) {
                return {}
            }

            const newChildren = res.r.data.children
                ?.map((fInfo) => {
                    const f = new WeblensFile(fInfo)
                    f.displayable =
                        (f.contentID !== '' &&
                            res.r.data.medias?.findIndex((mediaInfo) => mediaInfo.contentID === f.contentID) !== -1) ??
                        false
                    if (locationStore.highlightFileID !== '' && locationStore.highlightFileID === f.ID()) {
                        setSelected(f.ID(), true)
                    }
                    return f
                })
                .filter((file) => !file.IsTrash())

            const mediaMap = new Map<string, WeblensMedia>()
            res.r.data.medias?.forEach((mInfo) => {
                const m = new WeblensMedia(mInfo)
                mediaMap.set(m.contentID, m)
            })

            mediaStore.addMedia(...(res.r.data.medias ?? []))

            newChildren.forEach((f) => {
                const m = mediaMap.get(f.contentID)
                if (!m) {
                    return
                }

                f.contentCreationDate = new Date(m.createDate)
            })

            // children.value = newChildren

            const parents = res.r.data.parents?.map((fInfo) => new WeblensFile(fInfo))
            const activeFile = new WeblensFile(res.r.data.self)
            return { activeFile: activeFile, children: newChildren, parents }
        },
        {
            watch: [
                user,
                () => locationStore.activeFolderID,
                () => locationStore.viewTimestamp,
                () => locationStore.isInTimeline,
                () => locationStore.activeShare,
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
                (!locationStore.search && filterTagIDs.value.size === 0) ||
                !searchUpToDate.value ||
                locationStore.isInTimeline
            ) {
                return
            }

            const res = await useWeblensAPI().FilesAPI.searchFiles(
                locationStore.search as string,
                locationStore.activeFolderID,
                sortCondition.value,
                sortDirection.value,
                searchRecursively.value,
                searchWithRegex.value,
                Array.from(filterTagIDs.value).join(','),
                filterTagMode.value,
            )

            const results = res.data.files.map((f: FileInfo) => {
                return new WeblensFile(f)
            })

            mediaStore.addMedia(...(res.data.medias ?? []))

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
                filterTagIDs,
                filterTagMode,
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
        let result: WeblensFile[]

        if (searchResults.value !== undefined) {
            result = searchResults.value ?? []
        } else if (!filesResponse.value?.children) {
            result = []
        } else {
            result = filesResponse.value?.children.filter((f) => !f.IsTrash())
        }

        // Apply tag filter if any tags are selected
        if (filterTagIDs.value.size > 0) {
            const tagsStore = useTagsStore()
            const match = filterTagMode.value === 'and' ? 'every' : 'some'
            result = result.filter((f) => {
                const fileTags = tagsStore.getTagsByFileID(f.ID())
                return [...filterTagIDs.value][match]((tagID) => fileTags.some((t) => t.id === tagID))
            })
        }

        files.value = result
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

    // Insert a file into the files array based on the current sort condition and direction
    function insertFileSorted(filesList: WeblensFile[], newFile: WeblensFile): WeblensFile[] {
        const compare = (a: WeblensFile, b: WeblensFile) => {
            let aValue: string | number | Date = ''
            let bValue: string | number | Date = ''

            switch (sortCondition.value) {
                case 'name':
                    aValue = a.GetFilename()
                    bValue = b.GetFilename()
                    break
                case 'updatedAt':
                    aValue = a.GetModified()
                    bValue = b.GetModified()
                    break
                case 'size':
                    aValue = a.GetSize()
                    bValue = b.GetSize()
                    break
            }

            if (aValue < bValue) {
                return sortDirection.value === 'asc' ? -1 : 1
            } else if (aValue > bValue) {
                return sortDirection.value === 'asc' ? 1 : -1
            } else {
                return 0
            }
        }

        filesList.push(newFile)
        filesList.sort(compare)

        return filesList
    }

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

        // Inherit permissions from parent if permissions are missing.
        if (!newFile.permissions && newFile.ParentID() === locationStore.activeFolderID) {
            newFile.permissions = activeFile.value?.permissions
        }

        // Get index of file in files array, if it exists
        const index = files.value.findIndex((f) => f.ID() === newFile.ID())
        if (index === -1) {
            files.value = insertFileSorted(files.value, newFile)
        } else {
            files.value[index] = newFile
        }

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

    function setFilterTagIDs(tagIDs: Set<string>) {
        filterTagIDs.value = tagIDs
    }

    function setFilterTagMode(mode: 'and' | 'or') {
        filterTagMode.value = mode
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

    function clearSearch() {
        locationStore.search = ''
        searchRecursively.value = false
        searchWithRegex.value = false
        filterTagIDs.value = new Set()
        filterTagMode.value = 'and'
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

        filterTagIDs,
        setFilterTagIDs,
        filterTagMode,
        setFilterTagMode,

        clearSearch,
    }
})

export default useFilesStore
