import loaderStyle from './loaderDots.module.scss'

function LoaderDots({ className }: { className?: string }) {
    return <div className={loaderStyle.wlDotsLoader + ' ' + className} />
}

export default LoaderDots
