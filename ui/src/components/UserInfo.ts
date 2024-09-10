import API_ENDPOINT from '@weblens/api/ApiEndpoint'
import { getServerInfo } from '@weblens/api/ApiFetch'
import {
    LOGIN_TOKEN_COOKIE_KEY,
    ServerInfoT,
    UserInfoT,
    USERNAME_COOKIE_KEY,
} from '@weblens/types/Types'
import { useEffect } from 'react'
import { useCookies } from 'react-cookie'
import { create, StateCreator } from 'zustand'

const useR = () => {
    const [cookies] = useCookies([USERNAME_COOKIE_KEY, LOGIN_TOKEN_COOKIE_KEY])

    const { server, user, setUserInfo } = useSessionStore()

    useEffect(() => {
        if (!cookies[LOGIN_TOKEN_COOKIE_KEY]) {
            setUserInfo({ isLoggedIn: false } as UserInfoT)
        }
    }, [cookies])

    useEffect(() => {
        if (!server) {
            return
        }
        if (server.info.role === 'init') {
            setUserInfo({ isLoggedIn: false } as UserInfoT)
        }

        if (!user) {
            // Auth header set, but no user data, go get the user data
            const url = new URL(`${API_ENDPOINT}/user`)
            fetch(url.toString())
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
    }, [server])
}

export interface WeblensSessionT {
    user: UserInfoT
    server: { info: ServerInfoT; userCount: number; started: boolean }
    nav: (loc: string) => void

    setUserInfo: (user: UserInfoT) => void

    fetchServerInfo: () => void
    logout: (removeCookie: (cookieKey: string) => void) => void

    setNav: (navFunc: (loc: string) => void) => void
}

const WLStateControl: StateCreator<WeblensSessionT, [], []> = (set) => ({
    user: null,
    server: null,
    nav: null,

    setUserInfo: (user) => {
        if (user.isLoggedIn === undefined) {
            user.isLoggedIn = user.username !== ''
        }
        set({
            user: user,
        })
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
        })
        removeCoookie(USERNAME_COOKIE_KEY)
        removeCoookie(LOGIN_TOKEN_COOKIE_KEY)
    },

    setNav: (navFunc: (loc: string) => void) => {
        set({
            nav: navFunc,
        })
    },
})

export const useSessionStore = create<WeblensSessionT>()(WLStateControl)

export default useR
