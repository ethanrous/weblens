import { Box, Button, Checkbox, Button as ManButton, Paper, Space, Text, TextInput } from "@mantine/core"
import { useContext, useEffect, useMemo, useState } from "react"
import { userContext } from "../../Context"
import HeaderBar from "../../components/HeaderBar"
import { clearCache, adminCreateUser } from "../../api/ApiFetch"
import { ActivateUser, DeleteUser, GetUsersInfo } from "../../api/UserApi"
import { notifications } from "@mantine/notifications"
import { FlexColumnBox, FlexRowBox } from "../FileBrowser/FilebrowserStyles"

// const buttonSx = {
//     outline: "1px solid #444444",
//     fontSize: "16px",
//     margin: "10px",
//     padding: "20px",
//     "&:hover": {
//         outline: "0px",
//         boxShadow: "2px 2px 4px #222222",
//         background: "linear-gradient(45deg, rgba(2,0,36,1) -50%, rgba(94,43,173,1) 50%, rgba(0,212,255,1) 100%)",
//     }
// }

function CreateUserBox({ setAllUsersInfo, authHeader }) {
    const [userInput, setUserInput] = useState("")
    const [passInput, setPassInput] = useState("")
    const [makeAdmin, setMakeAdmin] = useState(false)
    return (
        <FlexColumnBox
            style={{ backgroundColor: "#333333", padding: "20px" }}
        >
            <TextInput
                value={userInput}
                placeholder="Username"
                style={{ margin: '8px' }}
                onChange={(e) => setUserInput(e.target.value)}
            />
            <TextInput
                value={passInput}
                placeholder="Password"
                style={{ margin: '8px' }}
                onChange={(e) => setPassInput(e.target.value)}
            />
            <Checkbox
                label='Admin'
                onChange={(e) => { setMakeAdmin(e.target.checked) }}
            />
            <Space h={20} />
            <Button color="#4444ff" onClick={() => adminCreateUser(userInput, passInput, makeAdmin, authHeader).then((_) => { GetUsersInfo(setAllUsersInfo, authHeader); setUserInput(""); setPassInput("") })}>
                Create User
            </Button>
        </FlexColumnBox>
    )
}

function UsersBox({ allUsersInfo, setAllUsersInfo, authHeader }) {
    const usersList = useMemo(() => {
        if (!allUsersInfo) {
            return null
        }
        const usersList = allUsersInfo.map((val) => {
            return (
                <FlexRowBox key={val.Username} style={{ width: '400px', justifyContent: 'space-between', alignItems: 'center', backgroundColor: '#4444ff', borderRadius: '6px', padding: 5, margin: 10 }}>
                    <FlexColumnBox>
                        <Text c={'white'}>
                            {val.Username}
                        </Text>
                        {val.Admin && (
                            <Text>Admin</Text>
                        )}
                    </FlexColumnBox>

                    {val.Activated == false && (
                        <ManButton onClick={() => { ActivateUser(val.Username, authHeader).then((_) => GetUsersInfo(setAllUsersInfo, authHeader)) }}>
                            Activate
                        </ManButton>
                    )}
                    <ManButton color="red" onClick={() => { DeleteUser(val.Username, authHeader).then((_) => GetUsersInfo(setAllUsersInfo, authHeader)) }}>
                        Delete
                    </ManButton>
                </FlexRowBox>
            )
        })
        return usersList
    }, [allUsersInfo])
    return (
        <FlexColumnBox
            style={{ padding: "10px", backgroundColor: "#333333" }}>
            <Text size={'20px'} fw={800} c={'white'} >
                Users
            </Text>
            <Space h={'10px'} />
            {usersList}
        </FlexColumnBox>
    )
}

const Admin = () => {
    const { authHeader } = useContext(userContext)
    const [allUsersInfo, setAllUsersInfo] = useState(null)

    useEffect(() => {
        if (authHeader.Authorization != "") {
            GetUsersInfo(setAllUsersInfo, authHeader)
        }
    }, [authHeader])



    return (
        <Box>
            <HeaderBar folderId={"home"} searchContent="" dispatch={() => { }} wsSend={() => { }} page={"admin"} searchRef={null} loading={false} progress={0} />
            <FlexColumnBox style={{ height: '100vh', justifyContent: 'center', alignItems: 'center' }}>
                <FlexRowBox>
                    <CreateUserBox setAllUsersInfo={setAllUsersInfo} authHeader={authHeader} />
                    <Space w={25} />
                    <UsersBox allUsersInfo={allUsersInfo} setAllUsersInfo={setAllUsersInfo} authHeader={authHeader} />
                </FlexRowBox>
                <Space h={50} />
                <ManButton color="red" onClick={() => { clearCache(authHeader).then((_) => notifications.show({ message: "Cache cleared" })) }}>
                    Clear Cache
                </ManButton>
            </FlexColumnBox >
        </Box>
    )
}

export default Admin