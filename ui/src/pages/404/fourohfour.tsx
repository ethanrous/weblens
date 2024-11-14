import { Button, Text } from '@mantine/core'
import { useNavigate } from 'react-router-dom'

const Fourohfour = () => {
    const nav = useNavigate()
    return (
        <div style={{ height: '50vh', justifyContent: 'center' }}>
            <Text style={{ padding: 20 }}>Page not found :(</Text>
            <Button onClick={() => nav('/')} color="primary">
                Go Home
            </Button>
        </div>
    )
}
export default Fourohfour
