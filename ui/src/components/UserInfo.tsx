import { useEffect, useState } from "react";
import { useCookies } from "react-cookie";
import API_ENDPOINT from "../api/ApiEndpoint";
import { notifications } from "@mantine/notifications";
import { UserContextT, UserInfoT } from "../types/Types";
import { useNavigate } from "react-router-dom";
import { getServerInfo } from "../api/ApiFetch";

const useR = () => {
    const nav = useNavigate();
    const [cookies, setCookie, removeCookie] = useCookies([
        "weblens-username",
        "weblens-login-token",
    ]);
    const [authHeader, setAuthHeader] = useState<{ Authorization: string }>({
        Authorization: "",
    });
    const [usr, setUserInfo]: [usr: UserInfoT, setUserInfo: any] = useState({
        homeId: "",
        trashId: "",
        username: "",
        admin: false,
        owner: false,
        activated: false,
        isLoggedIn: undefined,
    });
    const [serverInfo, setServerInfo] = useState(null);

    const clear = () => {
        setAuthHeader({ Authorization: "" });
        setUserInfo({
            admin: false,
            homeId: "",
            trashId: "",
            username: "",
            activated: false,
            owner: false,
            isLoggedIn: false,
        } as UserInfoT);
        removeCookie("weblens-username");
        removeCookie("weblens-login-token");
    };

    const inSetup = window.location.pathname === "/setup";
    useEffect(() => {
        getServerInfo().then((r) => {
            if ((r === 307 || r.info.userCount === 0) && !inSetup) {
                setServerInfo({ name: "" });
                nav("/setup");
                return;
            } else if (r !== 307 && r.info.userCount !== 0 && inSetup) {
                nav("/");
            }
            setServerInfo({ ...r.info });
        });
    }, []);

    useEffect(() => {
        if (
            authHeader.Authorization === "" &&
            cookies["weblens-username"] &&
            cookies["weblens-login-token"]
        ) {
            // Auth header unset, but the cookies are ready
            const loginStr = `${cookies["weblens-username"]}:${cookies["weblens-login-token"]}`;
            // console.log(loginStr.replace(/-/g, '+').replace(/_/g, '/'));
            const login64 = window.btoa(loginStr);
            setAuthHeader({
                Authorization: `Basic ${login64.toString()}`,
            });
        } else if (
            authHeader.Authorization !== "" &&
            (usr.username === "" || Object.keys(usr).length === 0)
        ) {
            // Auth header set, but no user data, go get the user data
            let url = new URL(`${API_ENDPOINT}/user`);
            fetch(url.toString(), { headers: authHeader })
                .then((res) => {
                    if (res.status === 307 && !inSetup) {
                        nav("/setup");
                    } else if (res.status !== 200) {
                        return Promise.reject(res.statusText);
                    }
                    return res.json();
                })
                .then((json) => {
                    if (!json) {
                        return Promise.reject("Invalid user data");
                    }
                    setUserInfo({ ...json, isLoggedIn: true });
                })
                .catch((r) =>
                    notifications.show({
                        title: "Failed to fetch user info",
                        message: String(r),
                        color: "red",
                    })
                );
        } else if (authHeader.Authorization === "") {
            setUserInfo((p) => {
                p.isLoggedIn = false;
                return { ...p };
            });
        }
    }, [authHeader, cookies]);

    return { authHeader, usr, setCookie, clear, serverInfo } as UserContextT;
};

export default useR;
