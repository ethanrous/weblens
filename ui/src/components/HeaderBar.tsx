import { Ref, memo, useCallback, useContext, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import WeblensLoader from "./Loading";

import { UserContext } from "../Context";
import { Box, Input, Menu, Space, Text, Tooltip } from "@mantine/core";
import {
    IconAlbum,
    IconFolder,
    IconInfoCircle,
    IconLibraryPhoto,
    IconLogin,
    IconLogout,
    IconPhoto,
    IconServerCog,
    IconUser,
    IconX,
} from "@tabler/icons-react";
import { RowBox } from "../Pages/FileBrowser/FileBrowserStyles";
import { IconSettings } from "@tabler/icons-react";
import { WeblensButton } from "./WeblensButton";
import { useKeyDown, useResize } from "./hooks";
import { UpdatePassword } from "../api/UserApi";
import { AuthHeaderT, UserContextT, UserInfoT } from "../types/Types";
import { IconArrowLeft } from "@tabler/icons-react";
import Admin from "../Pages/Admin Settings/Admin";
import "../components/style.scss";

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
    useKeyDown("Escape", (e) => {
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
                        height={50}
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
    ({ dispatch, loading }: HeaderBarProps) => {
        const { usr, authHeader, clear, serverInfo }: UserContextT =
            useContext(UserContext);
        const nav = useNavigate();
        const [userMenu, setUserMenu] = useState(false);
        const [settings, setSettings] = useState(false);
        const [admin, setAdmin] = useState(false);
        const [barRef, setBarRef] = useState(null);

        const barSize = useResize(barRef);

        return (
            <Box ref={setBarRef} className="z-50 h-max w-screen">
                {settings && (
                    <SettingsMenu
                        open={settings}
                        usr={usr}
                        setClosed={() => {
                            setSettings(false);
                            dispatch({ type: "set_block_focus", block: false });
                        }}
                        authHeader={authHeader}
                    />
                )}
                {admin && (
                    <div
                        className="settings-menu-container"
                        onClick={() => setAdmin(false)}
                    >
                        <Admin close={() => setAdmin(false)} />
                    </div>
                )}
                <div className=" absolute float-right right-10 bottom-8 z-20">
                    <WeblensLoader loading={loading} />
                </div>
                <div className="flex flex-row items-center h-14 pt-2 pb-2 border-b-2 border-neutral-700">
                    <div className="flex flex-row items-center w-96 shrink">
                        <div className="p-1" />
                        <WeblensButton
                            label="Timeline"
                            height={40}
                            width={barSize.width > 500 ? 112 : 40}
                            textMin={70}
                            centerContent
                            subtle
                            Left={<IconLibraryPhoto className="button-icon" />}
                            onClick={() => nav("/")}
                        />
                        <div className="p-1" />
                        <WeblensButton
                            label="Albums"
                            height={40}
                            width={barSize.width > 500 ? 110 : 40}
                            textMin={60}
                            centerContent
                            subtle
                            Left={<IconAlbum className="button-icon" />}
                            onClick={() => nav("/albums")}
                        />
                        <div className="p-1" />
                        <WeblensButton
                            label="Files"
                            height={40}
                            width={barSize.width > 500 ? 84 : 40}
                            textMin={50}
                            centerContent
                            subtle
                            Left={<IconFolder className="button-icon" />}
                            onClick={() => nav("/files/home")}
                        />
                    </div>
                    <div className="flex grow" />

                    {serverInfo && (
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
                            <p className="text-xs select-none">
                                {serverInfo.name}
                            </p>
                            <p className="text-xs select-none">
                                ({serverInfo.role})
                            </p>
                        </Box>
                    )}
                    <Tooltip
                        color="#222222"
                        label={
                            <Box>
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
                            </Box>
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
                                className="cursor-pointer shrink-0"
                            />
                        </Tooltip>
                    )}
                    <Menu opened={userMenu} onClose={() => setUserMenu(false)}>
                        <Menu.Target>
                            <Tooltip label={"User Settings"} color="#222222">
                                <IconUser
                                    size={32}
                                    onClick={() => setUserMenu(true)}
                                    className="cursor-pointer shrink-0"
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
                                    className="cursor-pointer shrink-0"
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
                </div>
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
