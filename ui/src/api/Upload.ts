import axios from 'axios'
import API_ENDPOINT from './ApiEndpoint'
import { useUploadStatus } from '@weblens/pages/FileBrowser/UploadStateControl'
import { FileApi } from './FileBrowserApi'
import { ErrorHandler } from '@weblens/types/Types'

export type fileUploadMetadata = {
    file?: File
    entry?: FileSystemEntry
    isDir: boolean
    folderId?: string
    parentId: string
    topLevelParentKey: string
    isTopLevel: boolean
}

type PromiseFunc<T> = () => Promise<T>

class PromiseQueue<T> {
    total: number
    todo: PromiseFunc<T>[]
    running: Promise<void>[]
    results: T[]
    count: number

    constructor(tasks: PromiseFunc<T>[], concurrentCount = 1) {
        this.total = tasks.length
        this.todo = tasks
        this.running = []
        this.results = []
        this.count = concurrentCount
    }

    runNext() {
        return this.running.length < this.count && this.todo.length
    }

    async workerMain(): Promise<void> {
        while (this.todo.length) {
            const task = this.todo.shift()
            this.results.push(await task())
        }
    }

    queueMore(tasks: PromiseFunc<T>[]) {
        this.todo.push(...tasks)
    }

    async run() {
        for (let workerNum = 0; workerNum < this.count; workerNum++) {
            this.running.push(this.workerMain())
        }
        await Promise.all(this.running)
        return this.results
    }
}

async function readFile(blob: Blob) {
    return new Promise<string | ArrayBuffer>(function (resolve, reject) {
        const fr = new FileReader()
        fr.onload = () => {
            resolve(fr.result)
        }
        fr.onerror = () => {
            reject(new Error('failed to read file for upload'))
        }
        if (blob) {
            fr.readAsArrayBuffer(blob)
        }
    })
}

const UPLOAD_CHUNK_SIZE: number = 51200000
// const UPLOAD_CHUNK_SIZE: number = 2000000
const CONCURRENT_UPLOAD_COUNT = 6

async function uploadChunk(
    fileData: File,
    low: number,
    high: number,
    uploadId: string,
    fileId: string,
    onProgress: (bytesWritten: number, MBpS: number) => void,
    onFinish: (rate: number) => void
): Promise<void> {
    const chunk = await readFile(fileData.slice(low, high))
    const url = `${API_ENDPOINT}/upload/${uploadId}/file/${fileId}`

    const start = Date.now()
    await axios
        .put(url, chunk, {
            headers: {
                'Content-Range': `${low}-${high - 1}/${fileData.size}`,
                'Content-Type': 'application/octet-stream',
            },
            onUploadProgress: (e) => {
                onProgress(e.loaded, e.rate)
            },
        })
        .catch((r) => console.error(r))
    const end = Date.now()
    onFinish((high - low) / ((end - start) / 1000))
}

async function queueChunks(
    uploadMeta: fileUploadMetadata,
    // isPublic: boolean,
    uploadId: string,
    // shareId: string,
    taskQueue: PromiseQueue<void>
) {
    const file: File = uploadMeta.file
    const key: string = uploadMeta.parentId + uploadMeta.file.name

    if (useUploadStatus.getState().uploads.get(key).error) {
        console.warn(`Skipping upload with error: ${key}`)
        return
    }

    const res = await FileApi.addFileToUpload(uploadId, {
        parentFolderId: uploadMeta.parentId,
        newFileName: uploadMeta.file.name,
        fileSize: uploadMeta.file.size,
    }).catch(ErrorHandler)

    if (!res || res.status !== 200) {
        useUploadStatus.getState().setError(key, `Failed`)
        return
    }

    const fileId = res.data.fileId

    const chunkTasks: PromiseFunc<void>[] = []
    let chunkIndex = 0
    const chunkSize = UPLOAD_CHUNK_SIZE
    while (chunkIndex * chunkSize < file.size) {
        const thisChunkIndex = chunkIndex
        const chunkLowByte = thisChunkIndex * chunkSize // Copy offset to appease eslint
        const maxChunkHigh = (thisChunkIndex + 1) * chunkSize
        const chunkHighByte =
            maxChunkHigh >= file.size ? file.size : maxChunkHigh

        useUploadStatus
            .getState()
            .createChunk(key, thisChunkIndex, chunkHighByte - chunkLowByte)

        chunkTasks.push(
            async () =>
                await uploadChunk(
                    file,
                    chunkLowByte,
                    chunkHighByte,
                    uploadId,
                    fileId,
                    (bytesWritten: number, bytesPerSecond: number) => {
                        useUploadStatus
                            .getState()
                            .updateProgress(
                                key,
                                thisChunkIndex,
                                bytesWritten,
                                bytesPerSecond ? Math.trunc(bytesPerSecond) : 0
                            )
                    },
                    () => {
                        useUploadStatus
                            .getState()
                            .chunkComplete(key, thisChunkIndex)
                    }
                )
        )
        chunkIndex++
    }

    taskQueue.queueMore(chunkTasks)
}

async function Upload(
    filesMeta: fileUploadMetadata[],
    isPublic: boolean,
    shareId: string,
    rootFolder: string
) {
    const newUpload = useUploadStatus.getState().newUpload

    if (isPublic && !shareId) {
        throw new Error('Cannot do public upload without shareId')
    }

    const topDirs: string[] = []
    let hasTopFile = false

    const taskQueue = new PromiseQueue<void>([], CONCURRENT_UPLOAD_COUNT)
    // let taskQPromise

    let totalUploadSize = 0
    filesMeta.forEach((v) => {
        if (v.file.size) {
            totalUploadSize += v.file.size
        }
    })

    const res = await FileApi.startUpload({
        rootFolderId: rootFolder,
        totalUploadSize: totalUploadSize,
        chunkSize: UPLOAD_CHUNK_SIZE,
    })

    if (!res || res.status !== 200 || !res.data.uploadId) {
        console.error('Failed to start upload')
        return
    }

    for (const meta of filesMeta) {
        if (meta.file.name.startsWith('.')) {
            continue
        }

        const key: string = meta.folderId || meta.parentId + meta.file.name

        if (meta.isTopLevel) {
            newUpload(
                key,
                meta.file.name,
                meta.isDir,
                meta.isDir ? 0 : meta.file.size
            )
            if (meta.isDir) {
                topDirs.push(meta.folderId)
            }
            hasTopFile = hasTopFile || !meta.isDir
        } else {
            newUpload(
                key,
                meta.file.name,
                meta.isDir,
                meta.isDir ? 0 : meta.file.size,
                meta.topLevelParentKey
            )
        }
        if (meta.isDir) {
            continue
        }
        await queueChunks(meta, res.data.uploadId, taskQueue)
        // if (!taskQPromise) {
        //     taskQPromise = taskQueue.run()
        // }
    }
    await taskQueue.run()
}

export default Upload
