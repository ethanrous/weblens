import { Box, Space, Text } from '@mantine/core';
import { FixedSizeList as ScrollList } from 'react-window';
import { FileInfoT } from '../../types/Types';
import { useState } from 'react';
import { IconDisplay } from './FilebrowserStyles';
import { useNavigate } from 'react-router-dom';
import { humanFileSize } from '../../util';

function FileRow({
    data,
    index,
    style,
}: {
    data: FileInfoT[];
    index: number;
    style;
}) {
    const nav = useNavigate();
    const [sz, un] = humanFileSize(data[index].size);

    return (
        <Box
            className="file-row"
            style={{
                ...style,
                borderBottom:
                    index === data.length - 1 ? '1px solid #555555' : '',
            }}
            onDoubleClick={() => {
                if (data[index].isDir) {
                    nav(`/files/${data[index].id}`);
                } else {
                    nav(
                        `/files/${data[index].parentFolderId}?jumpTo=${data[index].id}`
                    );
                }
            }}
        >
            <Box style={{ width: 24, height: 24 }}>
                <IconDisplay file={data[index]} />
            </Box>
            {/* <IconFolder /> */}
            <Text>{data[index].filename}</Text>
            <Space flex={1} />
            <Text>
                {sz}
                {un}
            </Text>
        </Box>
    );
}

export function FileRows({ files }: { files: FileInfoT[] }) {
    const [boxRef, setBoxRef] = useState(null);
    return (
        <Box ref={setBoxRef} className="file-rows-box">
            <ScrollList
                className="no-scrollbars"
                height={boxRef?.clientHeight ? boxRef.clientHeight : 0}
                width={boxRef?.clientWidth ? boxRef.clientWidth : 0}
                itemSize={48}
                itemCount={files.length}
                itemData={files}
                overscanRowCount={50}
            >
                {FileRow}
            </ScrollList>
        </Box>
    );
}
