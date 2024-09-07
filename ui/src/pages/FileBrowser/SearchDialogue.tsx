import { IconFile, IconFolder } from '@tabler/icons-react'
import { useQuery } from '@tanstack/react-query'
import { useResize } from '@weblens/components/hooks'
import { useSessionStore } from '@weblens/components/UserInfo'
import WeblensInput from '@weblens/lib/WeblensInput'
import { WeblensFileInfo } from '@weblens/types/files/File'
import React, { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { FixedSizeList as WindowList, List } from 'react-window'
import { autocompletePath, searchFilenames } from '@weblens/api/ApiFetch'
import { useFileBrowserStore } from './FBStateControl'

enum SearchModeT {
    global,
    path,
    local,
}

function SearchResult({ data, index, style }) {
    const alreadyHere =
        data.files[index].isDir && data.files[index].id === data.folderInfo.id

    let preText
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
            onClick={() => {
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

export default function SearchDialogue() {
    const auth = useSessionStore((state) => state.auth)
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)
    const setIsSearching = useFileBrowserStore((state) => state.setIsSearching)
    const nav = useNavigate()

    const [search, setSearch] = useState<string>('')
    const [highlightIndex, setHighlightIndex] = useState(-1)
    const [containerRef, setContainerRef] = useState<HTMLDivElement>()
    const containerSize = useResize(containerRef)
    const resultsRef = useRef<List>()

    const [files, setFiles] = useState<WeblensFileInfo[]>([])
    const searchResult = useQuery({
        queryKey: ['albums', search],
        queryFn: async () => {
            // if (search.length < 2) {
            //     return []
            // }
            if (search.startsWith('~/')) {
                return (await autocompletePath(search, auth)).children
            } else if (search.startsWith('./')) {
                const path =
                    folderInfo.pathFromHome.replace('HOME', '~/') +
                    '/' +
                    search.slice(2)
                console.log(path)
                return (await autocompletePath(path, auth)).children
            } else {
                return await searchFilenames(search, auth)
            }
        },
    })

    const visitHighlighted = useCallback(
        (f: WeblensFileInfo) => {
            if (search === '~') {
                nav('/files/home')
            } else if (!f || f.id === folderInfo.id) {
                return
            } else if (f.isDir) {
                nav(`/files/${f.id}`)
            } else {
                nav(`/files/${f.parentId}?jumpTo=${f.id}`)
            }
            setIsSearching(false)
        },
        [search, folderInfo]
    )

    const selectNext = useCallback(
        (i) => {
            let newI = Math.min(i + 1, files.length - 1)
            let newF = files[newI]
            while (newF.isDir && newF.id === folderInfo.id) {
                if (newI === files.length - 1) {
                    return i
                }
                newI++
                newF = files[newI]
            }
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
        if (searchResult.data) {
            setFiles(searchResult.data)
        }
    }, [searchResult.data])

    let searchType = SearchModeT.global
    if (search.startsWith('~/')) {
        searchType = SearchModeT.path
    } else if (search.startsWith('./')) {
        searchType = SearchModeT.local
    }

    return (
        <div
            className="flex items-center justify-center w-screen h-screen absolute z-50 backdrop-blur-sm bg-[#00000088]"
            onKeyDown={(e) => {
                if (e.key === 'ArrowUp') {
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
                }
                if (e.key === 'ArrowDown') {
                    e.preventDefault()
                    e.stopPropagation()
                    const newI = selectNext(highlightIndex)
                    resultsRef.current.scrollToItem(newI, 'smart')
                    setHighlightIndex(newI)
                }
                if (e.key === 'Enter') {
                    e.stopPropagation()
                    const f = searchResult.data[highlightIndex]
                    visitHighlighted(f)
                }
                if (e.key === 'Tab') {
                    e.preventDefault()
                    e.stopPropagation()
                    const portable =
                        searchResult.data[highlightIndex].portablePath

                    setSearch('~/' + portable.slice(portable.indexOf('/') + 1))
                }
            }}
        >
            <div
                ref={setContainerRef}
                className="flex h-[50%] max-h-[50%] w-[500px]"
            >
                <div className="flex flex-col items-center w-max h-max max-h-full p-2 bg-[#1F1D2A] rounded-lg relative">
                    <div className="flex shrink-0 h-16 m-3 w-full">
                        <WeblensInput
                            value={search}
                            placeholder={'Where To?'}
                            valueCallback={setSearch}
                            autoFocus
                            fillWidth
                            closeInput={() => setIsSearching(false)}
                            ignoreKeys={[
                                'ArrowDown',
                                'ArrowUp',
                                'Enter',
                                'Tab',
                            ]}
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
                            itemData={{
                                files,
                                highlightIndex,
                                setHighlightIndex,
                                folderInfo,
                                visitHighlighted,
                                searchType,
                            }}
                        >
                            {SearchResult}
                        </WindowList>
                    </div>

                    {search == '~' && (
                        <div className="flex items-center gap-1 text-sm">
                            <div className="p-1 rounded bg-background h-max ">
                                <p>Enter</p>
                            </div>
                            <p>to go home</p>
                        </div>
                    )}
                    {search.length === 0 && (
                        <div className="flex flex-row items-center gap-1 select-none text-sm">
                            <div className="p-1 rounded bg-background h-max">
                                <p>Tab</p>
                            </div>
                            <p className="mr-1">to fill</p>
                            <div className="p-1 rounded bg-background h-max">
                                <p>Enter</p>
                            </div>
                            <p className="mr-1">to navigate</p>
                            <div className="p-1 rounded bg-background h-max">
                                <p>~/</p>
                            </div>
                            <p className="mr-1">or</p>
                            <div className="p-1 rounded bg-background h-max">
                                <p>./</p>
                            </div>
                            <p>to find by path</p>
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}
