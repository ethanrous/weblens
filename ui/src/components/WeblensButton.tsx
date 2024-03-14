import { Box, Text } from "@mantine/core";
import { CSSProperties, useState } from "react";
import { ColumnBox, RowBox } from "../Pages/FileBrowser/FilebrowserStyles";
import { IconCheck } from "@tabler/icons-react";

type buttonProps = {
    label: string
    postScript?: string
    showSuccess?: boolean
    toggleOn?: boolean
    Left?
    Right?
    onToggle?: (t: boolean) => void
    onClick?: (e) => void | boolean
    style?: CSSProperties
}

export function WeblensButton({ label, postScript, showSuccess, toggleOn = undefined, onToggle = t => { }, Left = null, Right = null, onClick, style }: buttonProps) {
    const [success, setSuccess] = useState(false)
    if (toggleOn === undefined) {
        return (
            <Box className={`weblens-button${success && showSuccess ? '-success' : ''}`} style={style} onClick={e => { e.stopPropagation(); if (onClick(e) && showSuccess) { setSuccess(true); setTimeout(() => setSuccess(false), 1000) } }}>
                {(!success || !showSuccess) && (
                    <ColumnBox style={{ width: 'max-content' }}>
                        <RowBox style={{ justifyContent: 'space-evenly' }}>
                            {Left}
                            <Text fw={'inherit'} style={{ padding: 2 }}>{label}</Text>
                            {Right}
                        </RowBox>
                        {postScript && (
                            <Text fw={300} size="10px" style={{ padding: 2 }}>{postScript}</Text>
                        )}
                    </ColumnBox>
                )}
                {success && showSuccess && (
                    <ColumnBox style={{ width: 'max-content', height: '28px' }}>
                        <IconCheck />
                    </ColumnBox>
                )}
            </Box>
        )
    } else {
        return (
            <Box className={`weblens-toggle-button-${toggleOn ? "on" : "off"}`} style={style} onClick={() => onToggle(!toggleOn)}>
                <ColumnBox style={{ width: 'max-content' }}>
                    <RowBox style={{ justifyContent: 'space-evenly' }}>
                        {Left}
                        <Text fw={'inherit'} style={{ padding: 2 }}>{label}</Text>
                        {Right}
                    </RowBox>
                    {postScript && (
                        <Text fw={300} size="10px" style={{ padding: 2 }}>{postScript}</Text>
                    )}
                </ColumnBox>
            </Box>
        )
    }

}