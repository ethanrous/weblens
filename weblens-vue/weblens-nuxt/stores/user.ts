import type { UserInfo } from '@ethanrous/weblens-api'
import { useWeblensAPI } from '~/api/AllApi'
import User, { UNAUTHENTICATED_USER_NAME } from '~/types/user'

export const useUserStore = defineStore('user', () => {
    const user = shallowRef(new User())

    const loggedIn = computed(() => user.value.isLoggedIn.get({ default: false }))

    function setUser(info: UserInfo, isLoggedIn: boolean = false) {
        console.debug('Setting user info:', info, 'Logged in:', isLoggedIn)
        user.value = new User(info, isLoggedIn)
    }

    async function loadUser(): Promise<void> {
        if (user.value.isLoggedIn.isSet()) {
            return
        }

        await useWeblensAPI()
            .UsersAPI.getUser()
            .then((res: { data: UserInfo }) => setUser(res.data, true))
            .catch(() => setUser({} as UserInfo, false))

        console.debug('Loading user info...', user.value)
    }

    async function logout(): Promise<void> {
        if (loggedIn.value) {
            await useWeblensAPI().UsersAPI.logoutUser()
        }

        await navigateTo('/login')
    }

    function getActiveUsername(): string {
        console.debug('Getting active username...', user.value)
        return user.value.isLoggedIn.get({ default: false }) ? user.value.username : UNAUTHENTICATED_USER_NAME
    }

    onMounted(() => {
        loadUser()
    })

    return {
        user,
        loggedIn,

        setUser,
        logout,
        getActiveUsername,
    }
})
