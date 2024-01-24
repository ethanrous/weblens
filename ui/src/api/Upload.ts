import API_ENDPOINT from "./ApiEndpoint"
import axios from 'axios'
import { useUploadStatus } from "../components/UploadStatus";
import { dispatchSync } from "./Websocket";
import { notifications } from "@mantine/notifications";
import { dateFromFileData } from "../util";

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
    return new Promise<string>(function (resolve, reject) {
        let fr = new FileReader();

        fr.onload = function () {
            resolve(fr.result.toString())
            // resolve({ name: file.name, item64: fr.result.toString() });
        };

        fr.onerror = () => {
            reject(fr)
        }

        if (file) {
            fr.readAsDataURL(file);
        }
    })
}

const UPLOAD_CHUNK_SIZE: number = 51200000

async function uploadChunk(fileData: File, low: number, high: number, uploadId: string, authHeader, onProgress, onFinish) {
    let chunk = await readFile(fileData.slice(low, high))
    chunk = chunk.substring(chunk.indexOf(',') + 1, chunk.length)
    const url = `${API_ENDPOINT}/upload/${uploadId}`

    const start = Date.now() / 1000
    let res = await axios.put(url, chunk, {
        headers: { "Authorization": authHeader.Authorization, "Content-Range": `${low}-${high - 1}/${fileData.size}` },
        // onUploadProgress: (e) => {console.log(e.progress)}
    })
    // .catch(r => {notifications.show({title: `Failed to upload chunk of ${fileData.name}`, message: String(r), color: 'red'}); return String(r)})
    const end = Date.now() / 1000

    // Convert from base64 size to byte size
    let speed = ((UPLOAD_CHUNK_SIZE / 6) * 8) / (end - start)
    onProgress(high - low, fileData.size, speed)

    return null
}

async function queueChunks(uploadMeta: fileUploadMetadata, authHeader, uploadDispatch, taskQueue) {
    const file: File = uploadMeta.file
    const key: string = uploadMeta.parentId + uploadMeta.file.name

    const url = new URL(`${API_ENDPOINT}/upload`)
    url.searchParams.append("filename", uploadMeta.file.name)
    url.searchParams.append("parentFolderId", uploadMeta.parentId)
    let res = axios.post(url.toString(), null, {headers: authHeader})

    const onFinish = () => uploadDispatch({ type: "finished", key: key })
    const onProgress = (bytesWritten, totalBytes, MBpS) => { uploadDispatch({ type: "update_progress", key: key, progress: bytesWritten, speed: Math.trunc(MBpS) }) }

    const createFileResponse = await res
    console.log(createFileResponse)
    const uploadId = createFileResponse.data.uploadId

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

async function Upload(filesMeta: fileUploadMetadata[], rootFolder, authHeader, uploadDispatch, dispatch, wsSend: (action: string, content: any) => void) {
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
        await queueChunks(meta, authHeader, uploadDispatch, taskQueue)
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