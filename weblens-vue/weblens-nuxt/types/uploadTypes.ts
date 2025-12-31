export type rateSample = {
    totalBytes: number
    time: number
}

export type UploadInfo = {
    serverUploadID: string
    localUploadID: `${string}-${string}-${string}-${string}-${string}`
    name: string
    type: 'file' | 'folder'
    status: 'pending' | 'uploading' | 'completed' | 'failed'

    files: Record<string, FileUploadMetadata>

    totalSize: number
    uploadedSoFar: number

    rate: rateSample[]
    progressPercent: number

    startTime: number
    endTime: number

    error?: Error // Optional, only for failed status
}

export type FileUploadMetadata = {
    file?: File
    entry?: FileSystemEntry
    isDir: boolean
    folderID?: string
    parentID: string
    isTopLevel: boolean

    uploadID: string

    fileID?: string

    uploadedSoFar?: number
    chunks: Record<number, UploadChunkInfo>
}

export function setChunk(upload: UploadInfo, fileID: string, chunkKey: number, chunk: UploadChunkInfo): UploadInfo {
    if (!upload.files) {
        upload.files = {}
    }

    if (!upload.files[fileID]) {
        upload.files[fileID] = { chunks: {} } as FileUploadMetadata
    }

    if (!upload.files[fileID].chunks) {
        upload.files[fileID].chunks = {}
    }

    upload.files[fileID].chunks[chunkKey] = chunk

    return upload
}

export function fileUploadedSoFar(fileUpload: FileUploadMetadata): number {
    if (!fileUpload.chunks) {
        return 0
    }

    return Object.values(fileUpload.chunks).reduce((sum, chunk) => sum + (chunk.uploadedSoFar || 0), 0)
}

export function recomputeSoFar(upload: UploadInfo) {
    upload.uploadedSoFar = Object.values(upload.files).reduce((sum, file) => sum + fileUploadedSoFar(file), 0)
    upload.progressPercent = Math.round((upload.uploadedSoFar / upload.totalSize) * 100)
}

export function addRate(upload: UploadInfo) {
    if (!upload.rate) {
        upload.rate = []
    }

    while (
        upload.rate[upload.rate.length - 1] &&
        upload.rate[upload.rate.length - 1].totalBytes > upload.uploadedSoFar
    ) {
        console.warn('Popping reverse upload progress for', upload.localUploadID)
        upload.rate.pop()
    }

    upload.rate.push({ totalBytes: upload.uploadedSoFar, time: Date.now() })
    while (
        upload.rate[upload.rate.length - 1].time - upload.rate[0].time >
        10000 /* Always keep samples from last 10 seconds */
    ) {
        upload.rate.shift()
    }
}

export type UploadChunkInfo = {
    uploadedSoFar: number
    startByte: number
    endByte: number
}
