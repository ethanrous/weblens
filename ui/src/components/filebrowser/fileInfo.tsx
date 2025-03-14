import WeblensFile from '@weblens/types/files/File'
import { humanFileSize } from '@weblens/util'

function FileInfo({ file }: { file: WeblensFile }) {
    const [size, units] = humanFileSize(file.GetSize())
    return (
        <div className="mb-2 ml-1 flex w-max max-w-full flex-col justify-center gap-1 whitespace-nowrap">
            <p className="max-w-full text-3xl font-semibold">
                {file.GetFilename()}
            </p>
            {file.IsFolder() && (
                <div className="flex h-max w-full flex-row items-center justify-center">
                    <p className="max-w-full text-sm">
                        {file.GetChildren().length} Item
                        {file.GetChildren().length !== 1 ? 's' : ''}
                    </p>
                    <span>
                        {size}
                        {units}
                    </span>
                </div>
            )}
            {!file.IsFolder() && (
                <p className={'text-sm'}>
                    {size}
                    {units}
                </p>
            )}
        </div>
    )
}

export default FileInfo
