import { LinearProgress } from '@mui/material'

export default function WeblensLoader({ loading, progress }: { loading: boolean, progress: number }) {
    let loader
    if (!loading) {
        return null
    }
    if (progress) {
        loader = (
            <LinearProgress style={{ position: "absolute", width: "100%" }} variant="determinate" value={progress} color='primary' />
        )
    } else {
        loader = (
            <LinearProgress style={{ position: "absolute", width: "100%" }} color='primary' />
        )
    }

    return loader

}