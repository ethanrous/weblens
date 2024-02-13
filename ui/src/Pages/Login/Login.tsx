import { useCallback, useContext, useEffect, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { createUser, login } from '../../api/ApiFetch'
import { userContext } from '../../Context'
import { notifications } from '@mantine/notifications'
import { Button, Fieldset, Loader, PasswordInput, Space, Tabs, TextInput } from '@mantine/core'
import { RowBox } from '../FileBrowser/FilebrowserStyles'

function CheckCreds(username, password, setCookie, nav) {
    login(username, password)
        .then(res => { if (res.status === 401) { return Promise.reject("Incorrect username or password") } else { return res.json() } })
        .then(data => {
            setCookie('weblens-username', username, { sameSite: "strict" })
            setCookie('weblens-login-token', data.token, { sameSite: "strict" })
            nav("/")
        })
        .catch((r) => { notifications.show({ message: String(r), color: "red" }) })
}

function CreateUser(username, password) {
    createUser(username, password)
        .then(x => { notifications.show({ message: "Account created! Once an administrator activates your account you may login" }) })
        .catch((reason) => { notifications.show({ message: `Failed to create new user: ${String(reason)}`, color: 'red' }) })
}

export const useKeyDown = (login) => {

    const onKeyDown = useCallback((event) => {
        if (event.key === "Enter") {
            login()
        }
    }, [login])

    useEffect(() => {
        document.addEventListener('keydown', onKeyDown)
        return () => {
            document.removeEventListener('keydown', onKeyDown)
        }
    }, [onKeyDown])
}

const Login = () => {
    const [userInput, setUserInput] = useState("")
    const [passInput, setPassInput] = useState("")
    const [loading, setLoading] = useState(false)
    const [tab, setTab] = useState("login")
    const nav = useNavigate()
    const loc = useLocation()
    const { authHeader, setCookie } = useContext(userContext)

    useEffect(() => {
        if (loc.state == null && authHeader.Authorization !== "") {
            nav("/")
        }
    }, [authHeader, loc.state, nav])

    const login = useCallback(() => { setLoading(true); CheckCreds(userInput, passInput, setCookie, nav) }, [userInput, passInput, setCookie, nav])
    useKeyDown(login)

    return (
        <RowBox style={{ height: '100vh', width: '100vw', justifyContent: 'center', background: "linear-gradient(45deg, rgba(2,0,36,1) 0%, rgba(94,43,173,1) 50%, rgba(0,212,255,1) 100%)" }}>
            <Tabs value={tab} onChange={setTab} variant="pills">
                <Tabs.List grow>
                    <Tabs.Tab value="login" >
                        Login
                    </Tabs.Tab>
                    <Tabs.Tab value="signup" >
                        Sign Up
                    </Tabs.Tab>
                </Tabs.List>
                <Tabs.Panel value="login">
                    <Fieldset>
                        <TextInput value={userInput} label='Username' placeholder='Username' onChange={(event) => setUserInput(event.currentTarget.value)} />
                        <PasswordInput value={passInput} label='Password' placeholder='Password' onChange={(event) => setPassInput(event.currentTarget.value)} />
                        <Space h={'md'} />
                        <Button disabled={loading} fullWidth onClick={login}>{loading ? <Loader color='white' size={25} /> : "Login"}</Button>
                    </Fieldset >
                </Tabs.Panel>
                <Tabs.Panel value="signup">
                    <Fieldset>
                        <TextInput value={userInput} label='Username' placeholder='Username' onChange={(event) => setUserInput(event.currentTarget.value)} />
                        <PasswordInput value={passInput} label='Password' placeholder='Password' onChange={(event) => setPassInput(event.currentTarget.value)} />
                        <Space h={'md'} />
                        <Button fullWidth onClick={() => CreateUser(userInput, passInput)}>Sign Up</Button>
                    </Fieldset >
                </Tabs.Panel>
            </Tabs>
        </RowBox>
    )
}

export default Login