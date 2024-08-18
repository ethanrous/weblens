import { useEffect } from 'react'
import { useCookies } from 'react-cookie'
import API_ENDPOINT from '../api/ApiEndpoint'
import {
    AuthHeaderT,
    LOGIN_TOKEN_COOKIE_KEY,
    ServerInfoT,
    UserInfoT,
    USERNAME_COOKIE_KEY,
} from '../types/Types'
import { getServerInfo } from '../api/ApiFetch'
import { create, StateCreator } from 'zustand'

const useR = () => {
    const [cookies] = useCookies([USERNAME_COOKIE_KEY, LOGIN_TOKEN_COOKIE_KEY])

    const { server, user, auth, setAuthHeader, setUserInfo } = useSessionStore()

    useEffect(() => {
        if (cookies[LOGIN_TOKEN_COOKIE_KEY] && auth === null) {
            // Auth header unset, but the cookies are ready
            const loginStr = `${cookies[USERNAME_COOKIE_KEY]}:${cookies[LOGIN_TOKEN_COOKIE_KEY]}`
            const login64 = window.btoa(loginStr)
            setAuthHeader(login64.toString())
        } else if (!cookies[LOGIN_TOKEN_COOKIE_KEY]) {
            setUserInfo({ isLoggedIn: false } as UserInfoT)
        }
    }, [cookies])

    useEffect(() => {
        if (!server || server.info.role === 'init') {
            return
        }

        if (auth && !user) {
            // Auth header set, but no user data, go get the user data
            const url = new URL(`${API_ENDPOINT}/user`)
            fetch(url.toString(), { headers: auth })
                .then((res) => {
                    if (res.status !== 200) {
                        return Promise.reject(res.statusText)
                    }
                    return res.json()
                })
                .then((json) => {
                    if (!json) {
                        return Promise.reject('Invalid user data')
                    }
                    setUserInfo({ ...json, isLoggedIn: true })
                })
                .catch((r) => {
                    console.error(r)
                    setUserInfo({ isLoggedIn: false } as UserInfoT)
                })
        }
    }, [server, auth])
}

export interface WeblensSessionT {
    user: UserInfoT
    server: { info: ServerInfoT; userCount: number; started: boolean }
    auth: AuthHeaderT

    setUserInfo: (user: UserInfoT) => void
    setAuthHeader: (token: string) => void

    fetchServerInfo: () => void
    logout: (removeCookie: (cookieKey: string) => void) => void
}

const WLStateControl: StateCreator<WeblensSessionT, [], []> = (set) => ({
    user: null,
    server: null,
    auth: null,

    setUserInfo: (user) => {
        if (user.isLoggedIn === undefined) {
            user.isLoggedIn = user.username !== ''
        }
        set({
            user: user,
        })
    },

    setAuthHeader: (token: string) => {
        if (!token) {
            set({
                auth: null,
            })
        } else {
            set({
                auth: { Authorization: `Basic ${token}` },
            })
        }
    },

    fetchServerInfo: () => {
        getServerInfo().then((r) => {
            set({
                server: r,
            })
        })
    },

    logout: (removeCoookie) => {
        set({
            user: null,
            auth: null,
        })
        removeCoookie(USERNAME_COOKIE_KEY)
        removeCoookie(LOGIN_TOKEN_COOKIE_KEY)
    },
})

export const useSessionStore = create<WeblensSessionT>()(WLStateControl)

export default useR
