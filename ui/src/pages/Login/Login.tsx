import { Divider, Space } from '@mantine/core'
import { login } from '@weblens/api/ApiFetch'
import { useKeyDown } from '@weblens/components/hooks'
import WeblensLogo from '@weblens/components/Logo'
import { useSessionStore } from '@weblens/components/UserInfo'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import {
    LOGIN_TOKEN_COOKIE_KEY,
    USERNAME_COOKIE_KEY,
} from '@weblens/types/Types'
import { useCallback, useState } from 'react'
import { useCookies } from 'react-cookie'

const Login = () => {
    const [userInput, setUserInput] = useState('')
    const [passInput, setPassInput] = useState('')

    const setUser = useSessionStore((state) => state.setUserInfo)
    const [, setCookie] = useCookies([
        USERNAME_COOKIE_KEY,
        LOGIN_TOKEN_COOKIE_KEY,
    ])

    const [buttonRef, setButtonRef] = useState(null)
    useKeyDown('Enter', () => {
        if (buttonRef) {
            buttonRef.click()
        }
    })
    // const badUsername = userInput[0] === '.' || userInput.includes('/')

    const doLogin = useCallback(async (username: string, password: string) => {
        if (username === '' || password === '') {
            return Promise.reject('username and password must not be empty')
        }
        return login(username, password).then((data) => {
            setCookie(USERNAME_COOKIE_KEY, data.user.username)
            setCookie(LOGIN_TOKEN_COOKIE_KEY, data.token)
            setUser({ ...data.user, isLoggedIn: true })
        })
    }, [])

    return (
        <div className="flex flex-col h-screen w-screen items-center bg-wl-background pt-20">
            <WeblensLogo size={100} />
            <div className="bg-wl-barely-visible flex flex-col justify-center items-center wl-outline p-6 w-[400px] max-w-[600px] max-h-max mt-48">
                <h1 className="font-bold  m-6">Sign In</h1>
                <p className="w-full font-semibold">Username</p>
                <WeblensInput
                    value={userInput}
                    autoFocus
                    valueCallback={setUserInput}
                    squareSize={40}
                />
                <p className="w-full font-semibold">Password</p>
                <WeblensInput
                    value={passInput}
                    valueCallback={setPassInput}
                    squareSize={40}
                    password
                />
                <Space h={'md'} />
                <WeblensButton
                    label="Sign in"
                    fillWidth
                    squareSize={50}
                    disabled={userInput === '' || passInput === ''}
                    centerContent
                    onClick={async () => doLogin(userInput, passInput)}
                    setButtonRef={setButtonRef}
                />

            </div>
            <div className='flex flex-row items-center m-2 p-2 wl-outline-subtle'>
                <h3>New Here?</h3>
                <WeblensButton label="Sign up" />
            </div>
        </div>
    )
}

export default Login
