import { Box, Button, Input, Sheet, Typography, useTheme } from '@mui/joy'
import { useContext, useEffect, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { login } from '../../api/ApiFetch'
import { userContext } from '../../Context'
import { notifications } from '@mantine/notifications'

function CheckCreds(username, password, setCookie, nav) {
    login(username, password)
        .then(res => { if (res.status == 401) { return Promise.reject("Incorrect username or password") } else { return res.json() } })
        .then(data => {
            console.log("Setting username cookie to ", username)
            setCookie('weblens-username', username, { sameSite: "strict" })
            console.log("Setting session token to ", data.token)
            setCookie('weblens-login-token', data.token, { sameSite: "strict" })
            nav("/")
        })
        .catch((r) => { notifications.show({ message: r, color: "red" }) })
}

const Login = () => {
    const [userInput, setUserInput] = useState("")
    const [passInput, setPassInput] = useState("")
    const nav = useNavigate()
    const loc = useLocation()
    const theme = useTheme()
    const { authHeader, setCookie } = useContext(userContext)

    useEffect(() => {
        if (loc.state == null && authHeader.Authorization != "") {
            nav("/")
        }
    }, [authHeader])

    return (
        <Box height={"100vh"} width={"100vw"} display={"flex"} justifyContent={"center"} alignItems={"center"}
            sx={{ background: "linear-gradient(45deg, rgba(2,0,36,1) 0%, rgba(94,43,173,1) 50%, rgba(0,212,255,1) 100%);" }}
        >
            <Sheet
                sx={{
                    display: "flex",
                    flexDirection: "column",
                    justifyContent: "center",
                    alignItems: "center",
                    padding: "20px",
                    bgcolor: `rgba(${theme.colorSchemes.dark.palette.primary.solidBg} / 0.5)`,
                    backdropFilter: "blur(8px)",
                    borderRadius: "8px"
                }}
            >
                <Typography color={'primary'} component={'h1'} fontSize={30} style={{ marginTop: -100, marginBottom: 50 }}>
                    Weblens
                </Typography>
                <Input
                    sx={{ backgroundColor: '#00000000' }}
                    placeholder='Username'
                    onChange={(e) => setUserInput(e.target.value)}
                />
                <Input
                    sx={{ backgroundColor: '#00000000' }}
                    placeholder='Password'
                    type="password"
                    onChange={(e) => setPassInput(e.target.value)}
                    onKeyDown={(e) => { if (e.key === 'Enter') { CheckCreds(userInput, passInput, setCookie, nav) } }}
                />
                <Button
                    sx={{ border: (theme) => `1px solid ${theme.palette.divider}`, ":hover": { backgroundColor: (theme) => theme.palette.primary.solidActiveBg } }}
                    style={{ height: "40px", width: "150px", margin: 10 }}
                    onClick={() => { CheckCreds(userInput, passInput, setCookie, nav) }}
                >
                    Log in
                </Button>
                <Typography textColor={'white'} > or </Typography>
                <Button
                    sx={{ border: (theme) => `1px solid ${theme.palette.divider}`, ":hover": { backgroundColor: (theme) => theme.palette.primary.solidActiveBg } }}
                    style={{ height: "40px", width: "150px", margin: 10 }}
                    onClick={() => nav("/signup")}
                >
                    Sign Up
                </Button>
            </Sheet>
        </Box>
    )
}

export default Login