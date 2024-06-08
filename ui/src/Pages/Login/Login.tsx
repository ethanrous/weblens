import { useCallback, useContext, useEffect, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { createUser, login } from '../../api/ApiFetch'
import { UserContext } from '../../Context'
import { notifications } from '@mantine/notifications'
import { Input, Space, Tabs } from '@mantine/core'
import { WeblensButton } from '../../components/WeblensButton'
import { useKeyDown } from '../../components/hooks'
import { UserContextT } from '../../types/Types'

async function CheckCreds(username: string, password: string, setCookie, nav) {
    return await login(username, password)
        .then((res) => {
            if (res.status !== 200) {
                return Promise.reject('Incorrect username or password')
            } else {
                return res.json()
            }
        })
        .then((data) => {
            setCookie('weblens-username', username, { sameSite: 'strict' })
            setCookie('weblens-login-token', data.token, {
                sameSite: 'strict',
            })
            nav('/')
            return true
        })
        .catch((r) => {
            notifications.show({ message: String(r), color: 'red' })
            return false
        })
}

function CreateUser(username: string, password: string) {
    createUser(username, password)
        .then((x) => {
            notifications.show({
                message:
                    'Account created! Once an administrator activates your account you may login',
            })
        })
        .catch((reason) => {
            notifications.show({
                message: `Failed to create new user: ${String(reason)}`,
                color: 'red',
            })
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
    const nav = useNavigate()
    const loc = useLocation()
    const { authHeader, setCookie }: UserContextT = useContext(UserContext)

    useEffect(() => {
        if (loc.state == null && authHeader.Authorization !== '') {
            nav('/')
        }
    }, [authHeader, loc.state, nav])

    const [buttonRef, setButtonRef] = useState(null)
    useKeyDown('Enter', (e) => {
        if (buttonRef) {
            buttonRef.click()
        }
    })
    const badUsername = userInput[0] === '.' || userInput.includes('/')

    return (
        <div
            className="flex flex-row h-screen w-screen items-center justify-center"
            style={{
                background:
                    'linear-gradient(45deg, rgba(2,0,36,1) 0%, rgba(94,43,173,1) 50%, rgba(0,212,255,1) 100%)',
            }}
        >
            {/* <ScatteredPhotos /> */}
            <div className="flex flex-col justify-center items-center bg-dark-paper rounded-xl p-6 w-[400px] max-w-[600px] max-h-[400px]">
                <Tabs
                    value={tab}
                    onChange={setTab}
                    keepMounted={false}
                    variant="pills"
                    style={{
                        width: '100%',
                        height: '90%',
                        justifyContent: 'center',
                        alignItems: 'center',
                        display: 'flex',
                        flexDirection: 'column',
                        gap: 20,
                    }}
                >
                    <Tabs.List grow style={{ width: '100%' }}>
                        <Tabs.Tab value="login" className="menu-tab">
                            Login
                        </Tabs.Tab>
                        <Tabs.Tab value="sign-up" className="menu-tab">
                            Sign Up
                        </Tabs.Tab>
                    </Tabs.List>
                    <Tabs.Panel
                        value="login"
                        style={{
                            display: 'flex',
                            flexDirection: 'column',
                            justifyContent: 'center',
                            alignItems: 'center',
                            width: '100%',
                        }}
                    >
                        <Input
                            className="weblens-input-wrapper"
                            classNames={{ input: 'weblens-input' }}
                            variant="unstyled"
                            value={userInput}
                            placeholder="Username"
                            style={{ width: '100%' }}
                            onChange={(event) =>
                                setUserInput(event.currentTarget.value)
                            }
                        />
                        <Input
                            className="weblens-input-wrapper"
                            classNames={{ input: 'weblens-input' }}
                            variant="unstyled"
                            type="password"
                            value={passInput}
                            placeholder="Password"
                            style={{ width: '100%' }}
                            onChange={(event) =>
                                setPassInput(event.currentTarget.value)
                            }
                        />
                        <Space h={'md'} />
                        <WeblensButton
                            label="Login"
                            squareSize={50}
                            disabled={userInput === '' || passInput === ''}
                            centerContent
                            onClick={() =>
                                CheckCreds(userInput, passInput, setCookie, nav)
                            }
                            setButtonRef={setButtonRef}
                            style={{ width: '100%' }}
                        />
                    </Tabs.Panel>
                    <Tabs.Panel
                        value="sign-up"
                        style={{
                            display: 'flex',
                            flexDirection: 'column',
                            justifyContent: 'center',
                            alignItems: 'center',
                            width: '100%',
                        }}
                    >
                        <Input
                            className="weblens-input-wrapper"
                            classNames={{ input: 'weblens-input' }}
                            variant="unstyled"
                            value={userInput}
                            placeholder="Username"
                            error={badUsername}
                            onChange={(event) =>
                                setUserInput(event.currentTarget.value)
                            }
                            style={{ width: '100%' }}
                        />
                        {badUsername && (
                            <Input.Error style={{ width: '100%' }}>
                                Username must not begin with '.' and cannot
                                include '/'
                            </Input.Error>
                        )}
                        <Input
                            className="weblens-input-wrapper"
                            classNames={{ input: 'weblens-input' }}
                            variant="unstyled"
                            type="password"
                            value={passInput}
                            placeholder="Password"
                            onChange={(event) =>
                                setPassInput(event.currentTarget.value)
                            }
                            style={{ width: '100%' }}
                        />
                        <Space h={'md'} />
                        <WeblensButton
                            label="Sign Up"
                            squareSize={50}
                            disabled={
                                userInput === '' ||
                                passInput === '' ||
                                badUsername
                            }
                            centerContent
                            onClick={() => CreateUser(userInput, passInput)}
                            setButtonRef={setButtonRef}
                            style={{ width: '100%' }}
                        />
                    </Tabs.Panel>
                </Tabs>
            </div>
        </div>
    )
}

export default Login
