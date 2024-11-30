import { style } from '@weblens/components/style'

import Logo from './Logo'

export default function WeblensLoader({ loading }: { loading?: string[] }) {
    if (loading && loading.length === 0) {
        return null
    }
    return (
        <div
            className="flex justify-center"
            onClick={() => {
                console.log('Waiting on:', loading)
            }}
        >
            <Logo className={style['fade-blink']} />
        </div>
    )
}
