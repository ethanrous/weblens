import { Divider } from '@mantine/core'
import { IconFile, IconFolder, IconX } from '@tabler/icons-react'

import React, { useMemo } from 'react'
import { humanFileSize } from '../../util'

import './style/uploadStatusStyle.scss'
import '../../components/style.scss'
import { create, StateCreator } from 'zustand'
import WeblensButton from '../../components/WeblensButton'
import { WeblensProgress } from '../../components/WeblensProgress'

type chunkT = {
    bytesSoFar: number
    sizeTotal: number
    complete: boolean
    speed: number
}

export interface UploadStatusStateT {
    uploads: Map<string, SingleUpload>

    newUpload: (
        key: string,
        name: string,
        isDir: boolean,
        total: number,
        parent?: string
    ) => void
    createChunk: (key: string, chunkIndex: number, chunkSize: number) => void
    updateProgress: (
        key: string,
        chunk: number,
        progress: number,
        speed: number
    ) => void
    chunkComplete: (key: string, chunkIndex: number) => void
    setError: (key: string, error: string) => void
    clearUploads: () => void
}

class SingleUpload {
    key: string
    isDir: boolean
    friendlyName: string
    subProgress: number // bytes written in current chunk, files only
    bytes: number
    files: number
    total: number // total size in bytes of the file, or number of files in the dir
    // speed: { time: number; bytes: number }[]
    speed: number[]
    prevTime: number // milliseconds timestamp
    parent: string // For files if they have a directory parent at the top level
    complete: boolean

    chunks: chunkT[]
    prevBytes: number

    error: string

    constructor(
        key: string,
        name: string,
        isDir: boolean,
        total: number,
        parent?: string
    ) {
        this.key = key
        this.friendlyName = name
        this.isDir = isDir
        this.total = total
        this.parent = parent

        this.bytes = 0
        this.files = 0
        this.chunks = []

        this.prevBytes = 0
        this.speed = []
    }

    incFiles() {
        this.files += 1
        if (this.files == this.total) {
            this.complete = true
        }
    }

    addChunk(chunkNum: number, chunkSize: number): void {
        while (chunkNum >= this.chunks.length - 1) {
            this.chunks.push(null)
        }
        this.chunks[chunkNum] = {
            bytesSoFar: 0,
            sizeTotal: chunkSize,
            complete: false,
            speed: 0,
        }
    }

    // Bytes is the new number of bytes uploaded so far
    updateChunk(chunkIndex: number, bytes: number, speed: number): number {
        if (this.isDir && !this.prevTime) {
            this.prevTime = Date.now()
            this.bytes = bytes
            this.prevBytes = 0
        } else if (this.isDir) {
            this.bytes += bytes
            const now = Date.now()
            const msSinceLastSample = now - this.prevTime
            if (msSinceLastSample > 250) {
                this.addSpeed(
                    ((this.bytes - this.prevBytes) / msSinceLastSample) * 1000
                )
                this.prevBytes = this.bytes
                this.prevTime = now
            }
        } else {
            if (!this.chunks[chunkIndex]) {
                console.error(
                    'Cannot find chunk at index',
                    chunkIndex,
                    this.chunks
                )
                return 0
            }
            const newTime = Date.now()
            const difference = bytes - this.chunks[chunkIndex].bytesSoFar
            this.chunks[chunkIndex].bytesSoFar = bytes
            this.chunks[chunkIndex].speed = speed
            this.prevTime = newTime

            return difference
        }

        return 0
    }

    chunkComplete(chunkIndex: number) {
        this.chunks[chunkIndex].complete = true
    }

    areChunksComplete(): boolean {
        const allComplete =
            this.chunks.findIndex((c) => c && !c.complete) === -1
        if (allComplete) {
            this.complete = true
        }
        return allComplete
    }

    getSpeed() {
        if (this.speed.length === 0 || this.complete) {
            return 0
        }

        if (this.isDir) {
            return this.speed.reduce((acc, v) => acc + v, 0) / this.speed.length
        } else {
            const speed = this.chunks.reduce(
                (acc, c) => (!c || c.complete ? acc : c.speed + acc),
                0
            )
            return speed
        }
    }

    addSpeed(speed: number): void {
        if (this.speed.length > 5) {
            this.speed.shift()
        }
        this.speed.push(speed)
    }

    setError(error: string) {
        if (this.error) {
            console.error(
                'Trying to override upload error with another:',
                error
            )
            return
        }
        this.error = error
    }
}

function UploadCard({
    uploadMetadata,
    subUploads,
}: {
    uploadMetadata: SingleUpload
    subUploads: SingleUpload[]
}) {
    let prog = 0
    let statusText = ''
    let speed = 0
    if (uploadMetadata.isDir) {
        if (uploadMetadata.files === -1) {
            prog = -1
        } else {
            prog = (uploadMetadata.files / uploadMetadata.total) * 100
        }
        statusText = `${uploadMetadata.files} of ${uploadMetadata.total} files`
        speed = subUploads.reduce((acc, f) => f.getSpeed() + acc, 0)
    } else if (uploadMetadata.chunks.length !== 0) {
        const soFar = uploadMetadata.chunks.reduce(
            (acc, chunk) => acc + (chunk ? chunk.bytesSoFar : 0),
            0
        )

        prog = (soFar / uploadMetadata.total) * 100
        const [soFarString, soFarUnits] = humanFileSize(soFar)
        const [totalString, totalUnits] = humanFileSize(uploadMetadata.total)
        statusText = `${soFarString}${soFarUnits} of ${totalString}${totalUnits}`
        speed = uploadMetadata.getSpeed()
    }

    const [speedStr, speedUnits] = humanFileSize(speed)
    // console.log(uploadMetadata, uploadMetadata.error)

    return (
        <div className="flex w-full flex-col p-2">
            <div className="flex flex-row h-max min-h-[40px] shrink-0 m-[1px] items-center">
                <div className="flex flex-col h-max w-0 items-start justify-center grow">
                    <p className="truncate font-semibold text-white w-full">
                        {uploadMetadata.friendlyName}
                    </p>
                    {statusText && prog !== 100 && prog !== -1 && (
                        <div>
                            <p className="text-nowrap pr-[4px] text-sm mt-1">
                                {statusText}
                            </p>
                            <p className="text-nowrap pr-[4px] text-sm mt-1">
                                {speedStr} {speedUnits}/s
                            </p>
                        </div>
                    )}
                </div>
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
            </div>

            {!uploadMetadata.complete && (
                <WeblensProgress
                    value={prog}
                    failure={Boolean(uploadMetadata.error)}
                />
            )}
        </div>
    )
}

const UploadStatus = () => {
    const uploadsMap = useUploadStatus((state) => state.uploads)
    const clearUploads = useUploadStatus((state) => state.clearUploads)

    const uploadCards = useMemo(() => {
        const uploadCards = []

        const parents: SingleUpload[] = []
        const childrenMap = new Map<string, SingleUpload[]>()
        for (const upload of Array.from(uploadsMap.values())) {
            if (upload.parent) {
                if (childrenMap.get(upload.parent)) {
                    childrenMap.get(upload.parent).push(upload)
                } else {
                    childrenMap.set(upload.parent, [upload])
                }
            } else {
                parents.push(upload)
            }
        }
        parents.sort((a, b) => {
            const aVal = a.bytes / a.total
            const bVal = b.bytes / b.total
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

        for (const uploadMeta of parents) {
            uploadCards.push(
                <UploadCard
                    key={uploadMeta.key}
                    uploadMetadata={uploadMeta}
                    subUploads={childrenMap.get(uploadMeta.key)}
                />
            )
        }
        return uploadCards
    }, [uploadsMap])

    if (uploadCards.length === 0) {
        return null
    }

    const topLevelCount = Array.from(uploadsMap.values()).filter(
        (val) => !val.parent
    ).length

    return (
        <div className="upload-status-container">
            <div className="flex flex-col h-max max-h-full w-full bg-[#ffffff11] p-2 pb-0 mb-1 rounded overflow-hidden">
                <div className="flex no-scrollbar h-max min-h-[50px]">
                    <div className="h-max min-h-max w-full">{uploadCards}</div>
                </div>

                <Divider h={2} w={'100%'} />
                <div className="flex flex-row justify-center h-max p-2">
                    <div className="flex flex-row h-full w-[97%] items-center justify-between">
                        <p className="text-white font-semibold text-lg">
                            Uploading {topLevelCount} item
                            {topLevelCount !== 1 ? 's' : ''}
                        </p>
                        <WeblensButton Left={IconX} onClick={clearUploads} />
                    </div>
                </div>
            </div>
        </div>
    )
}

const UploadStatusControl: StateCreator<UploadStatusStateT, [], []> = (
    set,
    get
) => ({
    uploads: new Map<string, SingleUpload>(),

    newUpload: (
        key: string,
        name: string,
        isDir: boolean,
        total: number,
        parentId?: string
    ) => {
        set((state) => {
            const upload = new SingleUpload(key, name, isDir, total, parentId)

            state.uploads.set(key, upload)

            if (parentId) {
                const parent = state.uploads.get(parentId)
                if (!parent) {
                    console.error(
                        'Could not get parent of new upload: ',
                        parentId
                    )
                }
                parent.total += 1
            }

            return {
                uploads: new Map<string, SingleUpload>(state.uploads),
            }
        })
    },

    createChunk: (key: string, chunkIndex: number, chunkSize: number) => {
        set((state) => {
            const uploadToUpdate = state.uploads.get(key)

            if (!uploadToUpdate) {
                console.error('Trying to update upload with unknown key:', key)
                return
            }

            uploadToUpdate.addChunk(chunkIndex, chunkSize)

            return { uploads: new Map(state.uploads) }
        })
    },

    updateProgress: (
        key: string,
        chunk: number,
        progress: number,
        speed: number
    ) => {
        set((state) => {
            const uploadToUpdate = state.uploads.get(key)

            if (!uploadToUpdate) {
                console.error('Trying to update upload with unknown key:', key)
                return
            }

            const diff = uploadToUpdate.updateChunk(chunk, progress, speed)
            uploadToUpdate.addSpeed(speed)
            if (uploadToUpdate.parent) {
                state.uploads
                    .get(uploadToUpdate.parent)
                    .updateChunk(-1, diff, speed)
            }

            return { uploads: new Map(state.uploads) }
        })
    },

    chunkComplete: (key: string, chunkIndex: number) => {
        set((state) => {
            const uploadToUpdate = state.uploads.get(key)

            if (!uploadToUpdate) {
                console.error('Trying to update upload with unknown key:', key)
                return
            }

            uploadToUpdate.chunkComplete(chunkIndex)
            if (uploadToUpdate.parent && uploadToUpdate.areChunksComplete()) {
                state.uploads.get(uploadToUpdate.parent).incFiles()
            }

            return { uploads: new Map(state.uploads) }
        })
    },

    setError: (key: string, error: string) => {
        const upload = get().uploads.get(key)
        if (!upload) {
            console.error('Could not find upload with key', key)
            return
        }
        upload.setError(error)
        return { uploads: new Map(get().uploads) }
    },

    clearUploads: () => {
        set({ uploads: new Map() })
    },
})

export const useUploadStatus = create<UploadStatusStateT>()(UploadStatusControl)
export default UploadStatus
