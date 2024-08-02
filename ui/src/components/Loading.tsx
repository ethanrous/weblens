import { Loader, Progress, Text } from '@mantine/core';
import { useState } from 'react';

export default function WeblensLoader({ loading = [], progress = 0 }: { loading?: string[]; progress?: number }) {
    const [menuOpen, setMenuOpen] = useState(false);
    let loader;

    if ((!loading && !progress) || (loading.length === 0 && (progress === 0 || progress === 100))) {
        return null;
    }
    if (progress && progress !== 100) {
        loader = <Progress color="#4444ff" className="absolute w-full" value={Number(progress)} />;
    } else {
        loader = (
            <div
                className="flex cursor-pointer justify-center"
                onClick={() => {
                    setMenuOpen(!menuOpen);
                }}
            >
                <Loader color="#4444ff" type="bars" />
                {menuOpen && (
                    <div className="flex flex-col absolute rounded p-2 bg-[#222222] justify-center items-center bottom-[50px] w-max h-max">
                        <Text>Waiting For:</Text>
                        {loading.map(l => (
                            <Text key={l}>{l}</Text>
                        ))}
                    </div>
                )}
            </div>
        );
    }
    return loader;
}
