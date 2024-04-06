import { Box, Dialog, Divider, Text } from '@mantine/core';
import { FileInfoT } from '../../types/Types';
import { useResizeDrag, useWindowSize } from '../../components/hooks';
import { useMemo, useState } from 'react';
import {
    IconFile,
    IconFolder,
    IconInfoCircle,
    IconX,
} from '@tabler/icons-react';
import { ColumnBox, RowBox } from './FilebrowserStyles';
import { clamp } from '../../util';

const SIDEBAR_BREAKPOINT = 650;

export default function FileInfoPane({
    selectedFiles,
}: {
    selectedFiles: FileInfoT[];
}) {
    const windowSize = useWindowSize();

    const [resizing, setResizing] = useState(false);
    const [resizeOffset, setResizeOffset] = useState(
        windowSize?.width > SIDEBAR_BREAKPOINT ? 450 : 75
    );
    useResizeDrag(
        resizing,
        setResizing,
        (v) => setResizeOffset(clamp(v, 200, 800)),
        true
    );
    const [open, setOpen] = useState(false);

    const titleText = useMemo(() => {
        if (selectedFiles.length === 0) {
            return 'No files selected';
        } else if (selectedFiles.length === 1) {
            return selectedFiles[0].filename;
        } else {
            return `${selectedFiles.length} files selected`;
        }
    }, [selectedFiles]);

    const singleItem = selectedFiles.length === 1;
    const itemIsFolder = selectedFiles[0]?.isDir;

    if (!open) {
        return (
            <Box className="file-info-pane" style={{ width: 48 }}>
                <IconInfoCircle
                    id="info-drawer-control"
                    className="file-info-icon"
                    style={{ marginTop: 18 }}
                    onClick={() => setOpen(true)}
                />
            </Box>
        );
    }

    return (
        <Box
            className="file-info-pane"
            mod={{ 'data-resizing': resizing.toString() }}
            style={{ width: resizeOffset }}
        >
            <Box
                draggable={false}
                className="resize-bar-wrapper"
                onMouseDown={(e) => {
                    e.preventDefault();
                    setResizing(true);
                }}
            >
                <Box className="resize-bar" />
            </Box>
            <Box className="file-info-content">
                <RowBox
                    style={{ height: '58px', justifyContent: 'space-between' }}
                >
                    <IconX
                        id="info-drawer-control"
                        className="file-info-icon"
                        onClick={() => setOpen(false)}
                    />
                    <Text
                        size="24px"
                        fw={600}
                        style={{ textWrap: 'nowrap', paddingRight: 32 }}
                    >
                        {titleText}
                    </Text>
                </RowBox>
                {selectedFiles.length > 0 && (
                    <ColumnBox style={{ height: 'max-content' }}>
                        <RowBox>
                            {singleItem && itemIsFolder && (
                                <IconFolder size={'48px'} />
                            )}
                            {(!singleItem || !itemIsFolder) && (
                                <IconFile size={'48px'} />
                            )}
                            <Text size="20px">
                                {itemIsFolder ? 'Folder' : 'File'}
                                {singleItem ? '' : 's'} Info
                            </Text>
                        </RowBox>
                        <Divider h={2} w={'90%'} m={10} />
                    </ColumnBox>
                )}
            </Box>
        </Box>
    );
}
