import { useCallback, useEffect, useState } from 'react'
import { createUser, login } from '../../api/ApiFetch'
import { Divider, Input, Space, Tabs } from '@mantine/core'
import WeblensButton from '../../components/WeblensButton'
import { useKeyDown } from '../../components/hooks'
import WeblensInput from '../../components/WeblensInput'
import { useSessionStore } from '../../components/UserInfo'
import { useCookies } from 'react-cookie'
import {
    LOGIN_TOKEN_COOKIE_KEY,
    UserInfoT,
    USERNAME_COOKIE_KEY,
} from '../../types/Types'

async function CheckCreds(
    username: string,
    password: string,
    setCookie,
    setUser: (user: UserInfoT) => void
) {
    if (!username || !password) {
        return false
    }
    return await login(username, password)
        .then((data) => {
            setCookie(USERNAME_COOKIE_KEY, data.user.username)
            setCookie(LOGIN_TOKEN_COOKIE_KEY, data.token)

            setUser({ ...data.user, isLoggedIn: true })

            return true
        })
        .catch((r) => {
            console.error(r)
            return false
        })
}

async function CreateUser(
    username: string,
    password: string
): Promise<boolean> {
    return await createUser(username, password)
        .then(() => true)
        .catch((r) => {
            console.error(r)
            return false
        })
}

export const useKeyDownLogin = (login) => {
    const onKeyDown = useCallback(
        (event) => {
            if (event.key === 'Enter') {
                login()
            }
        },
        [login]
    )

    useEffect(() => {
        document.addEventListener('keydown', onKeyDown)
        return () => {
            document.removeEventListener('keydown', onKeyDown)
        }
    }, [onKeyDown])
}

const Login = () => {
    const [userInput, setUserInput] = useState('')
    const [passInput, setPassInput] = useState('')
    const [tab, setTab] = useState('login')

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
    const badUsername = userInput[0] === '.' || userInput.includes('/')

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
                    onComplete={() => {}}
                    valueCallback={setUserInput}
                    height={40}
                />
                <p className="w-full">Password</p>
                <WeblensInput
                    value={passInput}
                    onComplete={() => {}}
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
                    onClick={() =>
                        CheckCreds(userInput, passInput, setCookie, setUser)
                    }
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
