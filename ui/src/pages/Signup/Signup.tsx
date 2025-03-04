import { Loader } from '@mantine/core'
import { IconBrandGithub } from '@tabler/icons-react'
import UsersApi from '@weblens/api/UserApi'
import WeblensLogo from '@weblens/components/Logo'
// import { useKeyDown } from '@weblens/components/hooks'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import { ErrorHandler } from '@weblens/types/Types'
import { AxiosError } from 'axios'
import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'

export function SignupInputForm({
    setUsername,
    setPassword,
    setFullName,
    setError,
    disabled,
}: {
    setUsername: (v: string) => void
    setPassword: (v: string) => void
    setFullName: (v: string) => void
    setError: (v: string) => void
    disabled?: boolean
}) {
    const [nameInput, setNameInput] = useState('')
    const [userInput, setUserInput] = useState('')
    const [passInput, setPassInput] = useState('')
    const [usernameControlled, setUsernameControlled] = useState(true)
    const [validateLoading, setValidateLoading] = useState(false)
    const [formError, setFormError] = useState('')
    const [usernameValid, setUsernameValid] = useState<boolean | null>(null)
    const [usernameValidationError, setUsernameValidationError] =
        useState<string>('')

    useEffect(() => {
        setUsername(userInput)
    }, [userInput])
    useEffect(() => {
        setPassword(passInput)
    }, [passInput])
    useEffect(() => {
        setFullName(nameInput)
    }, [nameInput])
    useEffect(() => {
        setError(formError)
    }, [formError])

    useEffect(() => {
        if (formError !== '') {
            setFormError('')
        }
    }, [userInput, passInput])

    useEffect(() => {
        if (usernameControlled) {
            const newUsername = nameInput
                .replace(/[^a-zA-Z0-9]/g, '_')
                .toLowerCase()
            setUserInput(newUsername)
        }
    }, [nameInput])

    useEffect(() => {
        if (userInput === '') {
            setUsernameValid(null)
        } else {
            if (!userInput.match('^[a-z0-9_]+$')) {
                setUsernameValid(false)
                setUsernameValidationError(
                    "Username may only contain lowercase alphanumeric and '_'"
                )
                return () => {
                    setUsernameValid(null)
                    setUsernameValidationError('')
                }
            }

            let alive = true
            const timer = setTimeout(() => {
                setValidateLoading(true)
            }, 200)
            UsersApi.checkUsernameUnique(userInput)
                .then(() => {
                    if (!alive) {
                        console.log('Canceling')
                        return
                    }
                    setUsernameValid(true)
                })
                .catch(() => {
                    if (!alive) {
                        console.log('Canceling')
                        return
                    }
                    setUsernameValidationError('Username is taken')
                    setUsernameValid(false)
                })
                .finally(() => {
                    clearTimeout(timer)
                    setValidateLoading(false)
                })
            return () => {
                alive = false
                setUsernameValid(null)
                setUsernameValidationError('')
                clearTimeout(timer)
            }
        }
    }, [userInput])

    return (
        <div>
            <label htmlFor="name">
                <span>Full Name</span>
                <sup className="text-red-500">*</sup>
            </label>

            <WeblensInput
                placeholder="Mark Scout"
                value={nameInput}
                autoFocus
                valueCallback={setNameInput}
                squareSize={44}
                autoComplete="name"
                className="mb-5"
                disabled={disabled}
            />

            <label className="flex items-center" htmlFor="username">
                <span>Username</span>
                <sup className="h-max text-red-500">*</sup>
            </label>
            <div className="flex flex-row items-center gap-2">
                <WeblensInput
                    placeholder="mark_scout78"
                    value={userInput}
                    valueCallback={(v) => {
                        if (usernameControlled && v !== '') {
                            setUsernameControlled(false)
                        } else if (!usernameControlled && v === '') {
                            setUsernameControlled(true)
                        }
                        setUserInput(v)
                    }}
                    squareSize={44}
                    autoComplete="username"
                    valid={usernameValid}
                    disabled={disabled}
                />
            </div>
            <div className="my-2">
                {usernameValid === null && validateLoading && (
                    <Loader size={14} color="var(--wl-text-color-primary)" />
                )}

                <span className="text-red-500">{usernameValidationError}</span>
            </div>

            <label htmlFor="new-password">
                <span>Password</span>
                <sup className="text-red-500">*</sup>
            </label>
            <WeblensInput
                placeholder="*******"
                value={passInput}
                valueCallback={setPassInput}
                squareSize={44}
                password
                autoComplete="new-password"
                disabled={disabled}
            />
            {formError && (
                <span className="text-center text-red-500">{formError}</span>
            )}
        </div>
    )
}

function Signup() {
    const [nameInput, setNameInput] = useState('')
    const [userInput, setUserInput] = useState('')
    const [passInput, setPassInput] = useState('')
    const [loading, setLoading] = useState(false)
    const [formError, setFormError] = useState('')

    const [userCreated, setUserCreated] = useState(false)

    const nav = useNavigate()

    const doSignup = useCallback(
        async (fullName: string, username: string, password: string) => {
            setLoading(true)
            if (username === '' || password === '' || fullName === '') {
                return Promise.reject(
                    new Error('Name, Username and Password must not be empty')
                )
            }
            return UsersApi.createUser({ fullName, username, password })
                .then(() => {
                    setUserCreated(true)
                })
                .catch((err: AxiosError) => {
                    console.error(err)
                    setLoading(false)
                    setFormError('An error occurred')
                })
        },
        []
    )

    return (
        <div className="bg-wl-background bg-wl-background my-auto flex h-screen max-h-screen flex-col items-center justify-center gap-2 sm:justify-normal lg:my-0">
            <div className="flex w-full justify-center text-center sm:mt-72">
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
                    doSignup(nameInput, userInput, passInput).catch(
                        ErrorHandler
                    )
                }}
            >
                <h4 className="border-b-2 border-wl-border-color-primary">
                    Create an Account
                </h4>

                <SignupInputForm
                    setUsername={setUserInput}
                    setPassword={setPassInput}
                    setFullName={setNameInput}
                    setError={setFormError}
                    disabled={loading || userCreated}
                />

                {userCreated && (
                    <>
                        <span className="mt-2 text-center text-green-500">
                            Account created!
                        </span>
                        <span>
                            A server admin must activate your account before you
                            can log in.
                        </span>
                        <WeblensButton
                            label="Go To Login"
                            onClick={() => {
                                nav('/login')
                            }}
                            centerContent
                            className="mx-auto"
                        />
                    </>
                )}
                {!userCreated && (
                    <>
                        <div className="my-3">
                            <WeblensButton
                                label={
                                    loading ? 'Creating Account...' : 'Sign up'
                                }
                                fillWidth
                                squareSize={50}
                                disabled={
                                    nameInput === '' ||
                                    userInput === '' ||
                                    passInput === '' ||
                                    formError !== '' ||
                                    loading
                                }
                                centerContent
                                type="submit"
                            />
                        </div>
                        <div className="flex items-center justify-center gap-2 border-t-[1px] border-wl-border-color-primary p-2">
                            <span className="ml-auto text-wl-text-color-primary">
                                Already have an account?
                            </span>
                            <a href="/login">Login</a>
                        </div>
                    </>
                )}
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

export default Signup
