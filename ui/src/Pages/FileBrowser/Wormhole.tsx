import { useParams } from "react-router-dom";
import { ColumnBox, Dropspot, RowBox } from "./FileBrowserStyles";
import { useCallback, useContext, useEffect, useState } from "react";
import { GetWormholeInfo } from "../../api/FileBrowserApi";
import { UserContext } from "../../Context";
import { shareData, UserContextT } from "../../types/Types";
import { Box, FileButton, Space, Text } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import UploadStatus, { useUploadStatus } from "../../components/UploadStatus";
import { IconFolder, IconUpload } from "@tabler/icons-react";
import { HandleDrop, HandleUploadButton } from "./FileBrowserLogic";

import "./style/fileBrowserStyle.css";

const UploadPlaque = ({
    wormholeId,
    uploadDispatch,
}: {
    wormholeId: string;
    uploadDispatch;
}) => {
    return (
        <ColumnBox style={{ height: "45vh" }}>
            <FileButton
                onChange={(files) => {
                    HandleUploadButton(
                        files,
                        wormholeId,
                        true,
                        wormholeId,
                        { Authorization: "" },
                        uploadDispatch,
                        () => {}
                    );
                }}
                accept="file"
                multiple
            >
                {(props) => {
                    return (
                        <ColumnBox
                            style={{
                                backgroundColor: "#111111",
                                height: "20vh",
                                width: "20vw",
                                padding: 10,
                                borderRadius: 4,
                                justifyContent: "center",
                            }}
                        >
                            <ColumnBox
                                onClick={() => {
                                    props.onClick();
                                }}
                                style={{
                                    cursor: "pointer",
                                    height: "max-content",
                                    width: "max-content",
                                }}
                            >
                                <IconUpload
                                    size={100}
                                    style={{ padding: "10px" }}
                                />
                                <Text size="20px" fw={600}>
                                    Upload
                                </Text>
                                <Space h={4}></Space>
                                <Text size="12px">Click or Drop</Text>
                            </ColumnBox>
                        </ColumnBox>
                    );
                }}
            </FileButton>
        </ColumnBox>
    );
};

const WormholeWrapper = ({
    wormholeId,
    wormholeName,
    fileId,
    validWormhole,
    uploadDispatch,
    children,
}: {
    wormholeId: string;
    wormholeName: string;
    fileId: string;
    validWormhole: boolean;
    uploadDispatch;
    children;
}) => {
    const { authHeader }: UserContextT = useContext(UserContext);
    const [dragging, setDragging] = useState(0);
    const [dropSpotRef, setDropSpotRef] = useState(null);
    const handleDrag = useCallback(
        (e) => {
            e.preventDefault();
            if (e.type === "dragenter" || e.type === "dragover") {
                if (!dragging) {
                    setDragging(2);
                }
            } else if (dragging) {
                setDragging(0);
            }
        },
        [dragging]
    );

    return (
        <Box className="wormhole-wrapper">
            <Box
                ref={setDropSpotRef}
                style={{ position: "relative", width: "98%", height: "98%" }}
                //                    See DirViewWrapper \/
                onMouseMove={(e) => {
                    if (dragging) {
                        setTimeout(() => setDragging(0), 10);
                    }
                }}
            >
                <Dropspot
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
                    dropspotTitle={wormholeName}
                    dragging={dragging}
                    dropAllowed={validWormhole}
                    handleDrag={handleDrag}
                    wrapperRef={dropSpotRef}
                />
                <ColumnBox
                    style={{ justifyContent: "center" }}
                    onDragOver={handleDrag}
                >
                    {children}
                </ColumnBox>
            </Box>
        </Box>
    );
};

export default function Wormhole() {
    const wormholeId = useParams()["*"];
    const { authHeader }: UserContextT = useContext(UserContext);
    const [wormholeInfo, setWormholeInfo]: [
        wormholeInfo: shareData,
        setWormholeInfo: any
    ] = useState(null);
    const { uploadState, uploadDispatch } = useUploadStatus();

    useEffect(() => {
        if (wormholeId !== "") {
            GetWormholeInfo(wormholeId, authHeader)
                .then((v) => {
                    if (v.status !== 200) {
                        return Promise.reject(v.statusText);
                    }
                    return v.json();
                })
                .then((v) => {
                    setWormholeInfo(v);
                })
                .catch((r) => {
                    notifications.show({
                        title: "Failed to get wormhole info",
                        message: String(r),
                        color: "red",
                    });
                });
        }
    }, [wormholeId, authHeader]);
    const valid = Boolean(wormholeInfo);

    return (
        <Box>
            <UploadStatus
                uploadState={uploadState}
                uploadDispatch={uploadDispatch}
            />
            <WormholeWrapper
                wormholeId={wormholeId}
                wormholeName={wormholeInfo?.ShareName}
                fileId={wormholeInfo?.fileId}
                validWormhole={valid}
                uploadDispatch={uploadDispatch}
            >
                <RowBox style={{ height: "20vh", width: "max-content" }}>
                    <ColumnBox
                        style={{ height: "max-content", width: "max-content" }}
                    >
                        <Text size="40" style={{ lineHeight: "40px" }}>
                            {valid ? "Wormhole to" : "Wormhole not found"}
                        </Text>
                        {!valid && (
                            <Text size="20" style={{ lineHeight: "40px" }}>
                                {"Wormhole does not exist or was closed"}
                            </Text>
                        )}
                    </ColumnBox>
                    {valid && (
                        <IconFolder size={40} style={{ marginLeft: "7px" }} />
                    )}
                    <Text
                        fw={700}
                        size="40"
                        style={{ lineHeight: "40px", marginLeft: 3 }}
                    >
                        {wormholeInfo?.ShareName}
                    </Text>
                </RowBox>
                {valid && (
                    <UploadPlaque
                        wormholeId={wormholeId}
                        uploadDispatch={uploadDispatch}
                    />
                )}
            </WormholeWrapper>
        </Box>
    );
}
