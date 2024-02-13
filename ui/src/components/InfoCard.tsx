import { Box, Text } from "@mantine/core";

export function InfoCard({ title, body }: { title: string, body: string }) {
    return (
        <Box style={{ position: 'absolute', height: '50vw', width: '50vw' }}>
            <Text>{title}</Text>
            <Text>{body}</Text>
        </Box>
    )
}