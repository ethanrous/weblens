import { WsAction, useWebsocketStore } from '@weblens/api/Websocket'
import { StateCreator, create } from 'zustand'

type chunkT = {
    bytesSoFar: number
    sizeTotal: number
    complete: boolean
    speed: number
}

export class SingleUpload {
    key: string
    uploadTaskId: string
    isDir: boolean
    friendlyName: string
    subProgress: number // bytes written in current chunk, files only
    bytes: number // number of bytes uploaded so far
    files: number // number of files uploaded so far
    bytesTotal: number // total size in bytes of the upload
    filesTotal: number // total number of files in the upload
    speed: { time: number; bytes: number; speed: number }[]
    parent: string // For files if they have a directory parent at the top level
    complete: boolean

    chunks: chunkT[]

    error: string

    constructor({
        key,
        name,
        isDir,
        totalBytes,
        uploadTaskId,
        parentId,
    }: {
        key: string
        name: string
        isDir: boolean
        totalBytes: number
        uploadTaskId: string
        parentId?: string
    }) {
        this.key = key
        this.uploadTaskId = uploadTaskId
        this.friendlyName = name
        this.isDir = isDir
        this.bytesTotal = totalBytes
        this.filesTotal = 0
        this.parent = parentId

        this.bytes = 0
        this.files = 0
        this.chunks = []

        this.speed = []
    }

    incFiles() {
        this.files += 1
        if (this.files === this.filesTotal) {
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
    updateChunk(chunkIndex: number, bytes: number): number {
        let speed = 0
        if (this.speed.length > 0) {
            const tail = this.speed[Math.max(this.speed.length - 6, 0)]
            const head = this.speed[this.speed.length - 1]
            const timeDiff = head.time - tail.time
            if (timeDiff !== 0) {
                speed = ((head.bytes - tail.bytes) / timeDiff) * 1000
            }
        }

        let difference = 0
        if (this.isDir) {
            this.bytes += bytes

            this.speed.push({
                time: Date.now(),
                bytes: this.bytes,
                speed: speed,
            })
        } else {
            if (!this.chunks[chunkIndex]) {
                console.error(
                    'Cannot find chunk at index',
                    chunkIndex,
                    this.chunks
                )
                return 0
            }
            difference = bytes - this.chunks[chunkIndex].bytesSoFar
            this.chunks[chunkIndex].bytesSoFar = bytes
            this.chunks[chunkIndex].speed = speed

            this.bytes += difference
        }

        this.speed.push({
            time: Date.now(),
            bytes: this.bytes,
            speed: speed,
        })

        return difference
    }

    getPercetageComplete(): number {
        if (this.bytesTotal === 0) {
            return 0
        }

        return (this.bytes / this.bytesTotal) * 100
    }

    chunkComplete(chunkIndex: number) {
        this.chunks[chunkIndex].complete = true
        this.areChunksComplete()
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

        const sampleCount = Math.min(this.speed.length, 20)

        return (
            this.speed
                .slice(-sampleCount)
                .map((s) => s.speed)
                .reduce((a, b) => a + b, 0) / sampleCount
        )
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

export interface UploadStatusStateT {
    uploads: Map<string, SingleUpload>

    newUpload: (
        key: string,
        uploadTaskId: string,
        name: string,
        isDir: boolean,
        totalBytes: number,
        parent?: string
    ) => void
    createChunk: (key: string, chunkIndex: number, chunkSize: number) => void
    updateProgress: (key: string, chunk: number, progress: number) => void
    chunkComplete: (key: string, chunkIndex: number) => void
    setError: (key: string, error: string) => void
    readError: (key: string) => string
    clearUploads: () => void
}
const UploadStatusControl: StateCreator<UploadStatusStateT, [], []> = (
    set,
    get
) => ({
    uploads: new Map<string, SingleUpload>(),

    newUpload: (
        key: string,
        uploadTaskId: string,
        name: string,
        isDir: boolean,
        totalBytes: number,
        parentId?: string
    ) => {
        set((state) => {
            const upload = new SingleUpload({
                key,
                uploadTaskId,
                name,
                isDir,
                totalBytes,
                parentId,
            })

            state.uploads.set(key, upload)

            if (parentId) {
                const parent = state.uploads.get(parentId)
                if (!parent) {
                    console.error(
                        'Could not get parent of new upload: ',
                        parentId
                    )
                }
                parent.filesTotal += 1
                parent.bytesTotal += totalBytes
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

    updateProgress: (key: string, chunk: number, progress: number) => {
        set((state) => {
            const uploadToUpdate = state.uploads.get(key)

            if (!uploadToUpdate) {
                console.error('Trying to update upload with unknown key:', key)
                return
            }

            const diff = uploadToUpdate.updateChunk(chunk, progress)
            if (uploadToUpdate.parent) {
                state.uploads.get(uploadToUpdate.parent).updateChunk(-1, diff)
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
        set((state) => {
            let upload = state.uploads.get(key)
            if (upload?.parent) {
                upload = state.uploads.get(upload.parent)
            }
            if (!upload) {
                console.error('Could not find upload with key', key)
                return
            }

            if (!upload.error) {
                upload.setError(error)
            }
            state.uploads.set(upload.key, upload)
            return { uploads: new Map(state.uploads) }
        })
    },

    readError: (key: string): string => {
        const state = get()
        let upload = state.uploads.get(key)
        if (upload?.parent) {
            upload = state.uploads.get(upload.parent)
        }
        if (!upload) {
            throw new Error(`Could not find upload with key: ${key}`)
        }

        return upload.error
    },

    clearUploads: () => {
        set((state) => {
            for (const upload of state.uploads.values()) {
                console.log('Clearing upload', upload)
                if (upload.error) {
                    continue
                }

                if (upload.complete) {
                    if (upload.parent) {
                        const parent = state.uploads.get(upload.parent)
                        if (parent && !parent.complete) {
                            continue
                        }
                    }
                    console.log('Its complete', upload)
                    state.uploads.delete(upload.uploadTaskId)
                    continue
                }

                upload.setError('Upload cancelled')

                useWebsocketStore.getState().wsSend({
                    action: WsAction.CancelTask,
                    subscribeKey: upload.uploadTaskId,
                })
            }
            return {
                uploads: new Map(state.uploads),
            }
        })
    },
})

export const useUploadStatus = create<UploadStatusStateT>()(UploadStatusControl)
