//document.cookie = "name="+username.value+";path=/" + ";expires="+expire.toUTCString();
import TextField from '@mui/material/TextField'
import { Box, Button } from '@mui/material'
import { useEffect, useState } from 'react'
import { useCookies } from 'react-cookie'
import { useNavigate } from 'react-router-dom'
import API_ENDPOINT from '../../api/ApiEndpoint'

function CheckCreds(username, password, setCookie, nav) {
    var url = new URL(`${API_ENDPOINT}/login`)

    let data = {
        username: username,
        password: password
    }

    fetch(url.toString(), { method: "POST", body: JSON.stringify(data) }).then(res => res.json()).then(data => {
        console.log("HERE!", data.token)
        setCookie("weblens-login-token", data.token)
        nav("/")
    })
}

const Login = () => {
    const [userInput, setUserInput] = useState("")
    const [passInput, setPassInput] = useState("")

    const nav = useNavigate()

    const [cookies, setCookie, removeCookie] = useCookies(['weblens-login-token']);

    useEffect(() => {
        if (cookies['weblens-login-token']) {
            nav("/")
        }
    }, [])

    return (
        <Box height={"100vh"} width={"100vw"} display={"flex"} flexDirection={"column"} justifyContent={"center"} alignItems={"center"}>
            <h1 style={{ marginTop: -100, marginBottom: 50 }}>Weblens</h1>
            <TextField

                label="Username"
                style={{ margin: 10 }}
                onChange={(e) => setUserInput(e.target.value)}
            />
            <TextField

                label="Password"
                type="password"
                style={{ margin: 10 }}
                onChange={(e) => setPassInput(e.target.value)}
            />
            <Button variant="contained" style={{ height: "40px", width: "150px", margin: 10 }} onClick={() => { CheckCreds(userInput, passInput, setCookie, nav) }} > Log in </Button>
        </Box>
    )
}

export default Login