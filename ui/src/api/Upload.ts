import { useUploadStatus } from '@weblens/store/UploadStateControl'
import { ErrorHandler } from '@weblens/types/Types'
import { AxiosProgressEvent } from 'axios'

import { FileApi } from './FileBrowserApi'
import { NewFileParams } from './swag'

export type FileUploadMetadata = {
    file?: File
    entry?: FileSystemEntry
    isDir: boolean
    folderId?: string
    parentId: string
    topLevelParentKey: string
    isTopLevel: boolean

    fileId?: string
}

type PromiseFunc<T> = () => Promise<T>

class PromiseQueue<T> {
    total: number
    todo: PromiseFunc<T>[]
    running: Promise<void>[]
    results: T[]
    count: number
    exit: boolean

    constructor(tasks: PromiseFunc<T>[], concurrentCount = 1) {
        this.total = tasks.length
        this.todo = tasks
        this.running = []
        this.results = []
        this.count = concurrentCount
        this.exit = false
    }

    runNext() {
        return (
            !this.exit && this.running.length < this.count && this.todo.length
        )
    }

    cancelQueue() {
        this.exit = true
    }

    async workerMain(): Promise<void> {
        while (this.todo.length) {
            const task = this.todo.shift()
            const result = await task()
            this.results.push(result)
        }
    }

    queueMore(tasks: PromiseFunc<T>[]) {
        this.todo.push(...tasks)
    }

    async run() {
        for (let workerNum = 0; workerNum < this.count; workerNum++) {
            this.running.push(this.workerMain())
        }
        console.debug('Running', this.running.length, 'upload tasks')
        await Promise.all(this.running)
        return this.results
    }
}

export const UPLOAD_CHUNK_SIZE: number = 51200000
const CONCURRENT_UPLOAD_COUNT = 4

function queueChunks(
    uploadMeta: FileUploadMetadata,
    uploadId: string,
    shareId: string,
    taskQueue: PromiseQueue<void>
) {
    const file: File = uploadMeta.file
    const key: string = uploadMeta.parentId + uploadMeta.file.name

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

        chunkTasks.push(async () => {
            if (useUploadStatus.getState().readError(key)) {
                console.warn(`Skipping upload with error: ${key}`)
                return
            }

            console.debug(
                `Uplading file ${file.name} ${uploadMeta.fileId} of size ${file.size}`
            )

            await FileApi.uploadFileChunk(
                uploadId,
                uploadMeta.fileId,
                file.slice(chunkLowByte, chunkHighByte) as File,
                shareId,
                {
                    headers: {
                        'Content-Range': `${chunkLowByte}-${chunkHighByte - 1}/${file.size}`,
                        'Content-Type': 'application/octet-stream',
                    },
                    onUploadProgress: (e: AxiosProgressEvent) => {
                        useUploadStatus
                            .getState()
                            .updateProgress(key, thisChunkIndex, e.loaded)
                    },
                }
            ).catch((err: Error) => {
                taskQueue.cancelQueue()
                useUploadStatus.getState().setError(key, String(err))
                ErrorHandler(err)
            })
            console.debug('Finished uploading', file.name)
            useUploadStatus.getState().chunkComplete(key, thisChunkIndex)
        })
        chunkIndex++
    }

    taskQueue.queueMore(chunkTasks)
}

async function Upload(
    filesMeta: FileUploadMetadata[],
    isPublic: boolean,
    shareId: string,
    uploadId: string
) {
    console.debug('Starting upload...')
    const newUpload = useUploadStatus.getState().newUpload

    if (isPublic && !shareId) {
        throw new Error('Cannot do public upload without shareId')
    }

    const topDirs: string[] = []
    let hasTopFile = false

    const taskQueue = new PromiseQueue<void>([], CONCURRENT_UPLOAD_COUNT)
    // let taskQPromise

    for (const meta of filesMeta) {
        if (!meta.file && !meta.entry.isDirectory) {
            ;(meta.entry as FileSystemFileEntry).file((f) => {
                meta.file = f
            })
        }
    }

    let count = 0
    while (
        filesMeta.findIndex((v) => !v.isDir && !v.file) !== -1 &&
        count < 1000
    ) {
        console.log(filesMeta[filesMeta.findIndex((v) => !v.file)])
        console.debug('Upload waiting...')
        await new Promise((r) => setTimeout(r, 10))
        count++
    }

    if (count >= 1000) {
        console.error('Upload failed: timeout waiting for file objects')
        return
    }

    filesMeta = filesMeta.filter(
        (meta) => meta.isDir || (meta.file && !meta.file.name.startsWith('.'))
    )

    for (const meta of filesMeta) {
        if (meta.isTopLevel) {
            const name = meta.file?.name ?? meta.entry.name
            const key: string = meta.folderId || meta.parentId + name
            newUpload(key, name, meta.isDir, meta.isDir ? 0 : meta.file.size)
            if (meta.isDir) {
                topDirs.push(meta.folderId)
            }
            hasTopFile = hasTopFile || !meta.isDir
        }
        if (!meta.isDir) {
            console.log(
                `File ${meta.file.name} ${meta.fileId} is of size ${meta.file.size}`
            )
        }
    }

    console.log('Collecting files to upload')
    const newFiles: NewFileParams[] = []
    filesMeta.forEach((v) => {
        if (v.isDir) {
            return
        }
        newFiles.push({
            parentFolderId: v.parentId,
            newFileName: v.file.name,
            fileSize: v.file.size,
        })
    })
    console.log('Adding all files to upload')

    const newFilesRes = await FileApi.addFilesToUpload(
        uploadId,
        {
            newFiles: newFiles,
        },
        shareId
    ).catch((err) => {
        console.error('Failed to add files to upload', err)
        for (const dir of topDirs) {
            console.log('Setting error for', dir)
            useUploadStatus.getState().setError(dir, String(err))
        }
    })

    if (!newFilesRes) {
        return
    }

    if (!newFilesRes || newFilesRes.status !== 201) {
        console.error('Failed to add files to upload', newFilesRes.data)
        return
    }

    if (newFilesRes.data.fileIds.length !== newFiles.length) {
        console.error('Mismatched fileIds length in upload')
        return
    }

    let index = 0
    for (const meta of filesMeta) {
        if (!meta.isDir && meta.file && meta.file.name.startsWith('.')) {
            index++
            continue
        }

        if (!meta.isDir) {
            meta.fileId = newFilesRes.data.fileIds[index]
            index++
        }

        if (!meta.isTopLevel) {
            const name = meta.file?.name ?? meta.entry.name
            const key: string = meta.folderId || meta.parentId + name
            newUpload(
                key,
                name,
                meta.isDir,
                meta.isDir ? 0 : meta.file.size,
                meta.topLevelParentKey
            )
        }

        if (meta.isDir) {
            continue
        }
        queueChunks(meta, uploadId, shareId, taskQueue)
    }
    await taskQueue.run()
}

export default Upload
