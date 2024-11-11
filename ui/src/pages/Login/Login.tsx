import { Space } from '@mantine/core'
import UsersApi from '@weblens/api/UserApi'
import { useKeyDown } from '@weblens/components/hooks'
import WeblensLogo from '@weblens/components/Logo'
import { useSessionStore } from '@weblens/components/UserInfo'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import User from '@weblens/types/user/user'
import { useCallback, useState } from 'react'

const Login = () => {
    const [userInput, setUserInput] = useState('')
    const [passInput, setPassInput] = useState('')

    const setUser = useSessionStore((state) => state.setUser)

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
        return UsersApi.loginUser({ username, password }).then((res) => {
            const user = new User(res.data)
            user.isLoggedIn = true
            setUser(user)
        })
    }, [])

    return (
        <div className="flex flex-col h-screen max-h-screen w-screen items-center bg-wl-background pt-[8vh]">
            <WeblensLogo size={100} />
            <div className="bg-wl-barely-visible flex flex-col justify-center items-center wl-outline p-6 w-[400px] max-w-[600px] max-h-[50vh] mt-[8vh]">
                <h1 className="font-bold  m-[2vh]">Sign In</h1>
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
            <div className="flex flex-row items-center m-8 p-4 wl-outline-subtle gap-2">
                <h3>New Here?</h3>
                <WeblensButton label="Sign up" />
            </div>
        </div>
    )
}

export default Login
