import { Box, Loader, Progress, Text } from '@mantine/core';
import { useState } from 'react';

export default function WeblensLoader({
    loading = [],
    progress = 0,
}: {
    loading?: string[];
    progress?: number;
}) {
    const [menuOpen, setMenuOpen] = useState(false);
    let loader;

    if (
        (!loading && !progress) ||
        (loading.length === 0 && (progress === 0 || progress === 100))
    ) {
        return null;
    }
    if (progress && progress !== 100) {
        loader = (
            <Progress
                color="#4444ff"
                style={{ position: 'absolute', width: '100%' }}
                value={Number(progress)}
            />
        );
    } else {
        loader = (
            <Box
                onClick={() => {
                    setMenuOpen(!menuOpen);
                }}
                style={{
                    cursor: 'pointer',
                    justifyContent: 'center',
                    display: 'flex',
                }}
            >
                <Loader color="#4444ff" type="bars" />
                {menuOpen && (
                    <Box
                        style={{
                            padding: 10,
                            backgroundColor: '#222222',
                            borderRadius: 4,
                            position: 'absolute',
                            display: 'flex',
                            flexDirection: 'column',
                            justifyContent: 'center',
                            alignItems: 'center',
                            bottom: '50px',
                            width: 'max-content',
                            height: 'max-content',
                            // right: 2,
                        }}
                    >
                        <Text>Waiting For:</Text>
                        {loading.map((l) => (
                            <Text key={l}>{l}</Text>
                        ))}
                    </Box>
                )}
            </Box>
        );
    }
    return loader;
}
