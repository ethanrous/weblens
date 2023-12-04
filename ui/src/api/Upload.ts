import API_ENDPOINT from "./ApiEndpoint"
import axios from 'axios'
import { useUploadStatus } from "../components/UploadStatus";

export type fileUploadMetadata = {
    file: FileSystemFileEntry
    parentId: string
    topLevelParentKey: string
    isTopLevel: boolean
}

function PromiseQueue(tasks = [], concurrentCount = 1) {
    this.total = tasks.length;
    this.todo = tasks;
    this.running = [];
    this.complete = [];
    this.count = concurrentCount;
}

PromiseQueue.prototype.runNext = function () {
    return ((this.running.length < this.count) && this.todo.length);
}

PromiseQueue.prototype.run = function () {
    while (this.runNext()) {
        const promiseFunc = this.todo.shift();
        promiseFunc().then(() => {
            this.complete.push(this.running.shift());
            this.run();
        });
        this.running.push(promiseFunc);
    }
}

async function getFile(fileEntry: FileSystemFileEntry): Promise<File> {
    try {
        return new Promise((resolve, reject) => fileEntry.file(resolve, reject));
    } catch (err) {
        console.error(err);
    }
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

async function readFile(file, blobStart?: number, blobEnd?: number) {
    return new Promise<string>(function (resolve, reject) {
        let fr = new FileReader();

        fr.onload = function () {
            resolve(fr.result.toString())
            // resolve({ name: file.name, item64: fr.result.toString() });
        };

        fr.onerror = () => {
            reject(fr)
        }

        if (blobStart == undefined) {
            blobStart = 0
            blobEnd = file.size
        }
        fr.readAsDataURL(file.slice(blobStart, blobEnd));
    })
}

const CHUNK_SIZE = 20000000

async function doChunkedUpload(fileData: File, parentFolderId, filename, authHeader, onProgress?, onFinish?) {
    let uploadId: string = ""
    try {
        const fileSize = fileData.size;
        let offset = 0;

        while (offset < fileSize) {
            const chunk = await readFile(fileData, offset, (offset + CHUNK_SIZE))

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
    const file: File = await getFile(uploadMeta.file)
    const key: string = uploadMeta.parentId + uploadMeta.file.name
    const onFinish = () => { uploadDispatch({ type: "finished", key: key }); if (uploadMeta.isTopLevel) { dispatch({ type: "add_skeleton", filename: uploadMeta.file.name }) } }

    if (file.size > CHUNK_SIZE) {
        // Upload is too large, do chunked upload
        const onProgress = (bytesWritten, totalBytes, MBpS) => { uploadDispatch({ type: "set_progress", key: key, progress: 100 * bytesWritten / totalBytes, speed: Math.trunc(MBpS) }) }
        return async () => await doChunkedUpload(file, uploadMeta.parentId, uploadMeta.file.name, authHeader, onProgress, onFinish)

    } else {
        // Upload is small enough for single upload
        const onProgress = (p) => { uploadDispatch({ type: "set_progress", key: key, progress: p.progress * 100 }) }
        return async () => await readFile(file).then(async (file64: string) => { await PostFile(file64, uploadMeta.parentId, uploadMeta.file.name, authHeader, onProgress, onFinish) })
    }
}

async function Upload(filesMeta: fileUploadMetadata[], authHeader, uploadDispatch, dispatch) {
    let uploads: (() => Promise<void>)[] = []

    // filesMeta.map((val) => {
    //     if (val.isTopLevel) {
    //         const key: string = val.parentId + val.file.name
    //         uploadDispatch({ type: 'add_new', isDir: val.file.isDirectory, key: key, name: val.file.name })
    //     }
    // })

    for (const meta of filesMeta) {
        const key: string = meta.parentId + meta.file.name
        if (meta.isTopLevel) {
            uploadDispatch({ type: 'add_new', isDir: meta.file.isDirectory, key: key, name: meta.file.name })
        }

        uploadDispatch({ type: 'add_new', isDir: meta.file.isDirectory, key: key, name: meta.file.name, parent: meta.topLevelParentKey })
        if (meta.file.isDirectory) {
            continue
        }
        const task = await singleUploadPromise(meta, authHeader, uploadDispatch, dispatch)
        uploads.push(task)
    }
    const taskQueue = new PromiseQueue(uploads, 5)
    taskQueue.run()
}

export default Upload