import { fetchJson } from '@weblens/api/ApiFetch'
import axios from 'axios'
import API_ENDPOINT from './ApiEndpoint'
import { useUploadStatus } from '@weblens/pages/FileBrowser/UploadStateControl'

export type fileUploadMetadata = {
    file: File
    isDir: boolean
    folderId?: string
    parentId: string
    topLevelParentKey: string
    isTopLevel: boolean
}

function PromiseQueue(tasks: (() => Promise<any>)[] = [], concurrentCount = 1) {
    this.total = tasks.length
    this.todo = tasks
    this.running = []
    this.results = []
    this.count = concurrentCount
}

PromiseQueue.prototype.runNext = function () {
    return this.running.length < this.count && this.todo.length
}

PromiseQueue.prototype.workerMain = async function () {
    while (this.todo.length) {
        const task = this.todo.shift()
        this.results.push(await task())
    }
}

PromiseQueue.prototype.queueMore = function (tasks: (() => Promise<any>)[]) {
    this.todo.push(...tasks)
}

PromiseQueue.prototype.run = async function () {
    for (let workerNum = 0; workerNum < this.count; workerNum++) {
        this.running.push(this.workerMain())
    }
    await Promise.all(this.running)
    return this.results
}

async function readFile(file) {
    return new Promise<any>(function (resolve, reject) {
        const fr = new FileReader()
        fr.onload = () => {
            resolve(fr.result)
        }
        fr.onerror = () => {
            reject(fr)
        }
        if (file) {
            fr.readAsArrayBuffer(file)
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
) {
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
    isPublic: boolean,
    uploadId: string,
    shareId: string,
    taskQueue
) {
    const file: File = uploadMeta.file
    const key: string = uploadMeta.parentId + uploadMeta.file.name

    const url = new URL(`${API_ENDPOINT}/upload/${uploadId}`)
    const body = {
        parentFolderId: uploadMeta.parentId,
        newFileName: uploadMeta.file.name,
        fileSize: uploadMeta.file.size,
    }

    if (useUploadStatus.getState().uploads.get(key).error) {
        console.warn(`Skipping upload with error: ${key}`)
        return
    }

    const data = await fetchJson<{ fileId: string }>(
        url.toString(),
        'POST',
        body
    )
    if (!data) {
        useUploadStatus.getState().setError(key, `Failed`)
    }

    const fileId = data.fileId

    const chunkTasks = []
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

async function NewUploadTask(
    rootFolderId: string,
    totalUploadSize: number,
    fileCount: number,
    isPublic: boolean,
    shareId: string
): Promise<string> {
    const url = new URL(`${API_ENDPOINT}/upload`)
    const body = {
        rootFolderId: rootFolderId,
        chunkSize: Math.min(
            UPLOAD_CHUNK_SIZE,
            Math.floor(totalUploadSize / fileCount)
        ),
        totalUploadSize: totalUploadSize,
    }
    if (isPublic) {
        url.searchParams.append('shareId', shareId)
    }
    return (await fetchJson<{ uploadId: string }>(url.toString(), 'POST', body))
        .uploadId
}

async function Upload(
    filesMeta: fileUploadMetadata[],
    isPublic: boolean,
    shareId: string,
    rootFolder
) {
    const newUpload = useUploadStatus.getState().newUpload

    if (isPublic && !shareId) {
        throw new Error('Cannot do public upload without shareId')
    }

    const topDirs: string[] = []
    let hasTopFile = false

    const taskQueue = new PromiseQueue([], CONCURRENT_UPLOAD_COUNT)
    let taskQPromise

    let totalUploadSize = 0
    let totalFileCount = 0
    filesMeta.forEach((v) => {
        if (v.file.size) {
            totalUploadSize += v.file.size
            totalFileCount += 1
        }
    })

    const uploadId = await NewUploadTask(
        rootFolder,
        totalUploadSize,
        totalFileCount,
        isPublic,
        shareId
    )
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
        await queueChunks(meta, isPublic, uploadId, shareId, taskQueue)
        if (!taskQPromise) {
            taskQPromise = taskQueue.run()
        }
    }
    await taskQPromise
}

export default Upload
