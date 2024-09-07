import { Divider, Space } from '@mantine/core'
import { useKeyDown } from '@weblens/components/hooks'
import { useSessionStore } from '@weblens/components/UserInfo'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import { useCallback, useState } from 'react'
import { useCookies } from 'react-cookie'
import { login } from '@weblens/api/ApiFetch'
import {
    LOGIN_TOKEN_COOKIE_KEY,
    USERNAME_COOKIE_KEY,
} from '@weblens/types/Types'

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
        <div
            className="flex flex-col h-screen w-screen items-center justify-center"
            style={{
                background:
                    'linear-gradient(45deg, rgba(2,0,36,1) 0%, rgba(94,43,173,1) 50%, rgba(0,212,255,1) 100%)',
            }}
        >
            <h1 className="text-5xl font-bold pb-12 select-none">
                Sign in to Weblens
            </h1>
            {/* <ScatteredPhotos /> */}
            <div className="flex flex-col justify-center items-center shadow-soft bg-bottom-grey outline outline-light-paper rounded-xl p-6 w-[400px] max-w-[600px] max-h-[400px]">
                <p className="w-full">Username</p>
                <WeblensInput
                    value={userInput}
                    autoFocus
                    valueCallback={setUserInput}
                    height={40}
                />
                <p className="w-full">Password</p>
                <WeblensInput
                    value={passInput}
                    valueCallback={setPassInput}
                    height={40}
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

                <Divider label={'or'} orientation={'horizontal'} />

                <p>New Here?</p>

                <WeblensButton label="Sign up" subtle />
            </div>
        </div>
    )
}

export default Login
