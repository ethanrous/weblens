import { Button, Space, Text } from "@mantine/core"
import { FlexColumnBox } from "../Pages/FileBrowser/FilebrowserStyles"
import { useNavigate } from "react-router-dom"

function NotFound({resourceType, link, setNotFound}: {resourceType: string, link: string, setNotFound: (b: boolean) => void}) {
    const nav = useNavigate()
    return (
        <FlexColumnBox style={{justifyContent: 'center'}}>
            <FlexColumnBox style={{height: 'max-content', width: 'max-content', padding: 50, marginBottom: '40vh', backgroundColor: '#ffffff22', borderRadius: 6}}>
                <Text c='white'>{`Could not find ${resourceType}`}</Text>
                <Space h={15}/>
                <Button fullWidth onClick={() => {setNotFound(false); nav(link)}} color="#4444ff">Go Back</Button>
            </FlexColumnBox>
        </FlexColumnBox>
    )
}

export default NotFound