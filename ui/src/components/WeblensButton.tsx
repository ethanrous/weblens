import { Box, Text } from "@mantine/core";
import { CSSProperties, useState } from "react";
import { ColumnBox, RowBox } from "../Pages/FileBrowser/FilebrowserStyles";

type buttonProps = {
    label: string
    postScript?: string
    toggleOn?: boolean
    Left?
    Right?
    onToggle?: (t: boolean) => void
    onClick?
    style?: CSSProperties
}

export function WeblensButton({ label, postScript, toggleOn = undefined, onToggle = t => { }, Left = null, Right = null, onClick, style }: buttonProps) {
    if (toggleOn === undefined) {
        return (
            <Box className='weblens-button' style={style} onClick={onClick}>
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