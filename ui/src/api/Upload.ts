import API_ENDPOINT from "./ApiEndpoint"
import axios, { AxiosError } from 'axios'
import { dispatchSync } from "./Websocket";
import { notifications } from "@mantine/notifications";

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
    return ((this.running.length < this.count) && this.todo.length);
}

PromiseQueue.prototype.workerMain = async function (workerNum: number) {
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
        let fr = new FileReader();
        fr.onload = () => {resolve(fr.result)}
        fr.onerror = () => {reject(fr)}
        if (file) {fr.readAsArrayBuffer(file)}
    })
}

const UPLOAD_CHUNK_SIZE: number = 51200000

async function uploadChunk(fileData: File, low: number, high: number, uploadId: string, authHeader, onProgress, onFinish) {
    let chunk = await readFile(fileData.slice(low, high))
    const url = `${API_ENDPOINT}/upload/${uploadId}`

    const start = Date.now()
    await axios.put(url, chunk, {
        headers: { "Authorization": authHeader.Authorization, "Content-Range": `${low}-${high - 1}/${fileData.size}`, "Content-Type": "application/octet-stream" },
        onUploadProgress: (e) => {onProgress(e.bytes, e.rate)}
    })
    const end = Date.now()
    onFinish(high - low, (high - low) / ((end - start) / 1000))

    return null
}

async function queueChunks(uploadMeta: fileUploadMetadata, isPublic: boolean, shareId: string, authHeader, uploadDispatch, taskQueue) {
    const file: File = uploadMeta.file
    const key: string = uploadMeta.parentId + uploadMeta.file.name

    let url
    if (isPublic) {
        url = new URL(`${API_ENDPOINT}/public/upload`)
        url.searchParams.append("shareId", shareId)
        url.searchParams.append("parentFolderId", uploadMeta.parentId)
    } else {
        url = new URL(`${API_ENDPOINT}/upload`)
        url.searchParams.append("parentFolderId", uploadMeta.parentId)
    }
    url.searchParams.append("filename", uploadMeta.file.name)
    let res = await axios.post(url.toString(), null, {headers: authHeader}).catch((r: AxiosError) => r.response)

    const onFinish = (chunkSize, rate) => uploadDispatch({ type: "finished_chunk", key: key, chunkSize: chunkSize, speed: rate })
    const onProgress = (bytesWritten, MBpS) => { uploadDispatch({ type: "update_progress", key: key, progress: bytesWritten, speed: Math.trunc(MBpS) }) }

    if (res.status === 409) {
        notifications.show({title: "Failed to upload", message: `${file.name} already exists`, color: 'red'})
        return
    } else if (res.status !== 201) {
        notifications.show({title: "Failed to upload", message: res.statusText, color: 'red'})
        return
    }
    const uploadId = res.data.uploadId

    let chunkTasks = []
    let offset = 0
    while (offset < file.size) {
        let innerOffset = offset // Copy offset to appease eslint
        let upperBound = offset + UPLOAD_CHUNK_SIZE >= file.size ? file.size : offset + UPLOAD_CHUNK_SIZE
        chunkTasks.push(async () => await uploadChunk(file, innerOffset, upperBound, uploadId, authHeader, onProgress, onFinish))
        offset += UPLOAD_CHUNK_SIZE
    }

    taskQueue.queueMore(chunkTasks)
}

async function Upload(filesMeta: fileUploadMetadata[], isPublic: boolean, shareId: string, rootFolder, authHeader, uploadDispatch, wsSend: (action: string, content: any) => void) {
    if (isPublic && !shareId) {
        throw new Error("Cannot do public upload without shareId");
    }

    let tlds: string[] = []
    let tlf = false

    const taskQueue = new PromiseQueue([], 5)
    for (const meta of filesMeta) {
        const key: string = meta.folderId || meta.parentId + meta.file.name

        if (meta.isTopLevel) {
            uploadDispatch({ type: 'add_new', isDir: meta.isDir, key: key, name: meta.file.name, size: meta.isDir ? 0 : meta.file.size })
            if (meta.isDir) {
                tlds.push(meta.folderId)
            }
            tlf = tlf || !meta.isDir
        } else {
            uploadDispatch({ type: 'add_new', isDir: meta.isDir, key: key, name: meta.file.name, parent: meta.topLevelParentKey, size: meta.isDir ? 0 : meta.file.size })
        }
        if (meta.isDir) {
            continue
        }
        await queueChunks(meta, isPublic, shareId, authHeader, uploadDispatch, taskQueue)
    }
    await taskQueue.run()

    if (tlf) {
        dispatchSync(rootFolder, wsSend, false, false)
    }

    for (const tld of tlds) {
        dispatchSync(tld, wsSend, true, false)
    }
}

export default Upload