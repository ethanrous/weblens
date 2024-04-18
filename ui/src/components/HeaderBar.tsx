import { Ref, memo, useCallback, useContext, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import WeblensLoader from "./Loading";

import { userContext } from "../Context";
import {
    ActionIcon,
    Box,
    Input,
    Loader,
    Menu,
    Space,
    Text,
    Tooltip,
} from "@mantine/core";
import {
    IconFolder,
    IconInfoCircle,
    IconLogin,
    IconLogout,
    IconPhoto,
    IconServerCog,
    IconTools,
    IconUser,
    IconX,
} from "@tabler/icons-react";
import { ColumnBox, RowBox } from "../Pages/FileBrowser/FilebrowserStyles";
import { IconSettings } from "@tabler/icons-react";
import { WeblensButton } from "./WeblensButton";
import { useKeyDown } from "./hooks";
import { UpdatePassword } from "../api/UserApi";
import { AuthHeaderT, UserContextT, UserInfoT } from "../types/Types";
import { IconArrowLeft } from "@tabler/icons-react";
import Admin from "../Pages/Admin Settings/Admin";

type HeaderBarProps = {
    dispatch: React.Dispatch<any>;
    page: string;
    loading: string[];
};

const SettingsMenu = ({
    open,
    setClosed,
    usr,
    authHeader,
}: {
    open: boolean;
    setClosed: () => void;
    usr: UserInfoT;
    authHeader: AuthHeaderT;
}) => {
    const [oldP, setOldP] = useState("");
    const [newP, setNewP] = useState("");
    useKeyDown("Escape", () => {
        setNewP("");
        setOldP("");
        setClosed();
    });

    const updateFunc = useCallback(async () => {
        const res =
            (await UpdatePassword(usr.username, oldP, newP, authHeader))
                .status === 200;
        if (res) {
            setOldP("");
            setNewP("");
        }
        setTimeout(() => setClosed(), 2000);
        return res;
    }, [usr.username, String(oldP), String(newP), authHeader]);

    if (!open) {
        return null;
    }
    return (
        <Box className="settings-menu-container" onClick={() => setClosed()}>
            <Box className="settings-menu" onClick={(e) => e.stopPropagation()}>
                <Box
                    style={{
                        width: "max-content",
                        display: "flex",
                        flexDirection: "column",
                        justifyContent: "center",
                        gap: 10,
                        padding: 100,
                    }}
                >
                    <Text
                        size="20px"
                        fw={600}
                        style={{ padding: 7, width: "90%", textWrap: "nowrap" }}
                    >
                        Change Password
                    </Text>
                    <Input
                        value={oldP}
                        variant="unstyled"
                        type="password"
                        placeholder="Old Password"
                        className="weblens-input-wrapper"
                        onChange={(v) => setOldP(v.target.value)}
                    />
                    <Input
                        value={newP}
                        variant="unstyled"
                        type="password"
                        placeholder="New Password"
                        className="weblens-input-wrapper"
                        onChange={(v) => setNewP(v.target.value)}
                    />
                    <WeblensButton
                        label="Update Password"
                        showSuccess
                        disabled={oldP == "" || newP == "" || oldP === newP}
                        width={"100%"}
                        onClick={updateFunc}
                    />
                </Box>
                <IconX
                    style={{
                        position: "absolute",
                        top: 0,
                        left: 0,
                        cursor: "pointer",
                        margin: 10,
                    }}
                    onClick={setClosed}
                />
            </Box>
        </Box>
    );
};

const HeaderBar = memo(
    ({ dispatch, page, loading }: HeaderBarProps) => {
        const { usr, authHeader, clear, serverInfo }: UserContextT =
            useContext(userContext);
        const nav = useNavigate();
        const [userMenu, setUserMenu] = useState(false);
        const [settings, setSettings] = useState(false);
        const [admin, setAdmin] = useState(false);
        const spacing = "8px";

        return (
            <Box style={{ zIndex: 3, height: "max-content", width: "100vw" }}>
                <SettingsMenu
                    open={settings}
                    usr={usr}
                    setClosed={() => {
                        setSettings(false);
                        dispatch({ type: "set_block_focus", block: false });
                    }}
                    authHeader={authHeader}
                />
                {admin && (
                    <Box
                        className="settings-menu-container"
                        onClick={() => setAdmin(false)}
                    >
                        <Admin close={() => setAdmin(false)} />
                    </Box>
                )}
                <Box
                    style={{
                        position: "absolute",
                        float: "right",
                        right: "40px",
                        bottom: "30px",
                        zIndex: 10,
                    }}
                >
                    <WeblensLoader loading={loading} />
                </Box>
                <RowBox
                    style={{
                        height: 56,
                        paddingTop: 8,
                        paddingBottom: 8,
                        borderBottom: "2px solid #222222",
                    }}
                >
                    <Box style={{ paddingLeft: "10px" }} />
                    {page === "gallery" && (
                        <WeblensButton
                            label="Files"
                            centerContent
                            subtle
                            Left={<IconArrowLeft className="button-icon" />}
                            Right={<IconFolder className="button-icon" />}
                            onClick={() => nav("/files/")}
                            width={"max-content"}
                            height={"40px"}
                        />
                    )}
                    {page === "files" && usr.isLoggedIn !== undefined && (
                        <WeblensButton
                            label="Gallery"
                            centerContent
                            subtle
                            Left={<IconArrowLeft className="button-icon" />}
                            Right={<IconPhoto className="button-icon" />}
                            onClick={() => nav("/")}
                            width={"max-content"}
                            height={"40px"}
                        />
                    )}
                    <Space style={{ flexGrow: 2 }} />
                    <Box
                        style={{
                            display: "flex",
                            flexDirection: "column",
                            alignItems: "flex-end",
                            height: "max-content",
                            width: "max-content",
                            paddingRight: 10,
                            color: "#575757",
                        }}
                    >
                        <Text size="12px">{serverInfo?.name}</Text>
                        <Text size="12px">({serverInfo?.role})</Text>
                    </Box>
                    <Tooltip
                        color="#222222"
                        label={
                            <ColumnBox>
                                <Text
                                    size="20px"
                                    fw={600}
                                    style={{ paddingBottom: "8px" }}
                                >
                                    {import.meta.env.VITE_APP_BUILD_TAG ||
                                        "local"}
                                </Text>
                                <Text size="12px">
                                    Click to report an issue
                                </Text>
                            </ColumnBox>
                        }
                    >
                        <IconInfoCircle
                            size={28}
                            style={{
                                cursor: "pointer",
                                display: "flex",
                                flexShrink: 0,
                            }}
                            onClick={() =>
                                window.open(
                                    `https://github.com/ethanrous/weblens/issues/new?title=Issue%20with%20${
                                        import.meta.env.VITE_APP_BUILD_TAG
                                            ? import.meta.env.VITE_APP_BUILD_TAG
                                            : "local"
                                    }`,
                                    "_blank"
                                )
                            }
                        />
                    </Tooltip>
                    {usr?.admin && (
                        <Tooltip label={"Admin Settings"} color="#222222">
                            <IconServerCog
                                size={32}
                                onClick={() => {
                                    dispatch({
                                        type: "set_block_focus",
                                        block: true,
                                    });
                                    setAdmin(true);
                                }}
                                style={{ margin: spacing, cursor: "pointer" }}
                            />
                        </Tooltip>
                    )}
                    <Menu opened={userMenu} onClose={() => setUserMenu(false)}>
                        <Menu.Target>
                            <Tooltip label={"User Settings"} color="#222222">
                                <IconUser
                                    size={32}
                                    onClick={() => setUserMenu(true)}
                                    style={{ cursor: "pointer" }}
                                />
                            </Tooltip>
                        </Menu.Target>

                        <Menu.Dropdown>
                            <Menu.Label>
                                {usr.username ? usr.username : "Not logged in"}
                            </Menu.Label>
                            <Box
                                className="menu-item"
                                mod={{
                                    "data-disabled": (
                                        usr.username === ""
                                    ).toString(),
                                }}
                                onClick={() => {
                                    setUserMenu(false);
                                    setSettings(true);
                                    dispatch({
                                        type: "set_block_focus",
                                        block: true,
                                    });
                                }}
                            >
                                <IconSettings
                                    color="white"
                                    size={20}
                                    style={{ flexShrink: 0 }}
                                />
                                <Text className="menu-item-text">Settings</Text>
                            </Box>
                            {usr.username === "" && (
                                <Box
                                    className="menu-item"
                                    onClick={() => {
                                        nav("/login", {
                                            state: { doLogin: false },
                                        });
                                    }}
                                >
                                    <IconLogin
                                        color="white"
                                        size={20}
                                        style={{ flexShrink: 0 }}
                                    />
                                    <Text className="menu-item-text">
                                        Login
                                    </Text>
                                </Box>
                            )}
                            {usr.username !== "" && (
                                <Box
                                    className="menu-item"
                                    onClick={() => {
                                        clear(); // Clears cred cookies from browser
                                        nav("/login", {
                                            state: { doLogin: false },
                                        });
                                    }}
                                >
                                    <IconLogout
                                        color="white"
                                        size={20}
                                        style={{ flexShrink: 0 }}
                                    />
                                    <Text className="menu-item-text">
                                        Logout
                                    </Text>
                                </Box>
                            )}
                        </Menu.Dropdown>
                    </Menu>

                    <Box style={{ paddingRight: "10px" }} />
                </RowBox>
            </Box>
        );
    },
    (prev, next) => {
        if (prev.loading !== next.loading) {
            return false;
        } else if (prev.page !== next.page) {
            return false;
        }
        return true;
    }
);

export default HeaderBar;
