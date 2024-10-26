import { create, StateCreator } from 'zustand'

type chunkT = {
    bytesSoFar: number
    sizeTotal: number
    complete: boolean
    speed: number
}

export class SingleUpload {
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
