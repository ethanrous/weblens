import { Box, Button, Checkbox, ScrollArea, Space, Text, TextInput } from "@mantine/core"
import { useContext, useEffect, useMemo, useState } from "react"
import { userContext } from "../../Context"
import HeaderBar from "../../components/HeaderBar"
import { clearCache, adminCreateUser } from "../../api/ApiFetch"
import { ActivateUser, DeleteUser, GetUsersInfo } from "../../api/UserApi"
import { notifications } from "@mantine/notifications"
import { FlexColumnBox, FlexRowBox } from "../FileBrowser/FilebrowserStyles"

function CreateUserBox({ setAllUsersInfo, authHeader }) {
    const [userInput, setUserInput] = useState("")
    const [passInput, setPassInput] = useState("")
    const [makeAdmin, setMakeAdmin] = useState(false)
    return (
        <FlexColumnBox
            style={{ backgroundColor: "#333333", padding: "20px", height: 'max-content', width: '300px' }}
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
            <Button color="#4444ff" onClick={() => adminCreateUser(userInput, passInput, makeAdmin, authHeader).then(() => { GetUsersInfo(setAllUsersInfo, authHeader); setUserInput(""); setPassInput("") })}>
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
                <FlexRowBox key={val.Username} style={{ width: '95%', height: '50px', justifyContent: 'space-between', alignItems: 'center', backgroundColor: '#4444ff', borderRadius: '6px', padding: 5, margin: 10 }}>
                    <FlexColumnBox style={{justifyContent: 'center', width: 'max-content', paddingLeft: '20px'}}>
                        <Text c={'white'}>
                            {val.Username}
                        </Text>
                        {val.Admin && (
                            <Text>Admin</Text>
                        )}
                    </FlexColumnBox>

                    {val.Activated === false && (
                        <Button onClick={() => { ActivateUser(val.Username, authHeader).then(() => GetUsersInfo(setAllUsersInfo, authHeader)) }}>
                            Activate
                        </Button>
                    )}
                    <Button color="red" onClick={() => { DeleteUser(val.Username, authHeader).then(() => GetUsersInfo(setAllUsersInfo, authHeader)) }}>
                        Delete
                    </Button>
                </FlexRowBox>
            )
        })
        return usersList
    }, [allUsersInfo, authHeader, setAllUsersInfo])

    return (
        <FlexColumnBox
            style={{ padding: "10px", backgroundColor: "#333333", height: '350px', width: '450px' }}>
            <Text size={'20px'} fw={800} c={'white'} >
                Users
            </Text>
            <Space h={'10px'} />
            <ScrollArea w={'100%'} type="never" maw={'450px'}>
                {usersList}
            </ScrollArea>
        </FlexColumnBox>
    )
}

const Admin = () => {
    const { authHeader } = useContext(userContext)
    const [allUsersInfo, setAllUsersInfo] = useState(null)

    useEffect(() => {
        if (authHeader.Authorization !== "") {
            GetUsersInfo(setAllUsersInfo, authHeader)
        }
    }, [authHeader])

    return (
        <Box>
            <HeaderBar searchContent="" dispatch={() => { }} page={"admin"} searchRef={null} loading={false} progress={0} />
            <FlexColumnBox style={{ height: '100vh', justifyContent: 'center', alignItems: 'center' }}>
                <FlexRowBox style={{justifyContent: 'center', height: 'max-content'}}>
                    <CreateUserBox setAllUsersInfo={setAllUsersInfo} authHeader={authHeader} />
                    <Space w={25} />
                    <UsersBox allUsersInfo={allUsersInfo} setAllUsersInfo={setAllUsersInfo} authHeader={authHeader} />
                </FlexRowBox>
                <Space h={25} />
                <Button color="red" onClick={() => { clearCache(authHeader).then(() => notifications.show({ message: "Cache cleared" })) }} >
                    Clear Cache
                </Button>
            </FlexColumnBox >
        </Box>
    )
}

export default Admin