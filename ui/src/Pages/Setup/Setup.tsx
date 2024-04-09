import { Box, Button, Divider, Input, Text } from "@mantine/core";
import { WeblensButton } from "../../components/WeblensButton";
import "../../style/setup.css";
import {
    IconArrowLeft,
    IconDatabaseImport,
    IconExclamationCircle,
    IconJumpRope,
    IconPackage,
    IconRocket,
} from "@tabler/icons-react";
import { useContext, useEffect, useMemo, useState } from "react";
import { getUsersPublic, initServer } from "../../api/ApiFetch";
import { useNavigate } from "react-router-dom";
import { UserInfoT } from "../../types/Types";
import { userContext } from "../../Context";

const UserSelect = ({
    users,
    username,
    setUsername,
    setPassword,
    owner,
}: {
    users: UserInfoT[];
    username;
    setUsername;
    setPassword;
    owner;
}) => {
    if (users.length === 0) {
        return null;
    }

    if (owner) {
        return (
            <Box className="caution-box">
                <Box className="caution-header">
                    <Text
                        className="subheader-text"
                        c="white"
                        style={{ paddingTop: 0 }}
                    >
                        This Server Already Has An Owner
                    </Text>
                    <IconExclamationCircle color="white" />
                </Box>
                <Text
                    className="body-text"
                    style={{ fontSize: "12px" }}
                    c="white"
                >
                    Log in as {owner.username} to continue
                </Text>
                <Input
                    variant="unstyled"
                    className="weblens-input"
                    type="password"
                    placeholder="Password"
                    style={{ width: "100%" }}
                    onChange={(e) => setPassword(e.target.value)}
                />
            </Box>
        );
    }

    return (
        <Box className="caution-box">
            <Box className="caution-header">
                <Text
                    className="subheader-text"
                    c="#ffffff"
                    style={{ paddingTop: 0 }}
                >
                    This Server Already Has Users
                </Text>
                <IconExclamationCircle color="white" />
            </Box>
            <Text className="body-text" c="#ffffff">
                Select a user to make owner
            </Text>
            <Box
                style={{ width: "100%", height: "max-content", flexShrink: 0 }}
            >
                {users.map((u) => {
                    return (
                        <WeblensButton
                            toggleOn={u.username === username}
                            allowRepeat={false}
                            key={u.username}
                            label={u.username}
                            style={{ padding: 2 }}
                            onClick={() => setUsername(u.username)}
                        />
                    );
                })}
            </Box>
            {username && (
                <Box style={{ width: "100%" }}>
                    <Text className="body-text" c="#ffffff">
                        Log in as {username} to continue
                    </Text>
                    <Input
                        variant="unstyled"
                        className="weblens-input"
                        type="password"
                        placeholder="Password"
                        style={{ width: "100%" }}
                        onChange={(e) => setPassword(e.target.value)}
                    />
                </Box>
            )}
        </Box>
    );
};

const Core = ({
    page,
    setPage,
    existingName,
}: {
    page: string;
    setPage: (page: string) => void;
    existingName: string;
}) => {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [serverName, setServerName] = useState(existingName);
    const [users, setUsers] = useState([]);
    const nav = useNavigate();

    useEffect(() => {
        getUsersPublic().then((r) => {
            setUsers(r.users);
        });
    }, []);

    const owner = useMemo(() => {
        const owner = users.filter((u) => u.owner)[0];
        if (owner) {
            setUsername(owner);
        }
        return owner;
    }, [users]);

    let ondeck;
    if (page === "core") {
        ondeck = "active";
    } else if (page === "landing") {
        ondeck = "next";
    }
    return (
        <Box className="setup-content-box" mod={{ "data-ondeck": ondeck }}>
            <Box style={{ width: "100%" }}>
                <WeblensButton
                    Left={<IconArrowLeft />}
                    subtle
                    onClick={() => setPage("landing")}
                />
            </Box>
            <Box style={{ display: "flex", alignItems: "center", gap: 20 }}>
                <IconPackage color="white" size={"60px"} />
                <Text className="header-text">Weblens Core</Text>
            </Box>
            {users.length === 0 && (
                <Box style={{ width: "100%" }}>
                    <Text className="body-text">Create the Owner Account</Text>
                    <Input
                        className="weblens-input"
                        variant="unstyled"
                        placeholder="Username"
                        onChange={(e) => {
                            setUsername(e.target.value);
                        }}
                    />
                    <Input
                        className="weblens-input"
                        variant="unstyled"
                        type="password"
                        placeholder="Password"
                        onChange={(e) => {
                            setPassword(e.target.value);
                        }}
                    />
                </Box>
            )}
            <UserSelect
                users={users}
                username={username}
                setUsername={setUsername}
                setPassword={setPassword}
                owner={owner}
            />

            <Divider />

            {existingName && (
                <Box className="caution-box">
                    <Box className="caution-header">
                        <Text
                            className="subheader-text"
                            c="#ffffff"
                            style={{ paddingTop: 0 }}
                        >
                            This server already has a name
                        </Text>
                        <IconExclamationCircle color="white" />
                    </Box>

                    <Input
                        className="weblens-input"
                        styles={{
                            input: {
                                backgroundColor: "#00000000",
                                color: "white",
                            },
                        }}
                        variant="unstyled"
                        disabled
                        value={serverName}
                    />
                </Box>
            )}
            {!existingName && (
                <Box>
                    <Text className="body-text">Name This Weblens Server</Text>
                    <Input
                        className="weblens-input"
                        classNames={{ input: "weblens-input" }}
                        variant="unstyled"
                        disabled={existingName !== ""}
                        value={serverName}
                        placeholder="My Radical Weblens Server"
                        onChange={(e) => {
                            setServerName(e.target.value);
                        }}
                    />
                </Box>
            )}

            <WeblensButton
                label="Start Weblens"
                Left={<IconRocket className="button-icon" />}
                disabled={
                    serverName === "" || username === "" || password === ""
                }
                doSuper
                onClick={async () => {
                    const ret = await initServer(
                        serverName,
                        "core",
                        username,
                        password,
                        "",
                        ""
                    );
                    if (ret.status !== 201) {
                        console.error(ret.statusText);
                        return false;
                    }

                    nav("/login");

                    return true;
                }}
            />
        </Box>
    );
};

const Backup = ({
    page,
    setPage,
}: {
    page: string;
    setPage: (page: string) => void;
}) => {
    const [serverName, setServerName] = useState("");
    const [coreAddress, setCoreAddress] = useState("");
    const [apiKey, setApiKey] = useState("");
    const nav = useNavigate();

    let ondeck;
    if (page === "backup") {
        ondeck = "active";
    } else if (page === "landing" || page === "core") {
        ondeck = "next";
    }
    return (
        <Box className="setup-content-box" mod={{ "data-ondeck": ondeck }}>
            <Box style={{ width: "100%" }}>
                <WeblensButton
                    Left={<IconArrowLeft />}
                    subtle
                    onClick={() => setPage("landing")}
                />
            </Box>
            <Box style={{ display: "flex", alignItems: "center", gap: 20 }}>
                <IconDatabaseImport color="white" size={"60px"} />
                <Text className="header-text">Weblens Backup</Text>
            </Box>
            <Text className="body-text">Name Local (Backup) Server</Text>
            <Input
                className="weblens-input"
                variant="unstyled"
                placeholder="Rad Backup Server"
                onChange={(e) => {
                    setServerName(e.target.value);
                }}
            />
            <Text className="body-text">Remote (Core) Weblens Address</Text>
            <Input
                className="weblens-input"
                variant="unstyled"
                placeholder="https://myremoteweblens.net/"
                onChange={(e) => {
                    setCoreAddress(e.target.value);
                }}
            />
            <Text className="body-text">API Key</Text>
            <Input
                className="weblens-input"
                variant="unstyled"
                type="password"
                placeholder="RUH8gHMH4EgQvw_n2..."
                onChange={(e) => {
                    setApiKey(e.target.value);
                }}
            />

            <WeblensButton
                label="Attach to Remote"
                Left={<IconRocket className="button-icon" />}
                disabled={
                    serverName === "" || coreAddress === "" || apiKey === ""
                }
                doSuper
                onClick={async () => {
                    const ret = await initServer(
                        serverName,
                        "backup",
                        "",
                        "",
                        coreAddress,
                        apiKey
                    );
                    if (ret.status !== 201) {
                        console.error(ret.statusText);
                        return false;
                    }

                    nav("/login");

                    return true;
                }}
            />
        </Box>
    );
};

const Landing = ({
    page,
    setPage,
}: {
    page: string;
    setPage: (page: string) => void;
}) => {
    let ondeck;
    if (page === "landing") {
        ondeck = "active";
    } else if (page === "core" || page === "backup") {
        ondeck = "prev";
    }

    return (
        <Box className="setup-content-box" mod={{ "data-ondeck": ondeck }}>
            <Text className="title-text">WEBLENS</Text>
            {/* <Text className="content-title-text">Set Up Weblens</Text> */}
            <WeblensButton
                Left={<IconPackage size={"100%"} />}
                label="Set Up Weblens Core"
                centerContent
                fontSize="40px"
                onClick={() => setPage("core")}
            />
            <Text>Or...</Text>
            <WeblensButton
                label="Set Up As Backup"
                Left={<IconDatabaseImport />}
                style={{ width: "200px" }}
                centerContent
                subtle
                onClick={() => setPage("backup")}
            />
        </Box>
    );
};

const Setup = () => {
    const { serverInfo } = useContext(userContext);
    const [page, setPage] = useState("landing");

    if (!serverInfo) {
        return null;
    }

    console.log(serverInfo.name);

    return (
        <Box className="setup-container">
            {/* <Box className="background-strip"></Box> */}
            <Box className="setup-content-pane" mod={{ "data-active": true }}>
                <Landing page={page} setPage={setPage} />
                <Core
                    page={page}
                    setPage={setPage}
                    existingName={serverInfo.name}
                />
                <Backup page={page} setPage={setPage} />
            </Box>
            {/* <ScatteredPhotos /> */}
        </Box>
    );
};

export default Setup;
