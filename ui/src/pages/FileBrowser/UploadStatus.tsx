import { IconFile, IconFolder, IconX } from '@tabler/icons-react'
import WeblensButton from '@weblens/lib/WeblensButton.tsx'
import WeblensProgress from '@weblens/lib/WeblensProgress.tsx'
import { humanFileSize } from '@weblens/util'
import { CSSProperties, useEffect, useMemo, useRef } from 'react'
import { VariableSizeList } from 'react-window'

import { SingleUpload, useUploadStatus } from '../../store/UploadStateControl'
import './style/uploadStatusStyle.scss'

function UploadCardWrapper({
    data,
    index,
    style,
}: {
    data: SingleUpload[]
    index: number
    style: CSSProperties
}) {
    const uploadMeta = data[index]
    return (
        <div style={style}>
            <UploadCard uploadMetadata={uploadMeta} />
        </div>
    )
}

function UploadCard({ uploadMetadata }: { uploadMetadata: SingleUpload }) {
    const { prog, statusText, speedStr, speedUnits } = useMemo(() => {
        const prog = uploadMetadata.getPercetageComplete()
        const [soFarString, soFarUnits] = humanFileSize(uploadMetadata.bytes)
        const [totalString, totalUnits] = humanFileSize(
            uploadMetadata.bytesTotal
        )

        let statusText = `${soFarString}${soFarUnits} of ${totalString}${totalUnits}`
        if (uploadMetadata.bytesTotal === 0) {
            statusText = 'Starting ...'
        } else if (uploadMetadata.complete) {
            const [totalString, totalUnits] = humanFileSize(
                uploadMetadata.bytesTotal
            )
            statusText = `${totalString}${totalUnits}`
        } else if (uploadMetadata.error) {
            statusText = 'Failed'
        }

        const speed = uploadMetadata.getSpeed()

        if (uploadMetadata.isDir) {
            if (uploadMetadata.complete) {
                statusText += ` (${uploadMetadata.filesTotal} files)`
            }
        }

        const [speedStr, speedUnits] = humanFileSize(speed)

        return { prog, statusText, speedStr, speedUnits }
    }, [
        uploadMetadata.chunks,
        uploadMetadata.complete,
        uploadMetadata.bytes,
        uploadMetadata.bytesTotal,
        uploadMetadata.files,
        uploadMetadata.error,
    ])

    return (
        <div className="flex h-max w-full flex-col gap-2 p-2">
            <div className="m-[1px] flex h-max min-h-[40px] shrink-0 flex-row items-center">
                <div className="flex h-max w-0 grow flex-col items-start justify-center">
                    <p className="w-full truncate font-semibold">
                        {uploadMetadata.friendlyName}
                    </p>
                    <div>
                        <p className="my-1 pr-[4px] text-sm text-nowrap text-(--color-text)">
                            {statusText}
                        </p>
                        {!uploadMetadata.complete && (
                            <p className="mt-1 pr-[4px] text-sm text-nowrap text-(--color-text)">
                                {speedStr} {speedUnits}/s
                            </p>
                        )}
                    </div>
                </div>
                {uploadMetadata.isDir && (
                    <IconFolder
                        className="text-(--color-text)"
                        style={{ minHeight: '25px', minWidth: '25px' }}
                    />
                )}
                {!uploadMetadata.isDir && (
                    <IconFile
                        className="text-(--color-text)"
                        style={{ minHeight: '25px', minWidth: '25px' }}
                    />
                )}
            </div>

            {/* <WeblensProgress */}
            {/*     value={prog} */}
            {/*     failure={Boolean(uploadMetadata.error)} */}
            {/*     loading={uploadMetadata.bytesTotal === 0} */}
            {/*     className="min-h-3" */}
            {/* /> */}
            {!uploadMetadata.complete && (
                <WeblensProgress
                    value={prog}
                    failure={Boolean(uploadMetadata.error)}
                    loading={uploadMetadata.bytesTotal === 0}
                    className="min-h-3"
                />
            )}
        </div>
    )
}

const UploadStatus = () => {
    const uploadsMap = useUploadStatus((state) => state.uploads)
    const clearUploads = useUploadStatus((state) => state.clearUploads)
    const listRef = useRef<VariableSizeList>()

    const uploads = useMemo<SingleUpload[]>(() => {
        const uploads: SingleUpload[] = []
        const childrenMap = new Map<string, SingleUpload[]>()
        for (const upload of Array.from(uploadsMap.values())) {
            if (upload.parent) {
                if (childrenMap.get(upload.parent)) {
                    childrenMap.get(upload.parent).push(upload)
                } else {
                    childrenMap.set(upload.parent, [upload])
                }
            } else {
                uploads.push(upload)
            }
        }
        uploads.sort((a, b) => {
            if (a.complete && !b.complete) {
                return 1
            } else if (!a.complete && b.complete) {
                return -1
            }

            const aVal = a.bytes / a.bytesTotal
            const bVal = b.bytes / b.bytesTotal
            if (aVal === bVal) {
                return 0
            } else if (aVal !== 1 && bVal === 1) {
                return -1
            } else if (bVal !== 1 && aVal === 1) {
                return 1
            } else if (aVal >= 0 && aVal <= 1) {
                return 1
            }

            return 0
        })

        return uploads
    }, [uploadsMap])

    useEffect(() => {
        listRef.current?.resetAfterIndex(0)
    }, [uploads])

    if (uploads.length === 0) {
        return null
    }

    const topLevelCount = Array.from(uploadsMap.values()).filter(
        (val) => !val.parent
    ).length

    let height = 0
    for (const upload of uploads) {
        if (upload.complete) {
            height += 70
        } else {
            height += 120
        }
        if (height > 250) {
            height = 250
            break
        }
    }

    return (
        <div className="upload-status-container">
            <div className="wl-static-card mb-1 flex h-max max-h-full w-full flex-col overflow-hidden rounded-sm p-2 pb-0">
                <div className="flex h-max min-h-[50px]">
                    <div className="h-max min-h-max w-full">
                        <VariableSizeList
                            ref={listRef}
                            itemCount={uploads.length}
                            height={height}
                            width={'100%'}
                            itemSize={(index) => {
                                const upload = uploads[index]
                                if (upload.complete) {
                                    return 70
                                }
                                return 120
                            }}
                            itemData={uploads}
                            overscanCount={5}
                        >
                            {UploadCardWrapper}
                        </VariableSizeList>
                    </div>
                </div>

                <div className="flex h-max w-full flex-row justify-center p-2">
                    <div className="flex h-full w-full flex-row items-center justify-between">
                        <p className="text-lg font-semibold text-(--color-text)">
                            Uploading {topLevelCount} item
                            {topLevelCount !== 1 ? 's' : ''}
                        </p>
                        <WeblensButton
                            Left={IconX}
                            squareSize={30}
                            onClick={clearUploads}
                        />
                    </div>
                </div>
            </div>
        </div>
    )
}

export default UploadStatus
