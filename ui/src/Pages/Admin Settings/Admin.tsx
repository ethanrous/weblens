import {
    Box,
    Checkbox,
    Input,
    ScrollArea,
    Space,
    Text,
    TextInput,
} from "@mantine/core";
import { useCallback, useContext, useEffect, useMemo, useState } from "react";
import { UserContext } from "../../Context";
import {
    clearCache,
    adminCreateUser,
    newApiKey,
    getApiKeys,
    doBackup,
    deleteApiKey,
    getRemotes,
    deleteRemote,
} from "../../api/ApiFetch";
import {
    ActivateUser,
    DeleteUser,
    GetUsersInfo,
    SetUserAdmin,
} from "../../api/UserApi";
import { notifications } from "@mantine/notifications";
import { ColumnBox, RowBox } from "../FileBrowser/FileBrowserStyles";
import {
    IconClipboard,
    IconRefresh,
    IconTrash,
    IconX,
} from "@tabler/icons-react";
import { RecurScanFolder } from "../../api/FileBrowserApi";
import {
    AuthHeaderT as AuthHeaderT,
    UserContextT,
    UserInfoT as UserInfoT,
} from "../../types/Types";
import { WeblensButton } from "../../components/WeblensButton";
import { useKeyDown } from "../../components/hooks";

function CreateUserBox({
    setAllUsersInfo,
    authHeader,
}: {
    setAllUsersInfo;
    authHeader: AuthHeaderT;
}) {
    const [userInput, setUserInput] = useState("");
    const [passInput, setPassInput] = useState("");
    const [makeAdmin, setMakeAdmin] = useState(false);
    return (
        <ColumnBox
            style={{
                backgroundColor: "#222222",
                padding: 20,
                height: "max-content",
                width: 450,
            }}
        >
            <Input
                className="weblens-input-wrapper"
                variant="unstyled"
                value={userInput}
                placeholder="Username"
                style={{ margin: "8px" }}
                onChange={(e) => setUserInput(e.target.value)}
            />
            <Input
                className="weblens-input-wrapper"
                variant="unstyled"
                value={passInput}
                placeholder="Password"
                style={{ margin: "8px" }}
                onChange={(e) => setPassInput(e.target.value)}
            />
            <Checkbox
                label="Admin"
                onChange={(e) => {
                    setMakeAdmin(e.target.checked);
                }}
            />
            <Space h={20} />
            <WeblensButton
                label="Create User"
                height={40}
                width={185}
                disabled={userInput === "" || passInput === ""}
                onClick={async () => {
                    await adminCreateUser(
                        userInput,
                        passInput,
                        makeAdmin,
                        authHeader
                    ).then(() => {
                        GetUsersInfo(setAllUsersInfo, authHeader);
                        setUserInput("");
                        setPassInput("");
                    });
                    return true;
                }}
            />
        </ColumnBox>
    );
}

const UserRow = ({
    rowUser,
    accessor,
    setAllUsersInfo,
    authHeader,
}: {
    rowUser: UserInfoT;
    accessor: UserInfoT;
    setAllUsersInfo;
    authHeader: AuthHeaderT;
}) => {
    return (
        <RowBox
            key={rowUser.username}
            style={{
                width: "95%",
                height: "65px",
                alignItems: "center",
                outline: "1px solid #888888",
                borderRadius: "2px",
                padding: 15,
                margin: 5,
            }}
        >
            <IconRefresh
                style={{ cursor: "pointer" }}
                onClick={() => RecurScanFolder(rowUser.homeId, authHeader)}
            />
            <ColumnBox
                style={{
                    justifyContent: "center",
                    width: "max-content",
                    paddingLeft: "10px",
                }}
            >
                <Text c={"white"} fw={600} style={{ width: "max-content" }}>
                    {rowUser.username}
                </Text>
                {rowUser.admin && !rowUser.owner && !accessor.owner && (
                    <Text c={"#aaaaaa"}>Admin</Text>
                )}
                {rowUser.owner && <Text c={"#aaaaaa"}>Owner</Text>}
                {!rowUser.admin && accessor.owner && (
                    <WeblensButton
                        label="Make Admin"
                        width={"max-content"}
                        height={20}
                        style={{ padding: 4 }}
                        onClick={() => {
                            SetUserAdmin(
                                rowUser.username,
                                true,
                                authHeader
                            ).then(() =>
                                GetUsersInfo(setAllUsersInfo, authHeader)
                            );
                        }}
                    />
                )}
                {!rowUser.owner && rowUser.admin && accessor.owner && (
                    <WeblensButton
                        label="Remove Admin"
                        width={"max-content"}
                        height={20}
                        style={{ padding: 4 }}
                        onClick={() => {
                            SetUserAdmin(
                                rowUser.username,
                                false,
                                authHeader
                            ).then(() =>
                                GetUsersInfo(setAllUsersInfo, authHeader)
                            );
                        }}
                    />
                )}
            </ColumnBox>
            <Space style={{ display: "flex", flexGrow: 1 }} />
            {rowUser.activated === false && (
                <WeblensButton
                    label="Activate"
                    height={20}
                    onClick={() => {
                        ActivateUser(rowUser.username, authHeader).then(() =>
                            GetUsersInfo(setAllUsersInfo, authHeader)
                        );
                    }}
                />
            )}
            <Space style={{ display: "flex", flexGrow: 1 }} />

            <WeblensButton
                label="Delete"
                height={20}
                width={"max-content"}
                danger
                centerContent
                disabled={rowUser.admin}
                onClick={() => {
                    DeleteUser(rowUser.username, authHeader).then(() =>
                        GetUsersInfo(setAllUsersInfo, authHeader)
                    );
                }}
            />
        </RowBox>
    );
};

function UsersBox({
    thisUserInfo,
    allUsersInfo,
    setAllUsersInfo,
    authHeader,
}: {
    thisUserInfo: UserInfoT;
    allUsersInfo: UserInfoT[];
    setAllUsersInfo;
    authHeader: AuthHeaderT;
}) {
    const usersList = useMemo(() => {
        if (!allUsersInfo) {
            return null;
        }
        allUsersInfo.sort((a, b) => {
            return a.username.localeCompare(b.username);
        });

        const usersList = allUsersInfo.map((val) => (
            <UserRow
                key={val.username}
                rowUser={val}
                accessor={thisUserInfo}
                setAllUsersInfo={setAllUsersInfo}
                authHeader={authHeader}
            />
        ));
        return usersList;
    }, [allUsersInfo, authHeader, setAllUsersInfo]);

    return (
        <ColumnBox
            style={{
                padding: "10px",
                backgroundColor: "#222222",
                height: "max-content",
                width: "450px",
            }}
        >
            <Text size={"20px"} fw={800} c={"white"}>
                Users
            </Text>
            <Space h={"10px"} />
            <ScrollArea w={"100%"} type="never" maw={450} mah={400}>
                {usersList}
            </ScrollArea>
        </ColumnBox>
    );
}

export function ApiKeys({ authHeader }) {
    const { serverInfo }: UserContextT = useContext(UserContext);
    const [keys, setKeys] = useState([]);

    const getKeys = useCallback(() => {
        getApiKeys(authHeader).then((r) => {
            setKeys(r.keys);
        });
    }, [setKeys, authHeader]);

    useEffect(() => {
        getKeys();
    }, []);

    const [remotes, setRemotes] = useState([]);
    useEffect(() => {
        getRemotes(authHeader).then((r) => {
            if (r >= 400) {
                return;
            }
            setRemotes(r.remotes);
        });
    }, []);

    return (
        <Box
            style={{
                width: 450,
                padding: 5,
                backgroundColor: "#222222",
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
            }}
        >
            {keys.length !== 0 && (
                <Box
                    style={{
                        backgroundColor: "#333333",
                        display: "flex",
                        flexDirection: "column",
                        alignItems: "center",
                        padding: 5,
                        paddingLeft: 10,
                        borderRadius: 4,
                        margin: 20,
                        maxWidth: "100%",
                    }}
                >
                    {keys.map((k) => {
                        return (
                            <Box
                                key={k.Key.slice(0, 6)}
                                style={{
                                    display: "flex",
                                    flexDirection: "row",
                                    alignItems: "center",
                                    maxWidth: "100%",
                                }}
                            >
                                <Box
                                    style={{
                                        display: "flex",
                                        flexDirection: "column",
                                        alignItems: "flex-start",
                                        flexGrow: 1,
                                        width: "50%",
                                    }}
                                >
                                    <Text
                                        truncate="end"
                                        style={{
                                            textWrap: "nowrap",
                                            width: "100%",
                                        }}
                                    >
                                        {k.Key}
                                    </Text>
                                    {k.RemoteUsing !== "" && (
                                        <Text>{k.RemoteUsing}</Text>
                                    )}
                                </Box>
                                <IconClipboard
                                    size={"40px"}
                                    style={{
                                        flexShrink: 0,
                                        margin: 4,
                                        backgroundColor: "#222222",
                                        borderRadius: 2,
                                        padding: 4,
                                        cursor: "pointer",
                                    }}
                                    onClick={() => {
                                        if (!window.isSecureContext) {
                                            notifications.show({
                                                message:
                                                    "Browser context is not secure, are you not using HTTPS?",
                                                color: "red",
                                            });
                                            return;
                                        }
                                        navigator.clipboard.writeText(k.Key);
                                    }}
                                />
                                <IconTrash
                                    size={"40px"}
                                    style={{
                                        flexShrink: 0,
                                        margin: 4,
                                        backgroundColor: "#222222",
                                        borderRadius: 2,
                                        padding: 4,
                                        cursor: "pointer",
                                    }}
                                    onClick={() => {
                                        deleteApiKey(k.Key, authHeader).then(
                                            () => {
                                                setKeys((ks) => {
                                                    ks = ks.filter(
                                                        (i) => i !== k
                                                    );
                                                    return [...ks];
                                                });
                                            }
                                        );
                                    }}
                                />
                            </Box>
                        );
                    })}
                </Box>
            )}
            <WeblensButton
                width={200}
                height={40}
                label="New Api Key"
                onClick={() => {
                    newApiKey(authHeader).then((k) =>
                        setKeys((ks) => {
                            ks.push(k.key);
                            return [...ks];
                        })
                    );
                }}
            />
            <Box
                style={{
                    display: "flex",
                    flexDirection: "column",
                    alignItems: "center",
                    padding: 5,
                    paddingLeft: 10,
                    borderRadius: 4,
                    margin: 20,
                    width: "100%",
                    gap: 10,
                }}
            >
                {remotes.map((r) => {
                    if (r.id === serverInfo.id) {
                        return null;
                    }
                    return (
                        <Box
                            key={r.name}
                            style={{
                                backgroundColor: "#333333",
                                display: "flex",
                                flexDirection: "row",
                                alignItems: "center",
                                width: "100%",
                                borderRadius: 4,
                                paddingLeft: 20,
                                justifyContent: "space-between",
                            }}
                        >
                            <Text>{r.name}</Text>
                            <IconTrash
                                size={"40px"}
                                style={{
                                    flexShrink: 0,
                                    margin: 4,
                                    backgroundColor: "#222222",
                                    borderRadius: 2,
                                    padding: 4,
                                    cursor: "pointer",
                                }}
                                onClick={() => {
                                    deleteRemote(r.id, authHeader).then(() => {
                                        getRemotes(authHeader).then((r) => {
                                            if (r >= 400) {
                                                return;
                                            }
                                            setRemotes(r.remotes);
                                        });
                                    });
                                }}
                            />
                        </Box>
                    );
                })}
            </Box>
        </Box>
    );
}

export function Admin({ close }) {
    const { authHeader, usr, serverInfo }: UserContextT =
        useContext(UserContext);
    const [allUsersInfo, setAllUsersInfo] = useState(null);
    useKeyDown("Escape", close);

    useEffect(() => {
        if (authHeader.Authorization !== "") {
            GetUsersInfo(setAllUsersInfo, authHeader);
        }
    }, [authHeader]);

    if (usr.isLoggedIn === undefined) {
        return null;
    }

    return (
        <Box className="settings-menu" onClick={(e) => e.stopPropagation()}>
            <RowBox style={{ position: "relative", height: "max-content" }}>
                <IconX
                    style={{
                        margin: 10,
                        cursor: "pointer",
                    }}
                    onClick={close}
                />
            </RowBox>
            <ColumnBox
                style={{
                    // marginTop: "50px",
                    // paddingBottom: "50px",
                    height: "max-content",
                    maxHeight: "80vh",
                    padding: 30,
                    justifyContent: "center",
                    alignItems: "center",
                    borderRadius: 8,
                }}
            >
                <UsersBox
                    thisUserInfo={usr}
                    allUsersInfo={allUsersInfo}
                    setAllUsersInfo={setAllUsersInfo}
                    authHeader={authHeader}
                />
                <Space h={10} />
                <CreateUserBox
                    setAllUsersInfo={setAllUsersInfo}
                    authHeader={authHeader}
                />

                <Space h={25} />
                <ApiKeys authHeader={authHeader} />
                <RowBox
                    style={{
                        height: "max-content",
                        justifyContent: "center",
                    }}
                >
                    <WeblensButton
                        label="Clear Cache"
                        height={40}
                        width={"200px"}
                        danger
                        onClick={() => {
                            clearCache(authHeader).then(() =>
                                notifications.show({
                                    message: "Cache cleared",
                                })
                            );
                        }}
                    />
                    <WeblensButton
                        label="Backup now"
                        height={40}
                        width={"200px"}
                        disabled={serverInfo.role === "core"}
                        postScript={
                            serverInfo.role === "core"
                                ? "Core servers do not support backup"
                                : ""
                        }
                        onClick={async () => {
                            const res = await doBackup(authHeader);
                            if (res >= 300) {
                                return false;
                            }
                            return true;
                        }}
                    />
                </RowBox>
            </ColumnBox>
        </Box>
    );
}

export default Admin;
