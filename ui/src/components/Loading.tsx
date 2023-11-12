import { LinearProgress } from '@mui/joy'

export default function WeblensLoader({ loading, progress }: { loading: boolean, progress: number }) {
    let loader
    if (!loading) {
        return null
    }
    if (progress) {
        loader = (
            <LinearProgress determinate style={{ position: "absolute", width: "100%" }} variant="soft" value={Number(progress)} />
        )
    } else {
        loader = (
            <LinearProgress style={{ position: "absolute", width: "100%" }} variant="soft" />
        )
    }

    return loader

}