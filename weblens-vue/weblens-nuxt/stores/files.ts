import { defineStore } from 'pinia'
import { SubToFolder, UnsubFromFolder } from '~/api/FileBrowserApi'
import WeblensFile from '~/types/weblensFile'
import useLocationStore from './location'
import { onWatcherCleanup } from 'vue'
import type { AxiosResponse } from 'axios'
import WeblensMedia from '~/types/weblensMedia'
import { useWeblensAPI } from '~/api/AllApi'
import { useStorage } from '@vueuse/core'
import type { FileInfo, FolderInfo } from '@ethanrous/weblens-api'

type sorterFunc = (f1: WeblensFile, f2: WeblensFile) => number

export type FileShape = 'square' | 'row' | 'column'
export type SortCondition = 'date' | 'filename' | 'size'
type SortDirection = 1 | -1

type FolderSettings = {
    sortCondition: SortCondition
    sortDirection: SortDirection
    fileShape: FileShape
}

const folderSettingsDefault: FolderSettings = {
    sortCondition: 'date',
    sortDirection: 1,
    fileShape: 'square',
}

function getSortFunc(sortCondition: SortCondition, sortDirection: SortDirection): sorterFunc {
    console.debug('Sorting files by', sortCondition, 'in direction', sortDirection)

    switch (sortCondition) {
        case 'filename': {
            return (f1, f2) => {
                return f1.GetFilename().localeCompare(f2.GetFilename(), undefined, { numeric: true }) * sortDirection
            }
        }
        case 'date': {
            return (f1, f2) => {
                return (f1.GetModified().getTime() - f2.GetModified().getTime()) * sortDirection
            }
        }
        case 'size': {
            return (f1, f2) => {
                return (f1.GetSize() - f2.GetSize()) * sortDirection
            }
        }
    }
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

    const sortDirection = ref<SortDirection>(1)
    const sortCondition = ref<SortCondition>('filename')

    const fileShape = ref<FileShape>('square')

    const dragging = ref<boolean>(false)

    const foldersSettings = useStorage('wl-folders-settings', {} as Record<string, FolderSettings>)

    const fileSearch = ref<string>('')
    const searchRecurively = ref<boolean>(false)

    const searchUpToDate = ref<boolean>(true)

    const loading = ref<boolean>(false)

    function getSortedChildren(newChildren?: WeblensFile[]): WeblensFile[] | undefined {
        if (!children.value && !newChildren) {
            console.error('Get sorted children no children to sort')
            return
        }

        if (!newChildren) {
            newChildren = [...children.value!]
        }

        newChildren.sort(getSortFunc(sortCondition.value, sortDirection.value))

        return newChildren
    }

    const { data, error, status } = useAsyncData(
        'files-' + locationStore.activeFolderID,
        async () => {
            if (!user.value.isLoggedIn.isSet() || !locationStore.activeFolderID) {
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

            children.value = getSortedChildren(newChildren)!

            const parents = res.data.parents?.map((fInfo) => new WeblensFile(fInfo))
            const activeFile = new WeblensFile(res.data.self)
            return { activeFile: activeFile, children: newChildren, parents }
        },
        { watch: [user, () => locationStore.activeFolderID, () => locationStore.viewTimestamp], lazy: true },
    )

    watch([() => locationStore.isInTimeline], () => {
        setFileSearch('')
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

        children.value = getSortedChildren()
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
        children.value = [...newChildren] // Trigger reactivity
    }

    function setMovedFile(fileIDs: string[], moved: boolean) {
        for (const fileID of fileIDs) {
            if (moved) {
                movedFiles.value.add(fileID)
            } else {
                movedFiles.value.delete(fileID)
            }
        }

        movedFiles.value = new Set(movedFiles.value)
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
        sortDirection.value *= -1

        saveFoldersSettings()
    }

    function setFileShape(newFileShape: FileShape) {
        fileShape.value = newFileShape

        saveFoldersSettings()
    }

    function setSortCondition(newSortCondition: 'date' | 'filename' | 'size') {
        sortCondition.value = newSortCondition

        saveFoldersSettings()
    }

    function setDragging(newDragging: boolean) {
        dragging.value = newDragging
    }

    function setFileSearch(search: string) {
        fileSearch.value = search
        searchUpToDate.value = false
    }

    function setSearchRecurively(recursive: boolean) {
        searchRecurively.value = recursive
    }

    function setLoading(load: boolean) {
        loading.value = load
    }

    const searchResults = shallowRef<WeblensFile[] | undefined>()

    async function doSearch() {
        if (!fileSearch.value || fileSearch.value.trim() === '') {
            searchResults.value = undefined
            return
        }

        loading.value = true

        const res = await useWeblensAPI().FilesAPI.searchByFilename(fileSearch.value, locationStore.activeFolderID)
        const results = res.data.map((f) => {
            return new WeblensFile(f)
        })

        searchResults.value = getSortedChildren(results) ?? []

        searchUpToDate.value = true
        loading.value = false
    }

    // Computed Properties //
    const activeFile = computed(() => {
        return data.value?.activeFile
    })

    const parents = computed(() => {
        return data.value?.parents
    })

    // Watchers //
    watchEffect(() => {
        if (children.value) {
            children.value = getSortedChildren()
        }
    })

    // When the active folder changes, clear selected files and reinitialize folder settings
    watch(
        () => locationStore.activeFolderID,
        () => {
            selectedFiles.value = new Set()
            fileSearch.value = ''
            searchResults.value = undefined
            searchUpToDate.value = false

            initFolderSettings()

            sortCondition.value = foldersSettings.value[locationStore.activeFolderID]?.sortCondition ?? 'date'
            sortDirection.value = foldersSettings.value[locationStore.activeFolderID]?.sortDirection ?? 1
            fileShape.value = foldersSettings.value[locationStore.activeFolderID]?.fileShape ?? 'square'
        },
        { immediate: true },
    )

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

    const isSearching = computed(() => {
        return fileSearch.value.trim() !== ''
    })

    const files = computed(() => {
        if (isSearching.value) {
            return searchResults.value ?? []
        }

        if (!children.value) {
            return []
        }

        let files = children.value
        if (fileSearch.value !== '' && !searchRecurively.value) {
            const search = fileSearch.value.toLowerCase()

            files = files.filter((f) => {
                return f.GetFilename().toLowerCase().includes(search)
            })
        }

        files = files.filter((f) => !f.IsTrash())

        return files
    })

    return {
        files,

        activeFile,

        dragging,

        children,
        parents,
        status,
        error,

        loading,
        setLoading,

        movedFiles,
        selectedFiles,

        lastSelected,
        lastSelectedIndex,
        nextSelectedIndex,
        setNextSelectedIndex,
        clearNextSelectedIndex,

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

        fileSearch,
        isSearching,
        setFileSearch,
        searchRecurively,
        setSearchRecurively,
        searchResults,
        searchUpToDate,
        doSearch,
    }
})

export default useFilesStore
