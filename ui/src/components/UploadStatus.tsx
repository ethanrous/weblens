
import { Card, Paper, Text, RingProgress, Box, ScrollArea, Button, CloseButton, Loader, Center, ActionIcon, Progress, Tooltip } from '@mantine/core';
import { IconCheck, IconFile, IconFolder, IconX } from '@tabler/icons-react';

import { memo, useEffect, useMemo, useReducer, useState } from "react"

function uploadReducer(state: UploadStateType, action) {
    switch (action.type) {
        case 'add_new': {
            if (state.uploadsMap.get(action.key)) {
                return { ...state }
            }
            const newUploadMeta: UploadMeta = { key: action.key, isDir: action.isDir, friendlyName: action.name, parent: action?.parent, progress: 0, totalFiles: 0, speed: 0 }
            if (action.parent) {
                const parent = state.uploadsMap.get(action.parent)
                parent.totalFiles += 1
                state.uploadsMap.set(action.parent, parent)
            }
            state.uploadsMap.set(newUploadMeta.key, newUploadMeta)
            return { ...state }
        }
        case 'set_progress': {
            if (!state.uploadsMap.get(action.key)) {
                return { ...state }
            }
            let newMap = new Map<string, UploadMeta>()

            state.uploadsMap.forEach((val, key) => {
                newMap.set(key, val)
            })

            let replaceItem = newMap.get(action.key)
            replaceItem.progress = action.progress
            replaceItem.speed = action.speed
            newMap.set(action.key, replaceItem)
            return { ...state, uploadsMap: newMap }
        }
        case 'clear': {
            return { ...state, uploadsMap: new Map<string, UploadMeta>() }
        }
        case 'finished': {
            const finishedItem = state.uploadsMap.get(action.key)
            if (finishedItem.parent) {
                const parent = state.uploadsMap.get(finishedItem.parent)
                parent.progress += 1
                state.uploadsMap.set(finishedItem.parent, parent)
            } else {
                finishedItem.progress = 100
                state.uploadsMap.set(action.key, finishedItem)
            }
            return { ...state }
        }
    }
}

type UploadMeta = {
    key: string
    isDir: boolean
    friendlyName: string
    progress: number // 0-100 for files, 0-{n} where n is number of files for directories
    totalFiles: number // only for directories, number of files in the dir
    speed: number
    parent: string // For files if they have a directory parent at the top level
}
type UploadStateType = {
    uploadsMap: Map<string, UploadMeta>
}

export function useUploadStatus() {
    const [uploadState, uploadDispatch]: [UploadStateType, React.Dispatch<any>] = useReducer(uploadReducer, {
        uploadsMap: new Map<string, UploadMeta>(),
    })

    return { uploadState, uploadDispatch }
}

function UploadCard({ uploadMetadata }: { uploadMetadata: UploadMeta }) {
    let prog = 0
    let statusText = ""
    if (uploadMetadata.isDir) {
        if (uploadMetadata.progress === -1) {
            prog = -1
        } else {
            prog = (uploadMetadata.progress / uploadMetadata.totalFiles) * 100
        }
        statusText = `${uploadMetadata.progress} of ${uploadMetadata.totalFiles} files`
    } else {
        prog = uploadMetadata.progress
        statusText = `${uploadMetadata.speed}MB/s`
    }

    return (
        <Card display={'flex'} pl={10} radius={0} style={{ width: "100%", backgroundColor: "#444444", height: '40px', minHeight: '40px', flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
            <Box display={'flex'} style={{ flexDirection: 'row' }}>
                {(uploadMetadata.isDir && (
                    <IconFolder color='white' />
                ))}
                {(!uploadMetadata.isDir && (
                    <IconFile color='white' />
                ))}
                <Text c="white" fw={500} pl={5}>{uploadMetadata.friendlyName}</Text>
            </Box>
            {(statusText && prog != 100 && prog != -1) && (
                <Text c="white" pl={10}>{statusText}</Text>
            )}
            {(uploadMetadata.progress === -1) && (
                <RingProgress
                    size={35}
                    thickness={5}
                    sections={[{ value: 100, color: 'red' }]}
                    label={
                        <Center>
                            <IconX color='red' />
                        </Center>
                    }
                />
            )}
            {(prog >= 0 && prog < 100) && (
                <RingProgress
                    size={35}
                    thickness={5}
                    sections={[{ value: prog, color: 'blue' }]}
                />
            )}
            {prog === 100 && (
                <RingProgress
                    sections={[{ value: prog, color: 'green' }]}
                    size={35}
                    thickness={5}
                    label={
                        <Center>

                            {/* <IconCheck style={{ width: "20px", height: rem(22) }} /> */}
                            <IconCheck color='white' />

                        </Center>
                    }
                />
            )}

        </Card>
    )
}

// const UploadStatus = ({ uploadState, uploadDispatch, count }: { uploadState: UploadStateType, uploadDispatch, count: number }) => {

const UploadStatus = ({ uploadState, uploadDispatch, count }: { uploadState: UploadStateType, uploadDispatch, count: number }) => {
    const uploadCards = useMemo(() => {
        let uploadCards = []
        const uploads = Array.from(uploadState.uploadsMap.values()).filter((val) => !val.parent)
        for (const uploadMeta of uploads) {
            uploadCards.push(<UploadCard key={uploadMeta.key} uploadMetadata={uploadMeta} />)
        }
        return uploadCards
    }, [uploadState.uploadsMap.values(), count])
    if (uploadState.uploadsMap.size === 0) {
        return null
    }
    const topLevelCount: number = Array.from(uploadState.uploadsMap.values()).filter((val) => val.parent === undefined).length
    return (
        <Paper pos={'fixed'} bottom={0} right={30} radius={10} style={{ backgroundColor: "#222222", zIndex: 2 }}>
            <Paper pt={8} pb={25} radius={10} mb={-10} ml={10} mr={10} style={{ backgroundColor: "transparent", display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Text c={'white'}>Uploading {topLevelCount} item{topLevelCount != 1 ? 's' : ''}</Text>
                <Tooltip label={"Close"}>
                    <CloseButton c={'white'} variant='transparent' onClick={() => uploadDispatch({ type: "clear" })} />
                </Tooltip>
            </Paper>
            <Card p={0} radius={0} style={{ backgroundColor: "transparent", height: "max-content", maxHeight: "200px", width: "400px" }}>
                <ScrollArea type='never' p={0} style={{ height: `${uploadCards.length * 40}px`, maxHeight: "200px" }}>
                    {uploadCards}
                </ScrollArea>
            </Card>
        </Paper>
    )
}

export default UploadStatus