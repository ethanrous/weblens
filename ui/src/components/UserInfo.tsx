import { useEffect, useState } from "react"
import { useCookies } from "react-cookie"
import API_ENDPOINT from "../api/ApiEndpoint"
import { useNavigate } from "react-router-dom"
import { notifications } from "@mantine/notifications"

const useR = () => {
    const nav = useNavigate()
    const [cookies, setCookie, removeCookie] = useCookies(['weblens-username', 'weblens-login-token'])
    const [authHeader, setAuthHeader] = useState<{ "Authorization": string }>({ "Authorization": "" })
    const [userInfo, setUserInfo] = useState({
        admin: false,
        homeFolderId: "",
        trashFolderId: "",
        username: ""
    })
    const clear = () => {
        setAuthHeader({ "Authorization": "" })
        setUserInfo({
            admin: false,
            homeFolderId: "",
            trashFolderId: "",
            username: ""
        })
        removeCookie("weblens-username")
        removeCookie("weblens-login-token")
    }

    useEffect(() => {
        if (authHeader.Authorization === "" && cookies['weblens-username'] && cookies['weblens-login-token']) {
            // Auth header unset, but the cookies are ready
            setAuthHeader({ "Authorization": `${cookies['weblens-username']},${cookies['weblens-login-token']}` })
        } else if (authHeader.Authorization !== "" && (userInfo.username === "" || Object.keys(userInfo).length === 0)) {
            // Auth header set, but no user data, go get the user data

            let url = new URL(`${API_ENDPOINT}/user`)
            fetch(url.toString(), { headers: authHeader })
                .then(res => res.json())
                .then(json => setUserInfo(json))
                .catch(r => notifications.show({ message: String(r) }))

        } else if (authHeader.Authorization === "") {
            // nav("/login")
        }
    }, [authHeader, cookies])

    return { authHeader, userInfo, setCookie, clear }
}

export default useR