import WeblensButton from '@weblens/lib/WeblensButton'
import { useNavigate } from 'react-router-dom'

const Fourohfour = () => {
    const nav = useNavigate()
    return (
        <div className="h-[50vh] justify-center">
            <span>Page not found :(</span>
            <WeblensButton onClick={() => nav('/')} label="Go Home" />
        </div>
    )
}
export default Fourohfour
