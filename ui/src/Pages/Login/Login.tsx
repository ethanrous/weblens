import { useCallback, useContext, useEffect, useRef, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { createUser, login } from "../../api/ApiFetch";
import { userContext } from "../../Context";
import { notifications } from "@mantine/notifications";
import {
    Box,
    Button,
    Fieldset,
    Input,
    Loader,
    PasswordInput,
    Space,
    Tabs,
    TextInput,
} from "@mantine/core";
import { RowBox } from "../FileBrowser/FilebrowserStyles";
import { WeblensButton } from "../../components/WeblensButton";
import { useKeyDown } from "../../components/hooks";
import { UserContextT } from "../../types/Types";
import { ScatteredPhotos } from "../../components/ScatteredPhotos";

async function CheckCreds(username: string, password: string, setCookie, nav) {
    return await login(username, password)
        .then((res) => {
            if (res.status !== 200) {
                return Promise.reject("Incorrect username or password");
            } else {
                return res.json();
            }
        })
        .then((data) => {
            setCookie("weblens-username", username, { sameSite: "strict" });
            setCookie("weblens-login-token", data.token, {
                sameSite: "strict",
            });
            nav("/");
            return true;
        })
        .catch((r) => {
            notifications.show({ message: String(r), color: "red" });
            return false;
        });
}

function CreateUser(username: string, password: string) {
    createUser(username, password)
        .then((x) => {
            notifications.show({
                message:
                    "Account created! Once an administrator activates your account you may login",
            });
        })
        .catch((reason) => {
            notifications.show({
                message: `Failed to create new user: ${String(reason)}`,
                color: "red",
            });
        });
}

export const useKeyDownLogin = (login) => {
    const onKeyDown = useCallback(
        (event) => {
            if (event.key === "Enter") {
                login();
            }
        },
        [login]
    );

    useEffect(() => {
        document.addEventListener("keydown", onKeyDown);
        return () => {
            document.removeEventListener("keydown", onKeyDown);
        };
    }, [onKeyDown]);
};

const Login = () => {
    const [userInput, setUserInput] = useState("");
    const [passInput, setPassInput] = useState("");
    const [tab, setTab] = useState("login");
    const nav = useNavigate();
    const loc = useLocation();
    const { authHeader, setCookie }: UserContextT = useContext(userContext);

    useEffect(() => {
        if (loc.state == null && authHeader.Authorization !== "") {
            nav("/");
        }
    }, [authHeader, loc.state, nav]);

    const [buttonRef, setButtonRef] = useState(null);
    useKeyDown("Enter", (e) => {
        if (buttonRef) {
            buttonRef.click();
        }
    });
    const badUsername = userInput[0] === "." || userInput.includes("/");

    return (
        <RowBox
            style={{
                height: "100vh",
                width: "100vw",
                justifyContent: "center",
                background:
                    "linear-gradient(45deg, rgba(2,0,36,1) 0%, rgba(94,43,173,1) 50%, rgba(0,212,255,1) 100%)",
            }}
        >
            {/*<ScatteredPhotos />*/}
            <Box
                style={{
                    width: 400,
                    maxWidth: 600,
                    maxHeight: 400,
                    display: "flex",
                    flexDirection: "column",
                    justifyContent: "center",
                    alignItems: "center",
                    backgroundColor: "#111111aa",
                    borderRadius: 8,
                    padding: 24,
                }}
            >
                <Tabs
                    value={tab}
                    onChange={setTab}
                    keepMounted={false}
                    variant="pills"
                    style={{
                        width: "100%",
                        height: "90%",
                        justifyContent: "center",
                        alignItems: "center",
                        display: "flex",
                        flexDirection: "column",
                        gap: 20,
                    }}
                >
                    <Tabs.List grow style={{ width: "100%" }}>
                        <Tabs.Tab value="login" className="menu-tab">
                            Login
                        </Tabs.Tab>
                        <Tabs.Tab value="signup" className="menu-tab">
                            Sign Up
                        </Tabs.Tab>
                    </Tabs.List>
                    <Tabs.Panel
                        value="login"
                        style={{
                            display: "flex",
                            flexDirection: "column",
                            justifyContent: "center",
                            alignItems: "center",
                            width: "100%",
                        }}
                    >
                        <Input
                            className="weblens-input-wrapper"
                            classNames={{ input: "weblens-input" }}
                            variant="unstyled"
                            value={userInput}
                            placeholder="Username"
                            style={{ width: "100%" }}
                            onChange={(event) =>
                                setUserInput(event.currentTarget.value)
                            }
                        />
                        <Input
                            className="weblens-input-wrapper"
                            classNames={{ input: "weblens-input" }}
                            variant="unstyled"
                            type="password"
                            value={passInput}
                            placeholder="Password"
                            style={{ width: "100%" }}
                            onChange={(event) =>
                                setPassInput(event.currentTarget.value)
                            }
                        />
                        <Space h={"md"} />
                        <WeblensButton
                            label="Login"
                            disabled={userInput === "" || passInput === ""}
                            centerContent
                            onClick={() =>
                                CheckCreds(userInput, passInput, setCookie, nav)
                            }
                            setButtonRef={setButtonRef}
                            style={{ width: "100%" }}
                        />
                    </Tabs.Panel>
                    <Tabs.Panel
                        value="signup"
                        style={{
                            display: "flex",
                            flexDirection: "column",
                            justifyContent: "center",
                            alignItems: "center",
                            width: "100%",
                        }}
                    >
                        <Input
                            className="weblens-input-wrapper"
                            classNames={{ input: "weblens-input" }}
                            variant="unstyled"
                            value={userInput}
                            placeholder="Username"
                            error={badUsername}
                            onChange={(event) =>
                                setUserInput(event.currentTarget.value)
                            }
                            style={{ width: "100%" }}
                        />
                        {badUsername && (
                            <Input.Error style={{ width: "100%" }}>
                                Username must not begin with '.' and cannot
                                include '/'
                            </Input.Error>
                        )}
                        <Input
                            className="weblens-input-wrapper"
                            classNames={{ input: "weblens-input" }}
                            variant="unstyled"
                            type="password"
                            value={passInput}
                            placeholder="Password"
                            onChange={(event) =>
                                setPassInput(event.currentTarget.value)
                            }
                            style={{ width: "100%" }}
                        />
                        <Space h={"md"} />
                        <WeblensButton
                            label="Sign Up"
                            disabled={
                                userInput === "" ||
                                passInput === "" ||
                                badUsername
                            }
                            centerContent
                            onClick={() => CreateUser(userInput, passInput)}
                            setButtonRef={setButtonRef}
                            style={{ width: "100%" }}
                        />
                    </Tabs.Panel>
                </Tabs>
            </Box>
        </RowBox>
    );
};

export default Login;
