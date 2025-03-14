import { IconBrandGithub } from '@tabler/icons-react'
import UsersApi from '@weblens/api/UserApi'
import WeblensLogo from '@weblens/components/Logo'
import { useSessionStore } from '@weblens/components/UserInfo'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import { ErrorHandler } from '@weblens/types/Types'
import User from '@weblens/types/user/User'
import { AxiosError } from 'axios'
import { useCallback, useEffect, useState } from 'react'

const Login = () => {
    const [userInput, setUserInput] = useState('')
    const [passInput, setPassInput] = useState('')
    const [loading, setLoading] = useState(false)
    const [formError, setFormError] = useState('')

    const setUser = useSessionStore((state) => state.setUser)
    const doLogin = useCallback(async (username: string, password: string) => {
        setLoading(true)
        if (username === '' || password === '') {
            return Promise.reject(
                new Error('username and password must not be empty')
            )
        }
        return UsersApi.loginUser({ username, password })
            .then((res) => {
                const user = new User(res.data)
                user.isLoggedIn = true
                setUser(user)
            })
            .catch((err: AxiosError) => {
                setLoading(false)
                if (err.status === 401) {
                    setFormError('Invalid username or password')
                } else {
                    setFormError('An error occurred')
                }
            })
    }, [])

    useEffect(() => {
        if (formError !== '') {
            setFormError('')
        }
    }, [userInput, passInput])

    return (
        <div className="bg-wl-background my-auto flex h-screen max-h-screen flex-col items-center justify-center gap-2 sm:justify-normal lg:my-0">
            <div className="flex w-full justify-center text-center sm:mt-80">
                <WeblensLogo size={100} />
                <h1 className="mt-auto">EBLENS</h1>
            </div>
            <form
                id="login"
                action="#"
                className="mx-auto mt-8 flex w-96 min-w-0 max-w-full flex-col gap-3 px-4"
                onSubmit={(e) => {
                    e.preventDefault()
                    e.stopPropagation()
                    doLogin(userInput, passInput).catch(ErrorHandler)
                }}
            >
                <WeblensInput
                    placeholder="Username"
                    value={userInput}
                    autoFocus
                    valueCallback={setUserInput}
                    squareSize={44}
                    autoComplete="username"
                />
                <WeblensInput
                    placeholder="Password"
                    value={passInput}
                    valueCallback={setPassInput}
                    squareSize={44}
                    password
                    autoComplete="current-password"
                />
                {formError && (
                    <span className="text-center text-red-500">
                        {formError}
                    </span>
                )}
                <div className="my-3">
                    <WeblensButton
                        label={loading ? 'Signing in...' : 'Sign in'}
                        fillWidth
                        squareSize={50}
                        disabled={
                            userInput === '' || passInput === '' || loading
                        }
                        centerContent
                        type="submit"
                        // setButtonRef={setButtonRef}
                    />
                </div>
                <div className="flex items-center justify-center gap-2 border-t-[1px] border-color-border-primary p-2">
                    <span className="ml-auto text-color-text-primary">
                        New Here?
                    </span>
                    <a href="/signup">Request an Account</a>
                </div>
            </form>
            <a
                href="https://github.com/ethanrous/weblens"
                className="absolute bottom-0 right-0 m-4 flex flex-row bg-transparent"
                target="_blank"
            >
                <IconBrandGithub />
                GitHub
            </a>
        </div>
    )
}

export default Login
