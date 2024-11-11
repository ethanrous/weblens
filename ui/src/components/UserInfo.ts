import ServerApi from '@weblens/api/ServerApi'
import { ServerInfo } from '@weblens/api/swag'
import UsersApi from '@weblens/api/UserApi'
import User from '@weblens/types/user/User'
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
                .catch((e) => {
                    console.error(e.response.statusText)

                    setUser(new User({}, false))
                    if (
                        e.response.status === 401 &&
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
    nav: NavigateFunction
    setUser: (user: User) => void

    fetchServerInfo: () => Promise<void>

    setNav: (navFunc: NavigateFunction) => void
}

const WLStateControl: StateCreator<WeblensSessionT, [], []> = (set) => ({
    user: null,
    server: null,
    nav: null,

    setUser: (user: User) => {
        if (user.isLoggedIn === undefined) {
            throw new Error('User must have isLoggedIn set')
        }

        set({
            user: user,
        })
    },

    fetchServerInfo: async () => {
        return ServerApi.getServerInfo().then((res) => {
            set({
                server: res.data,
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
