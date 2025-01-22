import { Divider } from '@mantine/core'
import { IconFile, IconFolder, IconX } from '@tabler/icons-react'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensProgress from '@weblens/lib/WeblensProgress'
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
        <div className="flex w-full flex-col p-2 gap-2">
            <div className="flex flex-row h-max min-h-[40px] shrink-0 m-[1px] items-center">
                <div className="flex flex-col h-max w-0 items-start justify-center grow">
                    <p className="truncate font-semibold w-full">
                        {uploadMetadata.friendlyName}
                    </p>
                    {/* {statusText && prog !== 100 && prog !== -1 && ( */}
                    <div>
                        <p className="text-[--wl-text-color] text-nowrap pr-[4px] text-sm my-1">
                            {statusText}
                        </p>
                        {!uploadMetadata.complete && (
                            <p className="text-[--wl-text-color] text-nowrap pr-[4px] text-sm mt-1">
                                {speedStr} {speedUnits}/s
                            </p>
                        )}
                    </div>
                </div>
                {uploadMetadata.isDir && (
                    <IconFolder
                        className="text-[--wl-text-color]"
                        style={{ minHeight: '25px', minWidth: '25px' }}
                    />
                )}
                {!uploadMetadata.isDir && (
                    <IconFile
                        className="text-[--wl-text-color]"
                        style={{ minHeight: '25px', minWidth: '25px' }}
                    />
                )}
            </div>

            {!uploadMetadata.complete && (
                <WeblensProgress
                    value={prog}
                    failure={Boolean(uploadMetadata.error)}
                    loading={uploadMetadata.bytesTotal === 0}
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
            <div className="flex flex-col h-max max-h-full w-full bg-[--wl-card-background] p-2 pb-0 mb-1 rounded overflow-hidden">
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

                <Divider h={2} w={'100%'} />
                <div className="flex flex-row justify-center w-full h-max p-2">
                    <div className="flex flex-row h-full w-full items-center justify-between">
                        <p className="text-[--wl-text-color] font-semibold text-lg">
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
