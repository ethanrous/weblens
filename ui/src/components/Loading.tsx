import { LinearProgress } from '@mui/joy'

export default function WeblensLoader({ loading, progress }: { loading: boolean, progress: number }) {
    let loader
    if ((!loading) && (progress === 0 || progress === 100)) {
        return null
    }
    if (progress && progress !== 100) {
        loader = (
            <LinearProgress determinate style={{ position: "absolute", width: "100%" }} variant="plain" value={Number(progress)}
                sx={{
                    "--LinearProgress-thickness": "5px",
                    "--LinearProgress-radius": "0px"
                }}
            />
        )
    } else {
        loader = (
            <LinearProgress style={{ position: "absolute", width: "100%" }} variant="plain"
                sx={{
                    "--LinearProgress-thickness": "5px",
                    "--LinearProgress-radius": "0px"
                }}
            />
        )
    }
    return loader
}