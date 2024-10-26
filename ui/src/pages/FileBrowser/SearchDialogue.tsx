import { IconFile, IconFolder } from '@tabler/icons-react'
import { useQuery } from '@tanstack/react-query'
import { autocompletePath, searchFilenames } from '@weblens/api/ApiFetch'
import { useResize } from '@weblens/components/hooks'
import { useSessionStore } from '@weblens/components/UserInfo'
import WeblensInput from '@weblens/lib/WeblensInput'
import { useFileBrowserStore } from '@weblens/pages/FileBrowser/FBStateControl'
import { WeblensFileInfo } from '@weblens/types/files/File'
import { useCallback, useEffect, useRef, useState } from 'react'
import { FixedSizeList as WindowList, List } from 'react-window'

enum SearchModeT {
    global,
    path,
    local,
}

function SearchResult({ data, index, style }) {
    // const alreadyHere =
    //     data.files[index].isDir && data.files[index].id === data.folderInfo.id
    const alreadyHere = false

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
            className="flex rounded p-2 gap-1 items-center justify-between cursor-pointer h-10 max-w-full"
            onMouseOver={() => data.setHighlightIndex(index)}
            onClick={(e) => {
                e.stopPropagation()
                const f = data.files[index]
                data.visitHighlighted(f)
            }}
            style={{
                ...style,
                backgroundColor: data.highlightIndex === index ? '#2C2A38' : '',
                pointerEvents: alreadyHere ? 'none' : 'auto',
                color: alreadyHere ? '#888888' : 'white',
            }}
        >
            {data.searchType !== 0 && (
                <div className="flex flex-row items-center gap-1 select-none">
                    <p className="text-transparent shrink-0 text-nowrap">
                        {preText}
                    </p>
                    <p className="text-nowrap truncate">
                        {data.files[index].portablePath.slice(
                            data.files[index].portablePath.lastIndexOf(
                                '/',
                                data.files[index].portablePath.length - 2
                            ) + 1
                        )}
                    </p>
                </div>
            )}
            {data.searchType === SearchModeT.global && (
                <div className="flex max-w-full w-1 grow">
                    <p className="select-none text-nowrap truncate">
                        ~/
                        {data.files[index].portablePath.slice(
                            data.files[index].portablePath.indexOf('/') + 1
                        )}
                    </p>
                </div>
            )}
            {data.files[index].isDir && <IconFolder className="shrink-0" />}
            {!data.files[index].isDir && <IconFile className="shrink-0" />}
        </div>
    )
}

export default function SearchDialogue({
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
    const resultsRef = useRef<List>()

    const [files, setFiles] = useState<WeblensFileInfo[]>([])
    const { data: searchResult } = useQuery({
        queryKey: ['albums', search],
        queryFn: async () => {
            if (search.startsWith('~/')) {
                return (await autocompletePath(search)).children
            } else if (search.startsWith('./')) {
                const path =
                    folderInfo.portablePath.replace('HOME', '~/') +
                    '/' +
                    search.slice(2)
                return (await autocompletePath(path)).children
            } else {
                return await searchFilenames(search)
            }
        },
    })

    const visitHighlighted = useCallback(
        (f: WeblensFileInfo) => {
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
        (i) => {
            let newI = Math.min(i + 1, files.length - 1)
            // let newF = files[newI]
            // while (newF.isDir && newF.id === folderInfo.id) {
            //     if (newI === files.length - 1) {
            //         return i
            //     }
            //     newI++
            //     newF = files[newI]
            // }
            return newI
        },
        [files]
    )

    useEffect(() => {
        if (search.length !== 0 && files.length !== 0) {
            setHighlightIndex(selectNext(-1))
        } else {
            setHighlightIndex(-1)
        }
    }, [selectNext])

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
            className="flex w-full h-full"
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
            <div className="flex flex-col items-center h-max max-h-full p-2 bg-wl-barely-visible rounded-lg relative">
                <div className="flex shrink-0 h-16 m-3 w-full">
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
                <div className="flex flex-col max-h-full h-full gap-1 w-full relative">
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
                        <p className="p-1 rounded bg-background h-max text-white">
                            Enter
                        </p>
                        <p>to go home</p>
                    </div>
                )}
                {search.length === 0 && (
                    <div className="flex flex-row items-center gap-1 select-none text-sm text-nowrap">
                        <p className="p-1 rounded bg-background h-max text-white">
                            Tab
                        </p>
                        <p className="mr-1">to fill</p>
                        <p className="p-1 rounded bg-background h-max text-white">
                            Enter
                        </p>
                        <p className="mr-1">to navigate</p>
                        <p className="p-1 rounded bg-background h-max text-white">
                            ~/
                        </p>
                        <p className="mr-1">or</p>
                        <p className="p-1 rounded bg-background h-max text-white">
                            ./
                        </p>
                        <p>to find by path</p>
                    </div>
                )}
            </div>
        </div>
    )
}
