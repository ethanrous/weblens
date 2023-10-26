//document.cookie = "name="+username.value+";path=/" + ";expires="+expire.toUTCString();
import TextField from '@mui/material/TextField'
import { Box, Button, Typography } from '@mui/material'
import { useEffect, useState } from 'react'
import { useCookies } from 'react-cookie'
import { useNavigate } from 'react-router-dom'
import API_ENDPOINT from '../../api/ApiEndpoint'
import { EnqueueSnackbar, useSnackbar } from 'notistack'

function CheckCreds(username, password, setCookie, nav, snack: EnqueueSnackbar) {
    var url = new URL(`${API_ENDPOINT}/login`)

    let data = {
        username: username,
        password: password
    }

    fetch(url.toString(), { method: "POST", body: JSON.stringify(data) })
        .then(res => { if (res.status == 401) { return Promise.reject("Incorrect username or password") } else { return res.json() } })
        .then(data => {
            setCookie("weblens-username", username)
            setCookie("weblens-login-token", data.token)
            nav("/")
        })
        .catch((r) => { snack(r, { variant: "error" }) })
}

const Login = () => {
    const [userInput, setUserInput] = useState("")
    const [passInput, setPassInput] = useState("")

    const nav = useNavigate()
    const { enqueueSnackbar } = useSnackbar()

    const [cookies, setCookie, removeCookie] = useCookies(['weblens-username', 'weblens-login-token']);

    useEffect(() => {
        if (cookies['weblens-username'] && cookies['weblens-login-token']) {
            nav("/")
        }
    }, [])

    return (
        <Box color={'primary'} height={"100vh"} width={"100vw"} display={"flex"} flexDirection={"column"} justifyContent={"center"} alignItems={"center"}>
            <Typography color={'primary'} component={'h1'} fontSize={30} style={{ marginTop: -100, marginBottom: 50 }}>
                Weblens
            </Typography>
            <TextField
                InputProps={{ inputProps: { style: { color: '#fff' } } }}
                label="Username"
                style={{ margin: 10 }}
                onChange={(e) => setUserInput(e.target.value)}
            />
            <TextField
                label="Password"
                type="password"
                InputProps={{ inputProps: { style: { color: '#fff' } } }}
                style={{ margin: 10 }}
                onChange={(e) => setPassInput(e.target.value)}
            />
            <Button sx={{ border: (theme) => `1px solid ${theme.palette.divider}` }} style={{ height: "40px", width: "150px", margin: 10 }} onClick={() => { CheckCreds(userInput, passInput, setCookie, nav, enqueueSnackbar) }} > Log in </Button>
        </Box>
    )
}

export default Login