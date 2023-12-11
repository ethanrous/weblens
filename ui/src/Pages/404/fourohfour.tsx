import { Box, Button, Typography } from "@mui/joy"
import { useNavigate } from "react-router-dom"

const Fourohfour = () => {
    const nav = useNavigate()
    return (
        <Box display={'flex'} flexDirection={'column'} height={'50vh'} alignItems={'center'} justifyContent={'center'}>
            <Typography color={'primary'} padding={'20px'}>
                Page not found :(
            </Typography>
            <Button onClick={() => nav('/')} color="primary">
                Go Home
            </Button>
        </Box>
    )
}
export default Fourohfour