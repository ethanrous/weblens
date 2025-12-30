import type { AxiosProgressEvent } from 'axios'

import type { FileUploadMetadata } from '~/types/uploadTypes'
import { useWeblensAPI } from './AllApi'
import WeblensFile from '~/types/weblensFile.js'
import type { NewFileParams } from '@ethanrous/weblens-api'

const MAX_RETRIES = 5

type ChunkUploadContext = {
    soFar: number
    retriesRemaining: number
    chunkSize: number
}

async function pushChunkNew(
    serverUploadID: string,
    localUploadID: string,
    shareID: string,
    uploadMeta: FileUploadMetadata,
    chunkSize: number,
    chunkLowByte: number,
    chunkHighByte: number,
    remainingRetries: number,
): Promise<ChunkUploadContext> {
    if (useUploadStore().uploads.get(localUploadID)?.error) {
        throw new Error(`Upload ${localUploadID} cancelled. This shouldn't happen.`)
    }

    if (!uploadMeta.fileID) {
        throw new Error(`FileID falsy for upload: ${localUploadID}`)
    }

    if (!uploadMeta.file) {
        throw new Error(`File not found for upload: ${uploadMeta.parentID} ${uploadMeta.fileID}`)
    }

    const headers = {
        'Content-Range': `bytes=${chunkLowByte}-${chunkHighByte - 1}/${uploadMeta.file.size}`,
        'Content-Type': 'application/octet-stream',
    }

    try {
        const res = await useWeblensAPI().FilesAPI.uploadFileChunk(
            serverUploadID,
            uploadMeta.fileID,
            uploadMeta.file.slice(chunkLowByte, chunkHighByte) as File,
            shareID,
            {
                headers: headers,
                onUploadProgress: (e: AxiosProgressEvent) => {
                    useUploadStore().setUploadProgress(
                        localUploadID,
                        uploadMeta.fileID!,
                        chunkLowByte,
                        chunkHighByte - chunkLowByte,
                        e.loaded,
                    )
                },
            },
        )

        if (res.status !== 200) {
            throw new Error(
                `Failed to upload chunk starting at ${chunkLowByte} of ${uploadMeta.file.name}: ${res.status}`,
            )
        }

        useUploadStore().finishChunk(localUploadID, uploadMeta.fileID, chunkLowByte, chunkHighByte - chunkLowByte)
    } catch (err) {
        if (useUploadStore().uploads.get(localUploadID)?.error) {
            throw new Error(`Upload ${localUploadID} cancelled. This shouldn't happen.`)
        }

        if (remainingRetries === 0) {
            useUploadStore().failUpload(localUploadID, Error(String(err)))

            throw err
        }

        console.warn(
            `Retrying upload of ${uploadMeta.file.name} chunk starting at ${chunkLowByte} (${remainingRetries} retries remaining)`,
        )

        useUploadStore().resetChunk(localUploadID, uploadMeta.fileID!, chunkLowByte)

        return {
            soFar: chunkLowByte,
            retriesRemaining: remainingRetries - 1,
            chunkSize: Math.round(chunkSize / 2),
        }
    }

    return {
        soFar: chunkHighByte,
        retriesRemaining: Math.min(MAX_RETRIES, remainingRetries + 1),
        chunkSize: Math.min(chunkSize + chunkSize * 2, useUploadStore().uploadChunkSize),
    }
}

function queueChunksNew(
    uploadMeta: FileUploadMetadata,
    serverUploadID: string,
    localUploadID: string,
    shareID: string,
) {
    if (!uploadMeta.file) {
        throw new Error(`File not found for upload: ${uploadMeta.parentID} ${uploadMeta.fileID}`)
    }

    useUploadStore().uploadTaskQueue.runWithNext<ChunkUploadContext>({
        getNext: (ctx) => {
            if (ctx.soFar >= uploadMeta.file!.size) {
                return
            }

            const chunkSize = ctx.chunkSize
            const chunkHighByte =
                ctx.soFar + chunkSize >= uploadMeta.file!.size ? uploadMeta.file!.size : ctx.soFar + chunkSize

            return async () => {
                return pushChunkNew(
                    serverUploadID,
                    localUploadID,
                    shareID,
                    uploadMeta,
                    chunkSize,
                    ctx.soFar,
                    chunkHighByte,
                    ctx.retriesRemaining,
                )
            }
        },
        initCtx: {
            soFar: 0,
            retriesRemaining: MAX_RETRIES,
            chunkSize: useUploadStore().uploadChunkSize,
        },
        onFailure: (err) => {
            console.error(`Failed to upload chunk for ${uploadMeta.file?.name}:`, err)
            useUploadStore().failUpload(localUploadID, Error(String(err)))
        },
    })
}

async function tryUpload(
    filesMeta: FileUploadMetadata[],
    isPublic: boolean,
    shareID: string,
    serverUploadID: string,
    localUploadID: string,
) {
    try {
        await upload(filesMeta, isPublic, shareID, serverUploadID, localUploadID)
    } catch (err) {
        console.error(`Upload ${serverUploadID} / ${localUploadID} failed:`, err)
        useUploadStore().failUpload(serverUploadID, Error(String(err)))
    }
}

async function upload(
    filesMeta: FileUploadMetadata[],
    isPublic: boolean,
    shareID: string,
    serverUploadID: string,
    localUploadID: string,
) {
    if (isPublic && !shareID) {
        throw new Error('Cannot do public upload without shareID')
    }

    const metaPromises = filesMeta.map((meta) => {
        if (!meta.file && !meta.entry) {
            throw new Error('File and entry are both undefined for upload metadata')
        }

        new Promise((resolve, reject) => {
            if (!meta.file && !meta.entry!.isDirectory) {
                ;(<FileSystemFileEntry>meta.entry).file(
                    (f) => {
                        meta.file = f
                        resolve(meta)
                    },
                    (err) => {
                        console.error('Failed to get file from entry', err)
                        reject(err)
                    },
                )
            }
        })
    })

    await Promise.all(metaPromises)

    let count = 0
    while (filesMeta.findIndex((v) => !v.isDir && !v.file) !== -1 && count < 1000) {
        console.warn('Waiting for file objects to be resolved...')
        await new Promise((r) => setTimeout(r, 10))
        count++
    }

    if (count >= 1000) {
        console.error('Upload failed: timeout waiting for file objects')
        return
    }

    filesMeta = filesMeta.filter((meta) => meta.isDir || (meta.file && !meta.file.name.startsWith('.')))

    const newFiles: NewFileParams[] = filesMeta.map((v) => ({
        parentFolderID: v.parentID,
        newFileName: v.file?.name,
        fileSize: v.file?.size,
    }))

    const newFilesRes = await useWeblensAPI().FilesAPI.addFilesToUpload(serverUploadID, { newFiles: newFiles }, shareID)
    if (newFilesRes.status !== 201 || !newFilesRes.data.fileIDs) {
        throw new Error('Failed to add files to upload' + newFilesRes.statusText + JSON.stringify(newFilesRes.data))
    }

    if (newFilesRes.data.fileIDs.length !== newFiles.length) {
        throw new Error('Mismatched fileIDs length in upload')
    }

    for (const [index, meta] of filesMeta.entries()) {
        if (meta.isDir) {
            throw new Error('Directories should not be in filesMeta at this point')
        }
        meta.fileID = newFilesRes.data.fileIDs[index]

        queueChunksNew(meta, serverUploadID, localUploadID, shareID)
    }

    useUploadStore().addFilesToUpload(localUploadID, ...filesMeta)
}

function readAllFiles(reader: FileSystemDirectoryReader): Promise<FileSystemEntry[]> {
    return new Promise((resolve) => {
        const allEntries: FileSystemEntry[] = []

        function readEntriesRecursively() {
            reader.readEntries((entries) => {
                if (entries.length === 0) {
                    // No more entries, resolve the promise with all entries
                    resolve(allEntries)
                } else {
                    // Add entries to the array and call readEntriesRecursively again
                    allEntries.push(...entries)
                    readEntriesRecursively()
                }
            })
        }

        readEntriesRecursively()
    })
}

const excludedFileNames = ['.DS_Store']

async function addDir(
    fsEntry: FileSystemEntry,
    uploadID: string,
    parentFolderID: string,
    rootFolderID: string,
    isPublic: boolean,
    shareID: string,
): Promise<FileUploadMetadata[]> {
    if (fsEntry.isDirectory) {
        const newDirRes = await useWeblensAPI()
            .FilesAPI.addFilesToUpload(
                uploadID,
                {
                    newFiles: [
                        {
                            isDir: true,
                            parentFolderID: parentFolderID,
                            newFileName: fsEntry.name,
                        },
                    ],
                },
                shareID,
            )
            .catch((err) => {
                console.error('Failed to add files to upload', err)
            })

        if (!newDirRes || !newDirRes.data.fileIDs) {
            throw new Error('Failed to add directory to upload: response is undefined or has no fileIDs')
        }

        const folderID = newDirRes.data.fileIDs[0]
        if (!folderID) {
            return Promise.reject(new Error('Failed to create folder: no folderID'))
        }

        const allEntries = await readAllFiles((fsEntry as FileSystemDirectoryEntry).createReader())

        return (
            await Promise.all(
                allEntries.map((entry) => {
                    return addDir(entry, uploadID, folderID, rootFolderID, isPublic, shareID)
                }),
            )
        ).flat()
    } else {
        if (excludedFileNames.includes(fsEntry.name)) {
            return []
        }
        const e: FileUploadMetadata = {
            entry: fsEntry,
            uploadID: uploadID,
            parentID: parentFolderID,
            isDir: false,
            chunks: {},
            isTopLevel: parentFolderID === rootFolderID,
        }
        return [e]
    }
}

export async function HandleDrop(
    items: DataTransferItemList,
    rootFolderID: string,
    isPublic: boolean,
    shareID: string,
) {
    if (!items || items.length === 0) {
        console.error('No items to upload')
        return
    }

    const files = Array.from(items)
        .map((item) => item.webkitGetAsEntry())
        .filter((item) => item !== null)

    if (files.length === 0) {
        console.error('No valid files or directories to upload')
        return
    }

    const uploads = files.map((file) => {
        const localUploadID = window.crypto.randomUUID()

        return useUploadStore().startUpload({
            localUploadID: localUploadID,
            name: file.name,
            type: file.isDirectory ? 'folder' : 'file',
        })
    })

    await useUploadStore().uploadTaskQueue.addTask(async () => {
        const res = await useWeblensAPI()
            .FilesAPI.startUpload(
                {
                    rootFolderID: rootFolderID,
                    chunkSize: 1, // TODO: remove this
                },
                shareID,
            )
            .catch((err) => {
                console.error('Failed to start upload:', err)
            })

        if (!res) {
            throw new Error('Failed to start upload: no response')
        }

        const uploadID = res.data.uploadID

        if (!uploadID) {
            throw new Error('Failed to start upload: no uploadID returned')
        }

        for (const [index, file] of files.entries()) {
            const upload = uploads[index]
            if (!upload) {
                console.error('No upload metadata found for file:', file.name)
                continue
            }

            useUploadStore().setServerUpload(upload.localUploadID, uploadID)

            try {
                const uploadFiles = await addDir(file, uploadID, rootFolderID, rootFolderID, isPublic, shareID)

                if (uploadFiles.length !== 0) {
                    await tryUpload(uploadFiles, isPublic, shareID, uploadID, upload.localUploadID)
                }
            } catch (err) {
                useUploadStore().failUpload(upload.localUploadID, Error(String(err)))
            }
        }
    })
}

export async function HandleFileSelect(files: FileList, rootFolderID: string, isPublic: boolean, shareID: string) {
    if (!files || files.length === 0) {
        console.error('No files selected for upload')
        return
    }

    const dirs: Map<string, WeblensFile> = new Map()
    const uploads: FileUploadMetadata[] = []

    for (const file of files) {
        let parentID: string = rootFolderID

        if (file.webkitRelativePath !== '') {
            const pathParts = file.webkitRelativePath.split('/')
            for (const [index, pathPart] of pathParts.slice(0, -1).entries()) {
                const dirPath = pathParts.slice(0, index + 1).join('/')
                const existingDir = dirs.get(dirPath)
                if (existingDir) {
                    parentID = existingDir.id
                    continue
                }

                let parentDirID: string | undefined = undefined
                if (index === 0) {
                    parentDirID = rootFolderID
                } else {
                    parentDirID = dirs.get(pathParts.slice(0, index).join('/'))?.id
                }

                if (!parentDirID) {
                    console.error('Parent directory not found for:', dirPath)
                    return
                }

                const createRes = await useWeblensAPI().FoldersApi.createFolder({
                    parentFolderID: parentDirID,
                    newFolderName: pathPart,
                })

                const newDir = new WeblensFile(createRes.data)

                dirs.set(dirPath, newDir)
                parentID = newDir.id
            }
        }

        uploads.push({
            file: file,
            isDir: false,
            parentID: parentID,
            chunks: {},
            isTopLevel: parentID === rootFolderID,
            uploadID: '',
        })
    }

    if (dirs.size > 0) {
        const baseDirName = dirs.keys().find((dirPath) => {
            return !dirPath.includes('/')
        })
        if (!baseDirName) {
            console.error('No base directory found for uploads', dirs)
            return
        }

        const baseDir = dirs.get(baseDirName)
        const localUploadID = window.crypto.randomUUID()

        useUploadStore().startUpload({
            localUploadID,
            name: baseDir?.GetFilename(),
            type: 'folder',
        })

        await useUploadStore().uploadTaskQueue.addTask(async () => {
            const res = await useWeblensAPI()
                .FilesAPI.startUpload(
                    {
                        rootFolderID: rootFolderID,
                        chunkSize: 1, // TODO: remove this
                    },
                    shareID,
                )
                .catch((err) => {
                    console.error('Failed to start upload:', err)
                })

            if (!res) {
                throw new Error('Failed to start upload: no response')
            }

            const uploadID = res.data.uploadID

            if (!uploadID) {
                throw new Error('Failed to start upload: no uploadID returned')
            }

            try {
                await tryUpload(uploads, isPublic, shareID, uploadID, localUploadID)
            } catch (err) {
                useUploadStore().failUpload(localUploadID, Error(String(err)))
            }
        })
    } else {
        for (const upload of uploads) {
            const localUploadID = window.crypto.randomUUID()

            useUploadStore().startUpload({
                localUploadID: localUploadID,
                name: upload.file?.name || 'Unknown File',
                type: 'file',
            })

            await useUploadStore().uploadTaskQueue.addTask(async () => {
                const res = await useWeblensAPI()
                    .FilesAPI.startUpload(
                        {
                            rootFolderID: rootFolderID,
                            chunkSize: 1, // TODO: remove this
                        },
                        shareID,
                    )
                    .catch((err) => {
                        console.error('Failed to start upload:', err)
                    })

                if (!res) {
                    throw new Error('Failed to start upload: no response')
                }

                const uploadID = res.data.uploadID

                if (!uploadID) {
                    throw new Error('Failed to start upload: no uploadID returned')
                }

                try {
                    await tryUpload([upload], isPublic, shareID, uploadID, localUploadID)
                } catch (err) {
                    useUploadStore().failUpload(localUploadID, Error(String(err)))
                }
            })
        }
    }
}

export default upload
