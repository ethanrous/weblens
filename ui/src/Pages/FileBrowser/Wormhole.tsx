import { useParams } from 'react-router-dom'
import { DropSpot } from './FileBrowserStyles'
import { useCallback, useContext, useEffect, useState } from 'react'
import { GetWormholeInfo } from '../../api/FileBrowserApi'
import { UserContext } from '../../Context'
import { UserContextT } from '../../types/Types'
import { FileButton, Space, Text } from '@mantine/core'
import { notifications } from '@mantine/notifications'
import UploadStatus, { useUploadStatus } from './UploadStatus'
import { IconFolder, IconUpload } from '@tabler/icons-react'
import { HandleDrop, HandleUploadButton } from './FileBrowserLogic'

import './style/fileBrowserStyle.scss'
import { ShareDataT } from '../../classes/Share'

const UploadPlaque = ({
    wormholeId,
    uploadDispatch,
}: {
    wormholeId: string
    uploadDispatch
}) => {
    return (
        <div className="h-[45vh]">
            <FileButton
                onChange={(files) => {
                    HandleUploadButton(
                        files,
                        wormholeId,
                        true,
                        wormholeId,
                        { Authorization: '' },
                        uploadDispatch,
                        () => {}
                    )
                }}
                accept="file"
                multiple
            >
                {(props) => {
                    return (
                        <div className="flex bg-bottom-grey h-[20vh] w-[20vw] p-3 rounded justify-center">
                            <div
                                className="cursor-pointer h-max w-max"
                                onClick={() => {
                                    props.onClick()
                                }}
                            >
                                <IconUpload
                                    size={100}
                                    style={{ padding: '10px' }}
                                />
                                <Text
                                    size="20px"
                                    className="select-none font-semibold"
                                >
                                    Upload
                                </Text>
                                <Space h={4}></Space>
                                <Text size="12px" className="select-none">
                                    Click or Drop
                                </Text>
                            </div>
                        </div>
                    )
                }}
            </FileButton>
        </div>
    )
}

const WormholeWrapper = ({
    wormholeId,
    wormholeName,
    fileId,
    validWormhole,
    uploadDispatch,
    children,
}: {
    wormholeId: string
    wormholeName: string
    fileId: string
    validWormhole: boolean
    uploadDispatch
    children
}) => {
    const { authHeader }: UserContextT = useContext(UserContext)
    const [dragging, setDragging] = useState(0)
    const [dropSpotRef, setDropSpotRef] = useState(null)
    const handleDrag = useCallback(
        (e) => {
            e.preventDefault()
            if (e.type === 'dragenter' || e.type === 'dragover') {
                if (!dragging) {
                    setDragging(2)
                }
            } else if (dragging) {
                setDragging(0)
            }
        },
        [dragging]
    )

    return (
        <div className="wormhole-wrapper">
            <div
                ref={setDropSpotRef}
                style={{ position: 'relative', width: '98%', height: '98%' }}
                //                    See DirViewWrapper \/
                onMouseMove={() => {
                    if (dragging) {
                        setTimeout(() => setDragging(0), 10)
                    }
                }}
            >
                <DropSpot
                    onDrop={(e) =>
                        HandleDrop(
                            e.dataTransfer.items,
                            fileId,
                            [],
                            true,
                            wormholeId,
                            authHeader,
                            uploadDispatch,
                            () => {}
                        )
                    }
                    dropSpotTitle={wormholeName}
                    dragging={dragging}
                    dropAllowed={validWormhole}
                    handleDrag={handleDrag}
                    wrapperRef={dropSpotRef}
                />
                <div className="justify-center" onDragOver={handleDrag}>
                    {children}
                </div>
            </div>
        </div>
    )
}

export default function Wormhole() {
    const wormholeId = useParams()['*']
    const { authHeader }: UserContextT = useContext(UserContext)
    const [wormholeInfo, setWormholeInfo]: [
        wormholeInfo: ShareDataT,
        setWormholeInfo: any,
    ] = useState(null)
    const { uploadState, uploadDispatch } = useUploadStatus()

    useEffect(() => {
        if (wormholeId !== '') {
            GetWormholeInfo(wormholeId, authHeader)
                .then((v) => {
                    if (v.status !== 200) {
                        return Promise.reject(v.statusText)
                    }
                    return v.json()
                })
                .then((v) => {
                    setWormholeInfo(v)
                })
                .catch((r) => {
                    notifications.show({
                        title: 'Failed to get wormhole info',
                        message: String(r),
                        color: 'red',
                    })
                })
        }
    }, [wormholeId, authHeader])
    const valid = Boolean(wormholeInfo)

    return (
        <div>
            <UploadStatus
                uploadState={uploadState}
                uploadDispatch={uploadDispatch}
            />
            <WormholeWrapper
                wormholeId={wormholeId}
                wormholeName={wormholeInfo?.shareName}
                fileId={wormholeInfo?.fileId}
                validWormhole={valid}
                uploadDispatch={uploadDispatch}
            >
                <div className="flex flex-row h-[20vh] w-max items-center">
                    <div className="h-max w-max">
                        <Text size="40" style={{ lineHeight: '40px' }}>
                            {valid ? 'Wormhole to' : 'Wormhole not found'}
                        </Text>
                        {!valid && (
                            <Text size="20" style={{ lineHeight: '40px' }}>
                                {'Wormhole does not exist or was closed'}
                            </Text>
                        )}
                    </div>
                    {valid && (
                        <IconFolder size={40} style={{ marginLeft: '7px' }} />
                    )}
                    <Text
                        fw={700}
                        size="40"
                        style={{ lineHeight: '40px', marginLeft: 3 }}
                    >
                        {wormholeInfo?.shareName}
                    </Text>
                </div>
                {valid && (
                    <UploadPlaque
                        wormholeId={wormholeId}
                        uploadDispatch={uploadDispatch}
                    />
                )}
            </WormholeWrapper>
        </div>
    )
}
