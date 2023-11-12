import { Box, Button, Checkbox, Sheet, Input, Typography, FormControl, FormLabel, useTheme } from "@mui/joy"
import { useContext, useEffect, useMemo, useState } from "react"
import API_ENDPOINT from "../../api/ApiEndpoint"
import { userContext } from "../../Context"
import { useSnackbar } from "notistack"
import HeaderBar from "../../components/HeaderBar"
import { createUser } from "../../api/ApiFetch"

const buttonSx = {
    outline: "1px solid #444444",
    fontSize: "16px",
    margin: "10px",
    padding: "20px",
    "&:hover": {
        outline: "0px",
        boxShadow: "2px 2px 4px #222222",
        background: "linear-gradient(45deg, rgba(2,0,36,1) -50%, rgba(94,43,173,1) 50%, rgba(0,212,255,1) 100%)",
    }
}

const getUsersInfo = (setAllUsersInfo, authHeader, enqueueSnackbar) => {
    const url = new URL(`${API_ENDPOINT}/users`)
    fetch(url, { headers: authHeader, method: "GET" })
        .then(res => { if (res.status != 200) { return Promise.reject(`Could not get user info list: ${res.statusText}`) } else { return res.json() } })
        .then(data => setAllUsersInfo(data))
        .catch(r => enqueueSnackbar(r, { variant: "error" }))
}

const Admin = () => {
    const [userInput, setUserInput] = useState("")
    const [passInput, setPassInput] = useState("")
    const [makeAdmin, setMakeAdmin] = useState(false)
    const { authHeader, userInfo } = useContext(userContext)
    const [allUsersInfo, setAllUsersInfo] = useState(null)
    const theme = useTheme()
    const { enqueueSnackbar } = useSnackbar()

    useEffect(() => {
        if (authHeader.Authorization != "") {
            getUsersInfo(setAllUsersInfo, authHeader, enqueueSnackbar)
        }
    }, [authHeader])

    const usersList = useMemo(() => {
        if (!allUsersInfo) {
            return null
        }
        const usersList = allUsersInfo.map((val) => {
            return (
                <Typography color={'primary'} key={val.Username}>
                    {val.Username} - Admin: {val.Admin.toString()}
                </Typography>
            )
        })
        return usersList
    }, [allUsersInfo])

    return (
        <Box>
            <HeaderBar path={"/"} dispatch={() => { }} wsSend={() => { }} page={"admin"} searchRef={null} loading={false} progress={0} />
            <Box height={"100vh"} display={"flex"} flexDirection={"column"} justifyContent={"center"} alignItems={"center"} sx={{ backgroundImage: "linear-gradient(to bottom right, rgb(89,54,146), rgb(89,54,246))" }}>
                <Sheet
                    sx={{ display: "flex", flexDirection: "column", backgroundColor: "rgba(0, 0, 0, 0.5)", backdropFilter: "blur(10px)", justifyContent: "center", alignItems: "center", padding: "20px", backgroundImage: "linear-gradient(to bottom right, rgba(100,100,255,0.2), rgba(100,100,255,0.1))", boxShadow: "8px 8px 10px rgba(30,30,30,0.5)" }}
                >
                    <Input
                        placeholder="Username"
                        sx={{ margin: '8px' }}
                        onChange={(e) => setUserInput(e.target.value)}
                    />
                    <Input
                        placeholder="Password"
                        sx={{ margin: '8px' }}
                        onChange={(e) => setPassInput(e.target.value)}
                    />
                    <Checkbox
                        label='Admin'
                        sx={{ margin: '8px' }}
                        onChange={(e) => { setMakeAdmin(e.target.checked) }}
                    />
                    <Button sx={buttonSx} onClick={() => createUser(userInput, passInput, makeAdmin, authHeader, enqueueSnackbar)}>
                        Create User
                    </Button>
                </Sheet>
                <Sheet
                    sx={{ marginTop: "50px", padding: "20px", backgroundColor: "rgba(0, 0, 0, 0.5)", backdropFilter: "blur(10px)", backgroundImage: "linear-gradient(to bottom right, rgba(100,100,255,0.2), rgba(100,100,255,0.1))", boxShadow: "8px 8px 10px rgba(30,30,30,0.5)" }}>
                    <Typography color={'primary'} >
                        Users:
                    </Typography>
                    {usersList}
                </Sheet>
            </Box >
        </Box>
    )
}

export default Admin