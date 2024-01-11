import { Loader, Progress } from "@mantine/core"


export default function WeblensLoader({ loading, progress }: { loading: boolean, progress: number }) {
    let loader
    if ((!loading) && (progress === 0 || progress === 100)) {
        return null
    }
    if (progress && progress !== 100) {
        loader = (
            <Progress color='#4444ff' style={{ position: "absolute", width: "100%" }} value={Number(progress)} />
        )
    } else {
        loader = (
            <Loader color='#4444ff' type='bars' style={{ position: "absolute", right: '2vh', top: '95vh' }} />
        )
    }
    return loader
}