import { Box, Text } from "@mantine/core";
import "./style/history.css";
import { getFileHistory, getSnapshots } from "../../api/FileBrowserApi";
import { useCallback, useContext, useEffect, useState } from "react";
import { userContext } from "../../Context";
import { FBDispatchT, UserContextT } from "../../types/Types";
import { IconHistory } from "@tabler/icons-react";
import { useClick } from "../../components/hooks";

function FileHistoryMenu({ fileId }: { fileId: string }) {
    const { authHeader }: UserContextT = useContext(userContext);
    const [history, setHistory] = useState(null);
    useEffect(() => {
        getFileHistory(fileId, authHeader).then((r) => setHistory(r));
    }, []);
    return (
        <Box className="file-history-menu">
            {history.events.map((e) => {
                return <Text>{e.action}</Text>;
            })}
        </Box>
    );
}

export function SnapshotMenu({ dispatch }: { dispatch: FBDispatchT }) {
    const { authHeader }: UserContextT = useContext(userContext);
    const [snapshots, setSnapshots] = useState(null);
    const [boxRef, setBoxRef] = useState(null);
    const [open, setOpen] = useState(false);
    const closeOnClick = useCallback((e) => setOpen(false), []);
    useClick(closeOnClick, boxRef);

    useEffect(() => {
        if (open && !snapshots) {
            getSnapshots(authHeader).then((r) => {
                const snaps = r.snapshots;
                console.log(snaps);
                snaps.sort((a, b) => {
                    return Date.parse(b.Timestamp) - Date.parse(a.Timestamp);
                });
                console.log(snaps);
                setSnapshots(snaps);
            });
        }
    }, [open]);

    return (
        <Box
            id="snapshot-menu"
            ref={setBoxRef}
            className={`snapshot-menu snapshot-menu-${
                open ? "open" : "closed"
            }`}
            onClick={() => setOpen(true)}
        >
            <Box
                style={{
                    display: "flex",
                    padding: 5,
                    width: 42,
                    height: 42,
                    flexShrink: 0,
                    marginRight: 5,
                }}
                onClick={() => {
                    setOpen(true);
                }}
            >
                <IconHistory className="button-icon" />
            </Box>
            {open && <Text style={{ textWrap: "nowrap" }}>Rewind Time</Text>}
            {snapshots && open && (
                <Box className="snapshot-menu-dropdown">
                    <Box
                        key={"now"}
                        className="snapshot-row"
                        onClick={() =>
                            dispatch({ type: "set_past_time", past: null })
                        }
                    >
                        <Text style={{ width: "100%" }}>Now</Text>
                    </Box>
                    {snapshots.map((s) => {
                        const d = new Date(Date.parse(s.Timestamp));
                        return (
                            <Box
                                key={d.toString()}
                                className="snapshot-row"
                                onClick={() =>
                                    dispatch({ type: "set_past_time", past: d })
                                }
                            >
                                <Text>
                                    {d.toDateString()} {d.toLocaleTimeString()}
                                </Text>
                            </Box>
                        );
                    })}
                </Box>
            )}
        </Box>
    );
}

export default FileHistoryMenu;
