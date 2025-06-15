import { IconBrandGithub } from '@tabler/icons-react'
import { useQuery } from '@tanstack/react-query'
import MediaApi from '@weblens/api/MediaApi'
import UsersApi from '@weblens/api/UserApi'
import WeblensLogo from '@weblens/components/Logo.tsx'
import GridMotion from '@weblens/components/ScatteredPhotos'
import { useSessionStore } from '@weblens/components/UserInfo'
import WeblensButton from '@weblens/lib/WeblensButton.tsx'
import WeblensInput from '@weblens/lib/WeblensInput.tsx'
import { useFileBrowserStore } from '@weblens/store/FBStateControl'
import { ErrorHandler } from '@weblens/types/Types'
import WeblensMedia from '@weblens/types/media/Media'
import User from '@weblens/types/user/User'
import { shuffleArray } from '@weblens/util'
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
                useFileBrowserStore.getState().reset()
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

    const { data: medias } = useQuery({
        queryKey: ['loginImages'],
        queryFn: async () => {
            const mediaInfos = (await MediaApi.getRandomMedia(50)).data
            const medias = mediaInfos.Media?.map((mediaInfo) => {
                return new WeblensMedia(mediaInfo)
            })

            if (!medias) {
                return []
            }

            shuffleArray(medias)
            return medias
        },
    })

    return (
        <>
            <div className="absolute top-0 left-0 z-0 h-screen w-screen opacity-50">
                <GridMotion items={medias} />
            </div>

            <div className="bg-background-primary border-border-primary z-10 m-auto flex h-max max-h-screen flex-col items-center justify-center gap-2 rounded-2xl border p-8 shadow-2xl sm:justify-normal lg:my-0">
                <div className="flex w-full justify-center text-center">
                    <WeblensLogo size={100} />
                    <h1 className="mt-auto">EBLENS</h1>
                </div>
                <form
                    id="login"
                    action="#"
                    className="mx-auto mt-8 flex w-96 max-w-full min-w-0 flex-col gap-3 px-4"
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
                    <div className="border-color-border-primary flex items-center justify-center gap-2 border-t-[1px] p-2">
                        <span className="text-color-text-primary ml-auto">
                            New Here?
                        </span>
                        <a href="/signup">Request an Account</a>
                    </div>
                </form>
                <a
                    href="https://github.com/ethanrous/weblens"
                    className="absolute right-0 bottom-0 m-4 flex flex-row bg-transparent"
                    target="_blank"
                    rel="noreferrer"
                >
                    <IconBrandGithub />
                    GitHub
                </a>
            </div>
        </>
    )
}

export default Login
