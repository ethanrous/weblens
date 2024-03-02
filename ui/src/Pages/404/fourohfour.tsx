import { Button, Text } from '@mantine/core'
import { ColumnBox } from '../FileBrowser/FilebrowserStyles'

const Fourohfour = () => {
    // const nav = useNavigate()
    const nav = null
    return (
        <ColumnBox style={{ height: '50vh', justifyContent: 'center' }}>
            <Text style={{ padding: 20 }}>
                Page not found :(
            </Text>
            <Button onClick={() => nav('/')} color="primary">
                Go Home
            </Button>
        </ColumnBox>
    )
}
export default Fourohfour