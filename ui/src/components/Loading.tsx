import { Loader } from '@mantine/core'

export default function WeblensLoader({
    loading,
}: {
    loading?: string[]
}) {
    if (loading && loading.length === 0) {
        return null
    }
    return (
        <div
            className="flex cursor-pointer justify-center"
            onClick={() => {
                console.log('Waiting on:', loading)
            }}
        >
            <Loader color="#4444ff" type="bars" />
        </div>
    )
}
