import { useCallback, useContext, useState } from "react";
import useWebSocket from "react-use-websocket";
import { API_WS_ENDPOINT } from "./ApiEndpoint";
import { userContext } from "../Context";
import { UserContextT } from "../types/Types";

export default function useWeblensSocket() {
    const { usr, authHeader }: UserContextT = useContext(userContext);
    const [givenUp, setGivenUp] = useState(false)
    const { sendMessage, lastMessage, readyState } = useWebSocket(API_WS_ENDPOINT, {
        onOpen: () => {
            setGivenUp(false)
            sendMessage(JSON.stringify({auth: authHeader.Authorization}))
        },
        reconnectAttempts: 5,
        reconnectInterval: (last) => {
            return ((last + 1) ^ 2) * 1000;
        },
        shouldReconnect: () => usr.username !== "",
        onReconnectStop: () => {
            setGivenUp(true)
        },
    });
    const wsSend = useCallback(
        (action: string, content: any) => {
            const msg = {
                action: action,
                content: JSON.stringify(content),
            };
            console.log("WSSend", msg);
            sendMessage(JSON.stringify(msg));
        },
        [sendMessage],
    );

    return {
        wsSend,
        lastMessage,
        readyState: givenUp ? -1 : readyState,
    };
}

export function dispatchSync(
    folderIds: string | string[],
    wsSend: (action: string, content: any) => void,
    recursive: boolean,
    full: boolean,
) {
    folderIds = folderIds instanceof Array ? folderIds : [folderIds];
    for (const folderId of folderIds) {
        wsSend("scan_directory", {
            folderId: folderId,
            recursive: recursive,
            full: full,
        });
    }
}
