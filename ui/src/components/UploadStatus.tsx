import { Box, Card, Typography } from "@mui/joy"
import { useMemo, useState } from "react"

const UploadStatus = ({ fileStrings }: { fileStrings: Map<string, boolean> }) => {
    const files = useMemo(() => {
        const files = []
        for (const file of fileStrings) {
            files.push(
                <Typography variant="outlined">
                    {`${file}`}
                </Typography>
            )
        }
        return files
    }, [fileStrings.keys()])
    if (files.length === 0) {
        return null
    }
    return (
        <Card sx={{ position: 'fixed', bottom: 10, right: 10, height: "100px", width: "100px", overflow: 'hidden' }}>
            <Box position={'absolute'} bottom={10} right={10} zIndex={3} sx={{ height: "100%", width: "100%", background: "linear-gradient(180deg, rgba(2,0,36,1) 0%, rgba(0,101,255,0) 100%)" }} />
            <Box position={'absolute'} bottom={10}>
                {files}
            </Box>
        </Card>
    )
}

export default UploadStatus