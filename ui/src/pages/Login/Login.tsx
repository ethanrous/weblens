import { IconBrandGithub } from '@tabler/icons-react'
import UsersApi from '@weblens/api/UserApi'
import WeblensLogo from '@weblens/components/Logo'
import { useSessionStore } from '@weblens/components/UserInfo'
// import { useKeyDown } from '@weblens/components/hooks'
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
        <div className="flex flex-col h-screen max-h-screen items-center bg-wl-background gap-2">
            <div className="flex justify-center w-full text-center mt-80">
                <WeblensLogo size={100} />
                <h1 className="mt-auto">EBLENS</h1>
            </div>
            <form
                id="login"
                action="#"
                className="flex flex-col gap-3 w-96 mx-auto mt-8"
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
                    <span className="text-red-500 text-center">
                        {formError}
                    </span>
                )}
				<div className='my-3'>
					<WeblensButton
						label={loading ? 'Signing in...' : 'Sign in'}
						fillWidth
						squareSize={50}
						disabled={userInput === '' || passInput === '' || loading}
						centerContent
						type="submit"
						// setButtonRef={setButtonRef}
					/>
				</div>
                <div className="flex justify-center items-center p-2 gap-2 border-t-2 border-wl-color-graphite-800">
                    <span className="text-wl-text-color-secondary ml-auto">
                        New Here?
                    </span>
                    <a href="/signup">Request an Account</a>
                </div>
            </form>
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
