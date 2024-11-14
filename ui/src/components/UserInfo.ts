import { ServersApi } from '@weblens/api/ServersApi'
import { ServerInfo } from '@weblens/api/swag'
import UsersApi from '@weblens/api/UserApi'
import User from '@weblens/types/user/User'
import { AxiosError } from 'axios'
import { useEffect } from 'react'
import { NavigateFunction, useNavigate } from 'react-router-dom'
import { create, StateCreator } from 'zustand'

const useR = () => {
    const nav = useNavigate()
    const { server, user, setUser } = useSessionStore()

    useEffect(() => {
        if (!server) {
            return
        }

        if (server.role === 'init' || !server.started) {
            const user = new User({})
            user.isLoggedIn = false
            setUser(user)
            return
        }

        if (!user || user.homeId === '') {
            UsersApi.getUser()
                .then((res) => {
                    setUser(new User(res.data, true))
                })
                .catch((err: AxiosError) => {
                    console.error(err.response.statusText)

                    setUser(new User({}, false))
                    if (
                        err.response.status === 401 &&
                        !window.location.pathname.includes('share')
                    ) {
                        console.debug('Going to login')
                        nav('/login', {
                            state: { returnTo: window.location.pathname },
                        })
                    }
                })
        }
    }, [server])
}

export interface WeblensSessionT {
    user: User
    server: ServerInfo
    serverFetchError: boolean
    nav: NavigateFunction
    setUser: (user: User) => void

    fetchServerInfo: () => Promise<void>

    setNav: (navFunc: NavigateFunction) => void
}

const WLStateControl: StateCreator<WeblensSessionT, [], []> = (set) => ({
    user: null,
    server: null,
    nav: null,
    serverFetchError: false,

    setUser: (user: User) => {
        if (user.isLoggedIn === undefined) {
            throw new Error('User must have isLoggedIn set')
        }

        set({
            user: user,
        })
    },

    fetchServerInfo: async () => {
        return ServersApi.getServerInfo()
            .then((res) => {
                set({
                    server: res.data,
                })
            })
            .catch((e) => {
                console.error('Failed to fetch server info', e)
                set({
                    serverFetchError: true,
                })
            })
    },

    setNav: (navFunc: NavigateFunction) => {
        set({
            nav: navFunc,
        })
    },
})

export const useSessionStore = create<WeblensSessionT>()(WLStateControl)

export default useR
