import { useParams } from "react-router-dom";
import { ColumnBox, RowBox, WormholeWrapper } from "./FilebrowserStyles";
import { useContext, useEffect, useState } from "react";
import { GetWormholeInfo } from "../../api/FileBrowserApi";
import { userContext } from "../../Context";
import { fileData } from "../../types/Types";
import { Box, FileButton, Space, Text } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import UploadStatus, { useUploadStatus } from "../../components/UploadStatus";
import { IconFolder, IconUpload } from "@tabler/icons-react";
import { HandleUploadButton } from "./FileBrowserLogic";

const UploadPlaque = ({ wormholeId, uploadDispatch }: { wormholeId: string, uploadDispatch }) => {
    return (
        <ColumnBox style={{ height: '45vh' }}>
            <FileButton onChange={(files) => { HandleUploadButton(files, wormholeId, true, wormholeId, {}, uploadDispatch, () => { }) }} accept="file" multiple>
                {(props) => {
                    return (
                        <ColumnBox style={{ backgroundColor: '#111111', height: '20vh', width: '20vw', padding: 10, borderRadius: 4, justifyContent: 'center' }}>
                            <ColumnBox onClick={() => { props.onClick() }} style={{ cursor: 'pointer', height: 'max-content', width: 'max-content' }}>
                                <IconUpload size={100} style={{ padding: "10px" }} />
                                <Text size='20px' fw={600}>
                                    Upload
                                </Text>
                                <Space h={4}></Space>
                                <Text size='12px'>Click or Drop</Text>
                            </ColumnBox>
                        </ColumnBox>
                    )
                }}
            </FileButton>
        </ColumnBox>
    )
}

export default function Wormhole() {
    const wormholeId = useParams()["*"]
    const { authHeader } = useContext(userContext)
    const [wormholeInfo, setWormholeInfo]: [wormholeInfo: fileData, setWormholeInfo: any] = useState(null)

    const { uploadState, uploadDispatch } = useUploadStatus()
    useEffect(() => {
        if (wormholeId !== "" && authHeader.Authorization !== "") {
            GetWormholeInfo(wormholeId, authHeader)
                .then(v => { if (v.status !== 200) { return Promise.reject(v.statusText) }; return v.json() })
                .then(v => { setWormholeInfo(v.shareInfo) })
                .catch(r => { notifications.show({ title: "Failed to get wormhole info", message: String(r), color: "red" }) })
        }
    }, [wormholeId, authHeader])
    // if (!wormholeInfo) {
    //     return null
    // }

    return (
        <Box>
            <UploadStatus uploadState={uploadState} uploadDispatch={uploadDispatch} />
            <WormholeWrapper wormholeId={wormholeId} wormholeName={wormholeInfo?.filename} validWormhole={Boolean(wormholeInfo)} uploadDispatch={uploadDispatch}>
                <RowBox style={{ height: '20vh', width: 'max-content' }}>
                    <Text size="40" style={{ lineHeight: "40px" }}>
                        {wormholeInfo?.filename ? "Wormhole to" : "Wormhole not found"}
                    </Text>
                    {wormholeInfo?.filename && (
                        <IconFolder size={40} style={{ marginLeft: '7px' }} />
                    )}
                    <Text fw={700} size="40" style={{ lineHeight: "40px", marginLeft: 3 }}>
                        {wormholeInfo?.filename}
                    </Text>
                </RowBox>
                <UploadPlaque wormholeId={wormholeId} uploadDispatch={uploadDispatch} />
            </WormholeWrapper>
        </Box>
    )
}