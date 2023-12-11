import API_ENDPOINT from "./ApiEndpoint"
import axios from 'axios'
import { useUploadStatus } from "../components/UploadStatus";
import { dispatchSync } from "./Websocket";

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

PromiseQueue.prototype.run = async function () {
    for (let workerNum = 0; workerNum < this.count; workerNum++) {
        this.running.push(this.workerMain())
    }
    await Promise.all(this.running)
    return this.results
}

const PostFile = async (file64, parentFolderId, filename, authHeader, onProgress, onFinish) => {
    await axios.post(
        `${API_ENDPOINT}/file`,
        JSON.stringify({
            parentFolderId: parentFolderId,
            filename: filename,
            file64: file64,
        }),
        {
            onUploadProgress: onProgress,
            headers: authHeader
        }
    )
    onFinish()
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

        fr.readAsDataURL(file);
    })
}

const CHUNK_SIZE = 20000000

async function doChunkedUpload(fileData: File, parentFolderId, filename, authHeader, onProgress?, onFinish?) {
    let uploadId: string = ""
    try {
        const fileSize = fileData.size;
        let offset = 0;

        const file64 = await readFile(fileData)
        while (offset < fileSize) {
            const chunk = file64.slice(offset, offset + CHUNK_SIZE)

            const formData = new FormData()
            formData.append("chunk", chunk)
            formData.append("offset", offset.toString())
            formData.append("totalSize", fileSize.toString())
            formData.append("filename", filename)
            formData.append("uploadId", uploadId)
            formData.append("parentFolderId", parentFolderId)

            const start = Date.now()
            let res = await axios.post(`${API_ENDPOINT}/file`, formData, {
                headers: { "Content-Type": "multipart/form-data", "Authorization": authHeader.Authorization, "Content-Range": `${offset}-${offset + CHUNK_SIZE - 1}/${fileSize}` }
            })
            const end = Date.now()
            if (uploadId === "") {
                uploadId = res.data.uploadId
            }
            offset += CHUNK_SIZE;
            onProgress(offset, fileSize, (CHUNK_SIZE / (end - start)) / 1000)
        }
        onFinish()
    } catch (error) {
        onProgress(-1, 100, 0)
        console.error("Error during chunk upload:", error);
    }
}

async function singleUploadPromise(uploadMeta: fileUploadMetadata, authHeader, uploadDispatch, dispatch) {
    const file: File = uploadMeta.file

    const key: string = uploadMeta.parentId + uploadMeta.file.name
    const onFinish = () => { uploadDispatch({ type: "finished", key: key }); if (uploadMeta.isTopLevel) { dispatch({ type: "add_skeleton", filename: uploadMeta.file.name }) } }

    if (file.size > CHUNK_SIZE) {
        // Upload is too large, do chunked upload
        const onProgress = (bytesWritten, totalBytes, MBpS) => { uploadDispatch({ type: "set_progress", key: key, progress: 100 * bytesWritten / totalBytes, speed: Math.trunc(MBpS) }) }
        return async () => {
            await doChunkedUpload(file, uploadMeta.parentId, uploadMeta.file.name, authHeader, onProgress, onFinish)
        }

    } else {
        // Upload is small enough for single upload
        const onProgress = (p) => { uploadDispatch({ type: "set_progress", key: key, progress: p.progress * 100 }) }
        return async () => {
            const file64 = await readFile(file)
            await PostFile(file64, uploadMeta.parentId, uploadMeta.file.name, authHeader, onProgress, onFinish)
        }
    }
}

async function Upload(filesMeta: fileUploadMetadata[], rootFolder, authHeader, uploadDispatch, dispatch, wsSend) {
    let uploads: (() => Promise<any>)[] = []

    let tlds: string[] = []
    let tlf = false

    for (const meta of filesMeta) {
        const key: string = meta?.folderId || meta.parentId + meta.file.name
        if (meta.isTopLevel) {
            uploadDispatch({ type: 'add_new', isDir: meta.isDir, key: key, name: meta.file.name })
            if (meta.isDir) {
                tlds.push(meta.folderId)
            }
            tlf = tlf || !meta.isDir
        }

        uploadDispatch({ type: 'add_new', isDir: meta.isDir, key: key, name: meta.file.name, parent: meta.topLevelParentKey })
        if (meta.isDir) {
            continue
        }
        const task = await singleUploadPromise(meta, authHeader, uploadDispatch, dispatch)
        uploads.push(task)
    }
    const taskQueue = new PromiseQueue(uploads, 5)
    await taskQueue.run()

    if (tlf) {
        dispatchSync(rootFolder, wsSend, false)
    }

    for (const tld of tlds) {
        dispatchSync(tld, wsSend, true)
    }
}

export default Upload