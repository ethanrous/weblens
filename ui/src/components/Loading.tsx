export default function WeblensLoader({
    size = 24,
    loading,
    className,
}: {
    size?: number
    loading?: string[]
    className?: string
}) {
    if (loading && loading.length === 0) {
        return null
    }
    return (
        <div
            className={
                'funky-spinner ' + className
            }
            style={{
                width: size,
                height: size,
            }}
            onClick={() => {
                console.log('Waiting on:', loading)
            }}
        />
    )
}
