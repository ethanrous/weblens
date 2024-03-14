import { Space, Text } from "@mantine/core"
import { ColumnBox } from "../Pages/FileBrowser/FilebrowserStyles"
import { useNavigate } from "react-router-dom"
import { WeblensButton } from "./WeblensButton"
import { useContext } from "react"
import { userContext } from "../Context"

function NotFound({ resourceType, link, setNotFound }: { resourceType: string, link: string, setNotFound: (b: boolean) => void }) {
    const { userInfo } = useContext(userContext)
    const nav = useNavigate()
    return (
        <ColumnBox style={{ justifyContent: 'center' }}>
            <ColumnBox style={{ height: 'max-content', width: 'max-content', padding: 50, marginBottom: '40vh', backgroundColor: '#4444ff33', outline: "1px solid #222277", borderRadius: 6 }}>
                <Text fw={600} size="25px" c='white'>{`Could not find ${resourceType}`}</Text>
                <Space h={15} />

                <WeblensButton label={userInfo.username ? "Go Back" : "Login"} onClick={(e) => { setNotFound(false); nav(userInfo.username ? link : "/login") }} />

            </ColumnBox>
        </ColumnBox>
    )
}

export default NotFound