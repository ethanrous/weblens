import style from '@weblens/components/style.module.scss'

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
            <Logo className={style.fadeBlink} />
        </div>
    )
}
