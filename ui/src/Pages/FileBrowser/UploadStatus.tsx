import {
    Center,
    CloseButton,
    Divider,
    RingProgress,
    Text,
    Tooltip,
} from '@mantine/core'
import { IconCheck, IconFile, IconFolder, IconX } from '@tabler/icons-react'

import React, { Dispatch, useMemo, useReducer } from 'react'
import { humanFileSize } from '../../util'

import './style/uploadStatusStyle.scss'
import '../../components/style.scss'

function uploadReducer(state: UploadStateType, action) {
    switch (action.type) {
        case 'add_new': {
            const newUploadMeta: UploadMeta = {
                key: action.key,
                isDir: action.isDir,
                friendlyName: action.name,
                parent: action?.parent,
                progress: 0,
                subProgress: 0,
                total: action.size ? action.size : 0,
                speed: [],
                complete: false,
            }
            if (action.parent) {
                const parent = state.uploadsMap.get(action.parent)
                parent.total += 1
                state.uploadsMap.set(action.parent, parent)
            }
            state.uploadsMap.set(newUploadMeta.key, newUploadMeta)
            return { ...state, uploadsMap: new Map(state.uploadsMap) }
        }
        case 'finished_chunk': {
            if (!state.uploadsMap.get(action.key)) {
                console.error("Looking for upload key that doesn't exist")
                return { ...state }
            }

            const replaceItem = state.uploadsMap.get(action.key)
            replaceItem.subProgress = 0

            // if (action.speed && replaceItem.speed.push(action.speed) >= 15) {
            //     // replaceItem.speed.shift()
            // }

            if (
                !replaceItem.complete &&
                replaceItem.progress === replaceItem.total &&
                replaceItem.parent
            ) {
                const parent = state.uploadsMap.get(replaceItem.parent)
                parent.progress += 1
                replaceItem.complete = true

                state.uploadsMap.set(replaceItem.parent, parent)
            }

            state.uploadsMap.set(action.key, replaceItem)
            return { ...state, uploadsMap: new Map(state.uploadsMap) }
        }
        case 'update_progress': {
            const replaceItem = state.uploadsMap.get(action.key)
            if (!replaceItem) {
                console.error("Looking for upload key that doesn't exist")
                return { ...state }
            }

            replaceItem.progress += action.progress

            const now = Date.now()
            if (
                replaceItem.speed.push({
                    time: now,
                    bytes: replaceItem.progress,
                }) >= 100
            ) {
                replaceItem.speed.shift()
            }

            state.uploadsMap.set(action.key, replaceItem)
            return { ...state, uploadsMap: new Map(state.uploadsMap) }
        }
        case 'clear': {
            state.uploadsMap.clear()
            return { ...state, uploadsMap: new Map(state.uploadsMap) }
        }
        default: {
            console.error('Got unexpected upload status action', action.type)
            return { ...state }
        }
    }
}

type UploadMeta = {
    key: string
    isDir: boolean
    friendlyName: string
    subProgress: number // bytes written in current chunk, files only
    progress: number
    total: number // total size in bytes of the file, or number of files in the dir
    speed: { time: number; bytes: number }[]
    parent: string // For files if they have a directory parent at the top level
    complete: boolean
}
type UploadStateType = {
    uploadsMap: Map<string, UploadMeta>
}

export function useUploadStatus() {
    const [uploadState, uploadDispatch]: [UploadStateType, Dispatch<any>] =
        useReducer(uploadReducer, {
            uploadsMap: new Map<string, UploadMeta>(),
        })

    return { uploadState, uploadDispatch }
}

const getSpeed = (stamps: { time: number; bytes: number }[]) => {
    let speed = 0
    if (stamps.length !== 0) {
        speed =
            (stamps[stamps.length - 1].bytes - stamps[0].bytes) /
            ((stamps[stamps.length - 1].time - stamps[0].time) / 1000)
    }

    return speed
}

function UploadCard({ uploadMetadata }: { uploadMetadata: UploadMeta }) {
    let prog = 0
    let statusText = ''
    if (uploadMetadata.isDir) {
        if (uploadMetadata.progress === -1) {
            prog = -1
        } else {
            prog = (uploadMetadata.progress / uploadMetadata.total) * 100
        }
        statusText = `${uploadMetadata.progress} of ${uploadMetadata.total} files`
    } else if (uploadMetadata.subProgress || uploadMetadata.progress) {
        prog =
            ((uploadMetadata.subProgress + uploadMetadata.progress) /
                uploadMetadata.total) *
            100
        const [val, unit] = humanFileSize(getSpeed(uploadMetadata.speed), true)
        statusText = `${val}${unit}/s`
    }

    return (
        <div className="flex flex-row h-max min-h-[50px] shrink-0 p-1 m-[1px] items-center">
            {uploadMetadata.isDir && (
                <IconFolder
                    color="white"
                    style={{ minHeight: '25px', minWidth: '25px' }}
                />
            )}
            {!uploadMetadata.isDir && (
                <IconFile
                    color="white"
                    style={{ minHeight: '25px', minWidth: '25px' }}
                />
            )}

            <div className="flex flex-col h-max w-0 items-start justify-center p-2 grow">
                <p className="truncate font-semibold text-white w-full">
                    {uploadMetadata.friendlyName}
                </p>
                {statusText && prog !== 100 && prog !== -1 && (
                    <p className="text-nowrap pr-[4px] text-sm mt-1">
                        {statusText}
                    </p>
                )}
            </div>

            {uploadMetadata.progress === -1 && (
                <RingProgress
                    size={10}
                    thickness={1}
                    sections={[{ value: 100, color: 'red' }]}
                    style={{ justifySelf: 'flex-end' }}
                    label={
                        <Center>
                            <IconX color="red" />
                        </Center>
                    }
                />
            )}
            {prog >= 0 && prog < 100 && (
                <RingProgress
                    size={35}
                    thickness={5}
                    style={{ justifySelf: 'flex-end' }}
                    sections={[{ value: prog, color: '#4444ff' }]}
                />
            )}
            {prog === 100 && (
                <RingProgress
                    sections={[{ value: prog, color: '#44ee44' }]}
                    size={30}
                    thickness={4}
                    style={{ justifySelf: 'flex-end' }}
                    label={
                        <Center>
                            <IconCheck color="white" />
                        </Center>
                    }
                />
            )}
        </div>
    )
}

const UploadStatus = ({
    uploadState,
    uploadDispatch,
}: {
    uploadState: UploadStateType
    uploadDispatch
}) => {
    const uploadCards = useMemo(() => {
        const uploadCards = []

        const uploads = Array.from(uploadState.uploadsMap.values())
            .filter((val) => !val.parent)
            .sort((a, b) => {
                const aVal = a.progress / a.total
                const bVal = b.progress / b.total
                if (aVal === bVal) {
                    return 0
                } else if (aVal !== 1 && bVal === 1) {
                    return -1
                } else if (bVal !== 1 && aVal === 1) {
                    return 1
                } else if (aVal >= 0 && aVal <= 1) {
                    return 1
                }

                return 0
            })

        for (const uploadMeta of uploads) {
            uploadCards.push(
                <UploadCard key={uploadMeta.key} uploadMetadata={uploadMeta} />
            )
        }
        return uploadCards
    }, [uploadState.uploadsMap])

    if (uploadState.uploadsMap.size === 0) {
        return null
    }

    const topLevelCount: number = Array.from(
        uploadState.uploadsMap.values()
    ).filter((val) => !val.parent).length
    return (
        <div className="upload-status-container">
            <div className="flex flex-col h-max max-h-full w-full bg-[#ffffff11] p-2 pb-0 mb-1 rounded overflow-hidden">
                <div className="flex no-scrollbar h-max min-h-[50px]">
                    <div className="h-max min-h-max w-full">{uploadCards}</div>
                </div>

                <Divider h={2} w={'100%'} />
                <div className="flex flex-row justify-center h-max p-2">
                    <div className="flex flex-row h-full w-[97%] items-center justify-between">
                        <Text c={'white'} fw={600} size="16px">
                            Uploading {topLevelCount} item
                            {topLevelCount !== 1 ? 's' : ''}
                        </Text>
                        <Tooltip label={'Clear'}>
                            <CloseButton
                                c={'white'}
                                variant="transparent"
                                onClick={() =>
                                    uploadDispatch({ type: 'clear' })
                                }
                            />
                        </Tooltip>
                    </div>
                </div>
            </div>
        </div>
    )
}

export default UploadStatus
