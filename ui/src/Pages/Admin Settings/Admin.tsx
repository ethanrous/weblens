import { Box, Button, Checkbox, Sheet, Input, Typography, FormControl, FormLabel, useTheme } from "@mui/joy"
import { Button as ManButton } from "@mantine/core"
import { useContext, useEffect, useMemo, useState } from "react"
import API_ENDPOINT from "../../api/ApiEndpoint"
import { userContext } from "../../Context"
import { useSnackbar } from "notistack"
import HeaderBar from "../../components/HeaderBar"
import { createUser, clearCache } from "../../api/ApiFetch"
import { ActivateUser, DeleteUser, GetUsersInfo } from "../../api/UserApi"
import { notifications } from "@mantine/notifications"

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
            GetUsersInfo(setAllUsersInfo, authHeader, enqueueSnackbar)
        }
    }, [authHeader])

    const usersList = useMemo(() => {
        if (!allUsersInfo) {
            return null
        }
        const usersList = allUsersInfo.map((val) => {
            return (
                <Box display={"flex"} alignItems={'center'} width={"400px"} justifyContent={"space-between"}>
                    <Typography variant="solid" key={val.Username}>
                        {val.Username} - Admin: {val.Admin.toString()}
                    </Typography>
                    {val.Activated == false && (
                        <ManButton onClick={() => { ActivateUser(val.Username, authHeader).then((_) => GetUsersInfo(setAllUsersInfo, authHeader, enqueueSnackbar)) }}>
                            Activate
                        </ManButton>
                    )}
                    <ManButton color="red" onClick={() => { DeleteUser(val.Username, authHeader).then((_) => GetUsersInfo(setAllUsersInfo, authHeader, enqueueSnackbar)) }}>
                        Delete
                    </ManButton>
                </Box>
            )
        })
        return usersList
    }, [allUsersInfo])

    return (
        <Box>
            <HeaderBar folderId={"home"} searchContent="" dispatch={() => { }} wsSend={() => { }} page={"admin"} searchRef={null} loading={false} progress={0} />
            <Box height={"100vh"} display={"flex"} flexDirection={"column"} justifyContent={"center"} alignItems={"center"} sx={{ backgroundImage: "linear-gradient(to bottom right, rgb(89,54,146), rgb(89,54,246))" }}>
                <Sheet
                    sx={{ display: "flex", flexDirection: "column", backgroundColor: "rgba(0, 0, 0, 0.5)", backdropFilter: "blur(10px)", justifyContent: "center", alignItems: "center", padding: "20px", backgroundImage: "linear-gradient(to bottom right, rgba(100,100,255,0.2), rgba(100,100,255,0.1))", boxShadow: "8px 8px 10px rgba(30,30,30,0.5)" }}
                >
                    <Input
                        value={userInput}
                        placeholder="Username"
                        sx={{ margin: '8px' }}
                        onChange={(e) => setUserInput(e.target.value)}
                    />
                    <Input
                        value={passInput}
                        placeholder="Password"
                        sx={{ margin: '8px' }}
                        onChange={(e) => setPassInput(e.target.value)}
                    />
                    <Checkbox
                        label='Admin'
                        sx={{ margin: '8px' }}
                        onChange={(e) => { setMakeAdmin(e.target.checked) }}
                    />
                    <Button sx={buttonSx} onClick={() => createUser(userInput, passInput, makeAdmin, authHeader, enqueueSnackbar).then((_) => { GetUsersInfo(setAllUsersInfo, authHeader, enqueueSnackbar); setUserInput(""); setPassInput("") })}>
                        Create User
                    </Button>
                </Sheet>
                <Sheet
                    sx={{ marginTop: "50px", padding: "20px", backgroundColor: "rgba(0, 0, 0, 0.5)", backdropFilter: "blur(10px)", backgroundImage: "linear-gradient(to bottom right, rgba(100,100,255,0.2), rgba(100,100,255,0.1))", boxShadow: "8px 8px 10px rgba(30,30,30,0.5)" }}>
                    <Typography variant="solid" >
                        Users:
                    </Typography>
                    {usersList}
                </Sheet>
                <ManButton color="red" onClick={() => { clearCache(authHeader).then((_) => notifications.show({ message: "Cache cleared" })) }}>
                    Clear Cache
                </ManButton>
            </Box >
        </Box>
    )
}

export default Admin