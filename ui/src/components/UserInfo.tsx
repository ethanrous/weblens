import { useEffect, useState } from "react"
import { useCookies } from "react-cookie"
import API_ENDPOINT from "../api/ApiEndpoint"
import { useNavigate } from "react-router-dom"

const useR = () => {
    const nav = useNavigate()
    const [cookies, setCookie, removeCookie] = useCookies(['weblens-username', 'weblens-login-token'])
    const [authHeader, setAuthHeader] = useState<{ "Authorization": string }>({ "Authorization": "" })
    const [userInfo, setUserInfo] = useState({})

    useEffect(() => {
        if (authHeader.Authorization === "" && cookies['weblens-username'] && cookies['weblens-login-token']) {
            // Auth header unset, but the cookies are ready
            setAuthHeader({ "Authorization": `${cookies['weblens-username']},${cookies['weblens-login-token']}` })
        } else if (authHeader.Authorization != "" && Object.keys(userInfo).length === 0) {
            // Auth header set, but no user data, go get the user data
            try {
                let url = new URL(`${API_ENDPOINT}/user`)
                fetch(url.toString(), { headers: authHeader })
                    .then(res => res.json())
                    .then(json => { setUserInfo(json) })
                    .catch(r => { console.log("useR naving", r); nav("/login", { state: { doLogin: false } }) })
            } catch {
                console.error("Failed to get user data outside promise")
            }
        } else if (authHeader.Authorization === "") {
            nav("/login", { state: { doLogin: false } })
        }
    }, [authHeader, cookies])
    return { authHeader, userInfo, setCookie, removeCookie }
}

export default useR