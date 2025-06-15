import { IconFile, IconFolder } from '@tabler/icons-react'
import { useQuery } from '@tanstack/react-query'
import { FileApi } from '@weblens/api/FileBrowserApi'
import { FileInfo } from '@weblens/api/swag'
import { useSessionStore } from '@weblens/components/UserInfo'
import WeblensInput from '@weblens/lib/WeblensInput.tsx'
import { useResize } from '@weblens/lib/hooks'
import { useFileBrowserStore } from '@weblens/store/FBStateControl'
import { ErrorHandler } from '@weblens/types/Types'
import { CSSProperties, useCallback, useEffect, useRef, useState } from 'react'
import { FixedSizeList as WindowList } from 'react-window'

enum SearchModeT {
    global,
    path,
    local,
}

type searchResultProps = {
    searchType: SearchModeT
    files: FileInfo[]
    highlightIndex: number
    folderInfo: FileInfo
    setHighlightIndex: (idx: number) => void
    visitHighlighted: (f: FileInfo) => void
}

function SearchResult({
    data,
    index,
    style,
}: {
    data: searchResultProps
    index: number
    style: CSSProperties
}) {
    const alreadyHere =
        data.files[index].isDir && data.files[index].id === data.folderInfo.id

    let preText: string
    if (data.searchType === SearchModeT.path) {
        preText =
            '~/' +
            data.files[index].portablePath.slice(
                data.files[index].portablePath.indexOf('/') + 1,
                data.files[index].portablePath.lastIndexOf(
                    '/',
                    data.files[index].portablePath.length - 2
                ) + 1
            )
    } else if (data.searchType === SearchModeT.local) {
        preText = './'
    }

    return (
        <div
            key={data.files[index].id}
            className="flex h-10 max-w-full cursor-pointer items-center justify-between gap-1 rounded-sm p-2"
            onMouseOver={() => data.setHighlightIndex(index)}
            onClick={(e) => {
                e.stopPropagation()
                const f = data.files[index]
                data.visitHighlighted(f)
            }}
            style={{
                ...style,
                backgroundColor:
                    data.highlightIndex === index
                        ? 'var(--color-background-secondary)'
                        : '',
                pointerEvents: alreadyHere ? 'none' : 'auto',
                color: alreadyHere
                    ? 'var(--color-text-tertiary)'
                    : 'var(--color-text-primary)',
            }}
        >
            {data.searchType !== SearchModeT.global && (
                <div className="flex flex-row items-center gap-1 select-none">
                    <span className="shrink-0 text-nowrap text-transparent">
                        {preText}
                    </span>
                    <span className="truncate text-nowrap text-inherit">
                        {data.files[index].portablePath.slice(
                            data.files[index].portablePath.lastIndexOf(
                                '/',
                                data.files[index].portablePath.length - 2
                            ) + 1
                        )}
                    </span>
                </div>
            )}
            {data.searchType === SearchModeT.global && (
                <div className="flex w-1 max-w-full grow text-inherit">
                    <span className="truncate text-nowrap select-none">
                        ~/
                        {data.files[index].portablePath.slice(
                            data.files[index].portablePath.indexOf('/') + 1
                        )}
                    </span>
                </div>
            )}
            {data.files[index].isDir && <IconFolder className="shrink-0" />}
            {!data.files[index].isDir && <IconFile className="shrink-0" />}
        </div>
    )
}

function SearchDialogue({
    text = '',
    visitFunc,
}: {
    text: string
    visitFunc: (l: string) => void
}) {
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const setIsSearching = useFileBrowserStore((state) => state.setIsSearching)
    const user = useSessionStore((state) => state.user)

    const [search, setSearch] = useState<string>(text)
    const [highlightIndex, setHighlightIndex] = useState(-1)
    const [containerRef, setContainerRef] = useState<HTMLDivElement>()
    const containerSize = useResize(containerRef)
    const resultsRef = useRef<WindowList>()

    const [files, setFiles] = useState<FileInfo[]>([])
    const { data: searchResult } = useQuery({
        queryKey: ['albums', search],
        queryFn: async () => {
            if (search.length === 0) {
                return []
            }

            let res: FileInfo[] | void = []

            if (search.startsWith('~/')) {
                const newSearch = search.replace(
                    '~/',
                    `USERS:${user.username}/`
                )
                res = await FileApi.autocompletePath(newSearch)
                    .then((res) => [...res.data.children, res.data.self])
                    .catch(ErrorHandler)
            } else if (search.startsWith('./')) {
                const path =
                    folderInfo.portablePath.replace('HOME', '~/') +
                    '/' +
                    search.slice(2)
                res = await FileApi.autocompletePath(path)
                    .then((res) => res.data.children)
                    .catch(ErrorHandler)
            } else {
                res = await FileApi.searchByFilename(search)
                    .then((res) => res.data)
                    .catch(ErrorHandler)
            }
            console.log('searchResult2', res)

            if (!res) {
                return []
            }

            return res
        },
    })

    const visitHighlighted = useCallback(
        (f: FileInfo) => {
            if (search === '~') {
                visitFunc(user.homeId)
            } else if (search === '..') {
                if (folderInfo.parentId) {
                    visitFunc(folderInfo.parentId)
                }
            } else if (!f) {
                return
            } else {
                visitFunc(f.id)
            }
            setIsSearching(false)
        },
        [search, folderInfo]
    )

    const selectNext = useCallback(
        (i: number) => {
            return Math.min(i + 1, files.length - 1)
        },
        [files]
    )

    useEffect(() => {
        if (search.length !== 0 && files.length !== 0) {
            setHighlightIndex(selectNext(-1))
        } else {
            setHighlightIndex(-1)
        }
    }, [selectNext, files.length, search.length])

    useEffect(() => {
        if (searchResult) {
            setFiles(searchResult)
        }
    }, [searchResult])

    let searchType = SearchModeT.global
    if (search.startsWith('~/')) {
        searchType = SearchModeT.path
    } else if (search.startsWith('./')) {
        searchType = SearchModeT.local
    }

    console.log('searchResult', files)

    const resultsData = {
        files,
        highlightIndex,
        setHighlightIndex,
        folderInfo,
        visitHighlighted,
        searchType,
    }

    return (
        <div
            ref={setContainerRef}
            className="flex h-full w-full"
            onKeyDown={(e) => {
                if (e.key === 'ArrowUp' || (e.ctrlKey && e.key === 'k')) {
                    e.preventDefault()
                    e.stopPropagation()

                    let newI = Math.max(highlightIndex - 1, 0)
                    let newF = files[newI]
                    while (newF.isDir && newF.id === folderInfo.id) {
                        newI--
                        if (newI === -1) {
                            return
                        }
                        newF = files[newI]
                    }

                    resultsRef.current.scrollToItem(newI, 'smart')
                    setHighlightIndex(newI)
                } else if (
                    e.key === 'ArrowDown' ||
                    (e.ctrlKey && e.key === 'j')
                ) {
                    e.preventDefault()
                    e.stopPropagation()
                    const newI = selectNext(highlightIndex)
                    resultsRef.current.scrollToItem(newI, 'smart')
                    setHighlightIndex(newI)
                } else if (e.key === 'Enter') {
                    e.stopPropagation()
                    const f = searchResult[highlightIndex]
                    visitHighlighted(f)
                } else if (e.key === 'Tab') {
                    e.preventDefault()
                    e.stopPropagation()
                    const portable = searchResult[highlightIndex].portablePath

                    setSearch('~/' + portable.slice(portable.indexOf('/') + 1))
                }
            }}
        >
            <div className="bg-background-primary relative flex h-max max-h-full flex-col items-center rounded-lg p-2">
                <div className="m-3 flex h-16 w-full shrink-0">
                    <WeblensInput
                        value={search}
                        placeholder={'Where To?'}
                        valueCallback={setSearch}
                        autoFocus
                        fillWidth
                        closeInput={() => setIsSearching(false)}
                        ignoreKeys={['ArrowDown', 'ArrowUp', 'Enter', 'Tab']}
                    />
                </div>
                <div className="relative flex h-full max-h-full w-full flex-col gap-1">
                    <WindowList
                        ref={resultsRef}
                        className="no-scrollbar"
                        height={Math.min(
                            files.length * 44,
                            containerSize.height - 104
                        )}
                        width={containerSize.width - 16}
                        itemSize={44}
                        itemCount={files.length}
                        itemData={resultsData}
                    >
                        {SearchResult}
                    </WindowList>
                </div>

                {search == '~' && (
                    <div className="flex items-center gap-1 text-sm">
                        <span className="bg-background h-max rounded-sm p-1 text-white">
                            Enter
                        </span>
                        <span>to go home</span>
                    </div>
                )}
                {search.length === 0 && (
                    <div className="flex flex-row items-center gap-1 text-sm text-nowrap select-none">
                        <span className="bg-background h-max rounded-sm p-1 text-white">
                            Tab
                        </span>
                        <span className="mr-1">to fill</span>
                        <span className="bg-background h-max rounded-sm p-1 text-white">
                            Enter
                        </span>
                        <span className="mr-1">to navigate</span>
                        <span className="bg-background h-max rounded-sm p-1 text-white">
                            ~/
                        </span>
                        <span className="mr-1">or</span>
                        <span className="bg-background h-max rounded-sm p-1 text-white">
                            ./
                        </span>
                        <span>to find by path</span>
                    </div>
                )}
            </div>
        </div>
    )
}

export default SearchDialogue
