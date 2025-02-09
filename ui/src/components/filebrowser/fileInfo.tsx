import { Divider, Text } from '@mantine/core'
import WeblensFile from '@weblens/types/files/File'
import { humanFileSize } from '@weblens/util'

function FileInfo({ file }: { file: WeblensFile }) {
    const [size, units] = humanFileSize(file.GetSize())
    return (
        <div className="flex flex-col w-max whitespace-nowrap justify-center max-w-full ml-1 gap-1 mb-2">
            <p className="text-3xl font-semibold max-w-full">
                {file.GetFilename()}
            </p>
            {file.IsFolder() && (
                <div className="flex flex-row h-max w-full items-center justify-center">
                    <p className="text-sm max-w-full">
                        {file.GetChildren().length} Item
                        {file.GetChildren().length !== 1 ? 's' : ''}
                    </p>
                    <Divider orientation="vertical" size={2} mx={10} />
                    <Text style={{ fontSize: '25px' }}>
                        {size}
                        {units}
                    </Text>
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
