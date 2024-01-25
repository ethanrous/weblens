
import { Card, Paper, Text, RingProgress, ScrollArea, CloseButton, Center, Tooltip, Space } from '@mantine/core';
import { IconCheck, IconFile, IconFolder, IconX } from '@tabler/icons-react';

import { useMemo, useReducer } from "react"
import { humanFileSize } from '../util';
import { FlexColumnBox, FlexRowBox } from '../Pages/FileBrowser/FilebrowserStyles';


function uploadReducer(state: UploadStateType, action) {
    switch (action.type) {
        case 'add_new': {
            // let existingUpload = state.uploadsMap.get(action.key)
            // if (existingUpload?.progress > 0) {
            //     return { ...state }
            // }
            const newUploadMeta: UploadMeta = { key: action.key, isDir: action.isDir, friendlyName: action.name, parent: action?.parent, progress: 0, total: action.size ? action.size : 0, speed: [] }
            if (action.parent) {
                const parent = state.uploadsMap.get(action.parent)
                parent.total += 1
                state.uploadsMap.set(action.parent, parent)
            }
            state.uploadsMap.set(newUploadMeta.key, newUploadMeta)
            return { ...state }
        }
        case 'update_progress': {
            if (!state.uploadsMap.get(action.key)) {
                console.error("Looking for upload key that doesn't exist")
                return { ...state }
            }
            let newMap = new Map<string, UploadMeta>()

            state.uploadsMap.forEach((val, key) => {
                newMap.set(key, val)
            })

            let replaceItem = newMap.get(action.key)
            replaceItem.progress += action.progress

            if (replaceItem.speed.push(action.speed) >= 6) {
                replaceItem.speed.shift()
            }

            if (replaceItem.progress === replaceItem.total && replaceItem.parent) {
                const parent = state.uploadsMap.get(replaceItem.parent)
                parent.progress += 1
                state.uploadsMap.set(replaceItem.parent, parent)
            }

            newMap.set(action.key, replaceItem)
            return { ...state, uploadsMap: newMap }
        }
        case 'clear': {
            return { ...state, uploadsMap: new Map<string, UploadMeta>() }
        }
        default: {
            console.error("Got unexpected upload status action", action.type)
            return { ...state }
        }
    }
}

type UploadMeta = {
    key: string
    isDir: boolean
    friendlyName: string
    progress: number // 0-100 for files, 0-{n} where n is number of files for directories
    total: number // total size in bytes of the file, or number of files in the dir
    speed: number[]
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

const average = (array) => { return (array.reduce((a, b) => a + b) / array.length) }

function UploadCard({ uploadMetadata }: { uploadMetadata: UploadMeta }) {
    let prog = 0
    let statusText = ""
    if (uploadMetadata.isDir) {
        if (uploadMetadata.progress === -1) {
            prog = -1
        } else {
            prog = (uploadMetadata.progress / uploadMetadata.total) * 100
        }
        statusText = `${uploadMetadata.progress} of ${uploadMetadata.total} files`
    } else if (uploadMetadata.progress) {
        prog = (uploadMetadata.progress / uploadMetadata.total) * 100
        const [val, unit] = humanFileSize(average(uploadMetadata.speed), true)
        statusText = `${val}${unit}/s`
    }

    return (
        <FlexRowBox style={{ width: 400, backgroundColor: "#444444", height: '40px' }}>
            <Space w={2} />
            {(uploadMetadata.isDir && (
                <IconFolder color='white' style={{ minHeight: '30px', minWidth: '30px' }} />
            ))}
            {(!uploadMetadata.isDir && (
                <IconFile color='white' style={{ minHeight: '30px', minWidth: '30px' }} />
            ))}
            <Text truncate="end" c="white" fw={500} pl={5} pr={5}>{uploadMetadata.friendlyName}</Text>

            <Space style={{ flexGrow: 1 }} />
            {(statusText && prog !== 100 && prog !== -1) && (
                <Text c="white" pr={5} style={{ minWidth: 75 }}>{statusText}</Text>
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
                            <IconCheck color='white' />
                        </Center>
                    }
                />
            )}

        </FlexRowBox>
    )
}

// const UploadStatus = ({ uploadState, uploadDispatch, count }: { uploadState: UploadStateType, uploadDispatch, count: number }) => {

const UploadStatus = ({ uploadState, uploadDispatch }: { uploadState: UploadStateType, uploadDispatch }) => {
    const uploadCards = useMemo(() => {
        let uploadCards = []

        const uploads = Array.from(uploadState.uploadsMap.values())
            .filter((val) => !val.parent)
            .sort((a, b) => {
                if ((a.progress / a.total) !== 1 && (b.progress / b.total) === 1) { return -1 }
                else if ((b.progress / b.total) !== 1 && (a.progress / a.total) === 1) { return 1 }
                return 0
            })

        for (const uploadMeta of uploads) {
            console.log(uploadMeta.progress / uploadMeta.total)
            uploadCards.push(<UploadCard key={uploadMeta.key} uploadMetadata={uploadMeta} />)
        }
        return uploadCards
    }, [uploadState.uploadsMap.values(), uploadState.uploadsMap.size])

    const height = useMemo(() => uploadCards.length * 40 < 225 ? 'max-content' : '225px', [uploadCards.length])

    if (uploadState.uploadsMap.size === 0) {
        return null
    }

    const topLevelCount: number = Array.from(uploadState.uploadsMap.values()).filter((val) => !val.parent).length
    return (
        <Paper pos={'fixed'} bottom={0} right={30} radius={10} style={{ backgroundColor: "#222222", zIndex: 2 }}>
            <Paper pt={8} pb={25} radius={10} mb={-10} ml={10} mr={10} style={{ backgroundColor: "transparent", display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Text c={'white'}>Uploading {topLevelCount} item{topLevelCount !== 1 ? 's' : ''}</Text>
                <Tooltip label={"Close"}>
                    <CloseButton c={'white'} variant='transparent' onClick={() => uploadDispatch({ type: "clear" })} />
                </Tooltip>
            </Paper>
            <Card p={0} radius={0} style={{ backgroundColor: "transparent", height: height, maxHeight: 225, width: "400px" }}>
                <ScrollArea type='never' mih={40} maw={400}>

                    {uploadCards}

                </ScrollArea>
            </Card>
        </Paper>
    )
}

export default UploadStatus