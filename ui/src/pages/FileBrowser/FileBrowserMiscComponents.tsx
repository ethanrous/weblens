import { Text } from '@mantine/core'
import { ButtonIcon } from '@weblens/lib/buttonTypes'
import User from '@weblens/types/user/User'
import { friendlyFolderName } from '@weblens/util'

import fbStyle from './style/fileBrowserStyle.module.scss'

export const FileIcon = ({
    filename,
    id,
    Icon,
    usr,
    as,
}: {
    filename: string
    id: string
    Icon: ButtonIcon
    usr: User
    as?: string
}) => {
    return (
        <div className="flex items-center">
            <Icon className={fbStyle['icon-noshrink']} />
            <p className="font-medium text-white truncate text-nowrap p-2 shrink">
                {friendlyFolderName(filename, id, usr)}
            </p>
            {as && (
                <div className="flex flex-row items-center">
                    <Text size="12px">as</Text>
                    <Text
                        size="12px"
                        truncate="end"
                        style={{
                            fontFamily: 'monospace',
                            textWrap: 'nowrap',
                            padding: 3,
                            flexShrink: 2,
                        }}
                    >
                        {as}
                    </Text>
                </div>
            )}
        </div>
    )
}
