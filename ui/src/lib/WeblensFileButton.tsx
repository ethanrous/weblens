import { IconUpload } from '@tabler/icons-react'
import { HandleUploadButton } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import { ErrorHandler } from '@weblens/types/Types'
import { useRef } from 'react'

import WeblensButton from './WeblensButton'
import { ButtonActionHandler, ButtonProps } from './buttonTypes'

export interface FileButtonProps {
    folderId: string
    shareId?: string
    // onChange: (files: File[]) => void

    buttonProps?: ButtonProps
    multiple?: boolean
    accept?: string
    name?: string
    form?: string
    resetRef?: React.Ref<() => void>
    capture?: string
    inputProps?: React.InputHTMLAttributes<HTMLInputElement>
}

function WeblensFileButton(props: FileButtonProps) {
    const inputRef = useRef<HTMLInputElement>(null)

    const onClick: ButtonActionHandler = () => {
        if (!props.buttonProps?.disabled) {
            inputRef.current?.click()
        }
    }

    const reset = () => {
        if (inputRef.current) {
            inputRef.current.value = ''
        }
    }

    const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        const files = Array.from(event.currentTarget.files)
        HandleUploadButton(files, props.folderId, false, props.shareId)
            .then(() => {
                reset()
            })
            .catch(ErrorHandler)
    }

    return (
        <>
            <WeblensButton
                onClick={onClick}
                Left={IconUpload}
                tooltip="Upload"
                {...props.buttonProps}
            />
            <input
                className="hidden"
                type="file"
                ref={inputRef}
                multiple={props.multiple}
                onChange={handleChange}
                {...props.inputProps}
            />
        </>
    )
}

export default WeblensFileButton
