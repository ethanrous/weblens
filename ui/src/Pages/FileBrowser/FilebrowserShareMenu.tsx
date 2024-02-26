import { Box, Button, Space, Text } from "@mantine/core";
import { ShareInput } from "../../components/Share";
import { ShareFiles } from "../../api/FileBrowserApi";
import { notifications } from "@mantine/notifications";
import { useState } from "react";

export function ShareBox({ candidates, authHeader }: { candidates, authHeader }) {
    const [value, setValue] = useState([])

    return (
        <Box>
            <Text>Share Stuff</Text>
            <ShareInput valueSetCallback={setValue} initValues={[]} />
            <Space h={10} />
            {/* <Button fullWidth disabled={JSON.stringify(value) === JSON.stringify([])} color="#4444ff" onClick={() => { ShareFiles([], value, authHeader).then(() => { notifications.show({ message: "File(s) shared", color: 'green' }) }).catch((r) => notifications.show({ title: "Failed to share files", message: String(r), color: 'red' })) }}>
                Update
            </Button> */}

            <Text onClick={e => { e.stopPropagation(); ShareFiles(candidates, true, [], authHeader).then(v => navigator.clipboard.writeText(`${window.location.origin}/share/${v.shareId}`)) }}>
                Get public share link
            </Text>
        </Box>
    )
}