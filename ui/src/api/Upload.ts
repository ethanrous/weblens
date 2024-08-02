import API_ENDPOINT from './ApiEndpoint';
import axios from 'axios';
import { AuthHeaderT } from '../types/Types';

export type fileUploadMetadata = {
    file: File;
    isDir: boolean;
    folderId?: string;
    parentId: string;
    topLevelParentKey: string;
    isTopLevel: boolean;
};

function PromiseQueue(tasks: (() => Promise<any>)[] = [], concurrentCount = 1) {
    this.total = tasks.length;
    this.todo = tasks;
    this.running = [];
    this.results = [];
    this.count = concurrentCount;
}

PromiseQueue.prototype.runNext = function () {
    return this.running.length < this.count && this.todo.length;
};

PromiseQueue.prototype.workerMain = async function (workerNum: number) {
    while (this.todo.length) {
        const task = this.todo.shift();
        this.results.push(await task());
    }
};

PromiseQueue.prototype.queueMore = function (tasks: (() => Promise<any>)[]) {
    this.todo.push(...tasks);
};

PromiseQueue.prototype.run = async function () {
    for (let workerNum = 0; workerNum < this.count; workerNum++) {
        this.running.push(this.workerMain());
    }
    await Promise.all(this.running);
    return this.results;
};

async function readFile(file) {
    return new Promise<any>(function (resolve, reject) {
        const fr = new FileReader();
        fr.onload = () => {
            resolve(fr.result);
        };
        fr.onerror = () => {
            reject(fr);
        };
        if (file) {
            fr.readAsArrayBuffer(file);
        }
    });
}

const UPLOAD_CHUNK_SIZE: number = 51200000;
// const UPLOAD_CHUNK_SIZE: number = 500000;
const CONCURRENT_UPLOAD_COUNT = 6;

async function uploadChunk(
    fileData: File,
    low: number,
    high: number,
    uploadId: string,
    fileId: string,
    authHeader: AuthHeaderT,
    onProgress: (bytesWritten: number, MBpS: number) => void,
    onFinish: (rate: number) => void,
) {
    const chunk = await readFile(fileData.slice(low, high));
    const url = `${API_ENDPOINT}/upload/${uploadId}/file/${fileId}`;

    const start = Date.now();
    await axios
        .put(url, chunk, {
            headers: {
                Authorization: authHeader.Authorization,
                'Content-Range': `${low}-${high - 1}/${fileData.size}`,
                'Content-Type': 'application/octet-stream',
            },
            onUploadProgress: e => {
                onProgress(e.bytes, e.rate);
            },
        })
        .catch(r => console.error(r));
    const end = Date.now();
    onFinish((high - low) / ((end - start) / 1000));
}

async function queueChunks(
    uploadMeta: fileUploadMetadata,
    isPublic: boolean,
    uploadId: string,
    shareId: string,
    authHeader: AuthHeaderT,
    uploadDispatch,
    taskQueue,
) {
    const file: File = uploadMeta.file;
    const key: string = uploadMeta.parentId + uploadMeta.file.name;

    let url;
    let body;
    url = new URL(`${API_ENDPOINT}/upload/${uploadId}`);
    body = { parentFolderId: uploadMeta.parentId };
    // if (isPublic) {
    //     url = new URL(`${API_ENDPOINT}/public/upload`);
    //     url.searchParams.append("shareId", shareId);
    //     body = { parentFolderId: uploadMeta.parentId }
    // } else {
    // }

    body.newFileName = uploadMeta.file.name;
    body.fileSize = uploadMeta.file.size;

    const res = await fetch(url.toString(), {
        method: 'POST',
        headers: authHeader,
        body: JSON.stringify(body),
    })
        .then(async r => {
            return { data: await r.json(), code: r.status };
        })
        .catch(r => {
            return { code: r.response, data: null };
        });

    if (res.code === 409) {
        console.error(`Failed to greate file: ${file.name} already exists`);
        return;
    } else if (res.code !== 201) {
        console.error(`Failed to greate file: ${res.code}`);
        return;
    }

    const fileId = res.data.fileId;

    const chunkTasks = [];
    let offset = 0;
    // const chunkSize = Math.ceil(uploadMeta.file.size / (CONCURRENT_UPLOAD_COUNT / 2));
    // const chunkSize = uploadMeta.file.size;
    const chunkSize = UPLOAD_CHUNK_SIZE;
    while (offset < file.size) {
        const innerOffset = offset; // Copy offset to appease eslint
        const upperBound = offset + chunkSize >= file.size ? file.size : offset + chunkSize;
        chunkTasks.push(
            async () =>
                await uploadChunk(
                    file,
                    innerOffset,
                    upperBound,
                    uploadId,
                    fileId,
                    authHeader,
                    (bytesWritten: number, MBpS: number) =>
                        uploadDispatch({
                            type: 'update_progress',
                            key: key,
                            progress: bytesWritten,
                            speed: Math.trunc(MBpS),
                        }),
                    (rate: number) =>
                        uploadDispatch({
                            type: 'finished_chunk',
                            key: key,
                            chunkSize: chunkSize,
                            speed: rate,
                        }),
                ),
        );
        offset += chunkSize;
    }

    taskQueue.queueMore(chunkTasks);
}

async function NewUploadTask(
    rootFolderId: string,
    totalUploadSize: number,
    fileCount: number,
    isPublic: boolean,
    shareId: string,
    authHeader: AuthHeaderT,
): Promise<string> {
    let url;
    let init;
    init = {
        method: 'POST',
        body: JSON.stringify({
            rootFolderId: rootFolderId,
            chunkSize: Math.min(UPLOAD_CHUNK_SIZE, Math.floor(totalUploadSize / fileCount)),
            totalUploadSize: totalUploadSize,
        }),
    };
    if (isPublic) {
        url.searchParams.append('shareId', shareId);
    } else {
        url = new URL(`${API_ENDPOINT}/upload`);
        init.headers = authHeader;
    }
    return (await fetch(url.toString(), init).then(r => r.json())).uploadId;
}

async function Upload(
    filesMeta: fileUploadMetadata[],
    isPublic: boolean,
    shareId: string,
    rootFolder,
    authHeader: AuthHeaderT,
    uploadDispatch,
    wsSend: (action: string, content: any) => void,
) {
    if (isPublic && !shareId) {
        throw new Error('Cannot do public upload without shareId');
    }

    const topDirs: string[] = [];
    let hasTopFile = false;

    const taskQueue = new PromiseQueue([], CONCURRENT_UPLOAD_COUNT);
    let taskQPromise;

    let totalUploadSize = 0;
    let totalFileCount = 0;
    filesMeta.forEach(v => {
        if (v.file.size) {
            totalUploadSize += v.file.size;
            totalFileCount += 1;
        }
    });

    const uploadId = await NewUploadTask(rootFolder, totalUploadSize, totalFileCount, isPublic, shareId, authHeader);
    for (const meta of filesMeta) {
        if (meta.file.name.startsWith('.')) {
            continue;
        }

        const key: string = meta.folderId || meta.parentId + meta.file.name;

        if (meta.isTopLevel) {
            uploadDispatch({
                type: 'add_new',
                isDir: meta.isDir,
                key: key,
                name: meta.file.name,
                size: meta.isDir ? 0 : meta.file.size,
            });
            if (meta.isDir) {
                topDirs.push(meta.folderId);
            }
            hasTopFile = hasTopFile || !meta.isDir;
        } else {
            uploadDispatch({
                type: 'add_new',
                isDir: meta.isDir,
                key: key,
                name: meta.file.name,
                parent: meta.topLevelParentKey,
                size: meta.isDir ? 0 : meta.file.size,
            });
        }
        if (meta.isDir) {
            continue;
        }
        await queueChunks(meta, isPublic, uploadId, shareId, authHeader, uploadDispatch, taskQueue);
        if (!taskQPromise) {
            taskQPromise = taskQueue.run();
        }
    }
    await taskQPromise;
}

export default Upload;
