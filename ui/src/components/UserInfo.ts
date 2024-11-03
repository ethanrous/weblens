import { getServerInfo } from '@weblens/api/ApiFetch'
import { GetUserInfo } from '@weblens/api/UserApi'
import {
    LOGIN_TOKEN_COOKIE_KEY,
    ServerInfoT,
    UserInfoT,
    USERNAME_COOKIE_KEY,
} from '@weblens/types/Types'
import { useEffect } from 'react'
import { useCookies } from 'react-cookie'
import { NavigateFunction, useNavigate } from 'react-router-dom'
import { create, StateCreator } from 'zustand'

const useR = () => {
    const nav = useNavigate()
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

        if (server.info.role === 'init' || !server.started) {
            setUserInfo({ isLoggedIn: false } as UserInfoT)
            return
        }

        if (!user || user.homeId === '') {
            GetUserInfo()
                .then((info) => setUserInfo({ ...info, isLoggedIn: true }))
                .catch((r) => {
                    setUserInfo({ isLoggedIn: false } as UserInfoT)
                    if (r === 401) {
                        console.debug('Going to login')
                        nav('/login')
                    }
                    console.error(r)
                })
        }
    }, [server])
}

export interface WeblensSessionT {
    user: UserInfoT
    server: { info: ServerInfoT; userCount: number; started: boolean }
    nav: NavigateFunction
    setUserInfo: (user: UserInfoT) => void

    fetchServerInfo: () => Promise<void>
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

    fetchServerInfo: async () => {
        return getServerInfo().then((r) => {
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
