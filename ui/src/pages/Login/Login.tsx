import { Space } from '@mantine/core'
import { IconBrandGithub } from '@tabler/icons-react'
import UsersApi from '@weblens/api/UserApi'
import WeblensLogo from '@weblens/components/Logo'
import { useSessionStore } from '@weblens/components/UserInfo'
import { useKeyDown } from '@weblens/components/hooks'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import User from '@weblens/types/user/User'
import { useCallback, useState } from 'react'

import loginStyle from './loginStyle.module.scss'

const Login = () => {
    const [userInput, setUserInput] = useState('')
    const [passInput, setPassInput] = useState('')

    const setUser = useSessionStore((state) => state.setUser)

    const [buttonRef, setButtonRef] = useState<HTMLDivElement>(null)
    useKeyDown('Enter', () => {
        if (buttonRef) {
            buttonRef.click()
        }
    })
    // const badUsername = userInput[0] === '.' || userInput.includes('/')

    const doLogin = useCallback(async (username: string, password: string) => {
        if (username === '' || password === '') {
            return Promise.reject(
                new Error('username and password must not be empty')
            )
        }
        return UsersApi.loginUser({ username, password }).then((res) => {
            const user = new User(res.data)
            user.isLoggedIn = true
            setUser(user)
        })
    }, [])

    return (
        <div className="h-screen max-h-screen items-center bg-wl-background gap-2 my-0 m-[0 auto]">
            <div className="flex justify-center w-full text-center">
                <WeblensLogo className="mt-10" size={100} />
            </div>
            <div className={loginStyle['login-form']}>
                <div className="w-full text-center mb-4">
                    <h1>Sign in to Weblens</h1>
                </div>
                <div className={loginStyle['login-box']}>
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
                <div className="flex justify-center items-center p-4 wl-outline-subtle gap-2 mt-3">
                    <h3>New Here?</h3>
                    <a href="/signup">Request an Account</a>
                </div>
            </div>
            <a
                href="https://github.com/ethanrous/weblens"
                className="flex flex-row absolute bottom-0 right-0 m-4 bg-transparent"
                target="_blank"
            >
                <IconBrandGithub />
                GitHub
            </a>
        </div>
    )
}

export default Login
