import { useNavigate } from 'react-router-dom';

import { Box, Text } from '@mantine/core';
import { IconChevronRight } from '@tabler/icons-react';

import { RowBox } from '../Pages/FileBrowser/FilebrowserStyles';
import { memo, useContext, useEffect, useMemo, useState } from 'react';
import {
    UserInfoT,
    FileInfoT,
    getBlankFile,
    UserContextT,
} from '../types/Types';
import { userContext } from '../Context';
import { useResize } from './hooks';

type breadcrumbProps = {
    label: string;
    onClick?: React.MouseEventHandler<HTMLDivElement>;
    onMouseUp?: () => void;
    dragging?: number;
    fontSize?: number;
    compact?: boolean;
    alwaysOn?: boolean;
    setMoveDest?;
};

export const StyledBreadcrumb = ({
    label,
    onClick,
    dragging,
    onMouseUp = () => {},
    alwaysOn = false,
    fontSize = 25,
    compact = false,
    setMoveDest,
}: breadcrumbProps) => {
    const [hovering, setHovering] = useState(false);
    let outline;
    let bgColor = 'transparent';

    if (alwaysOn) {
        outline = '1px solid #aaaaaa';
        bgColor = 'rgba(30, 30, 30, 0.5)';
    } else if (dragging === 1 && hovering) {
        outline = '2px solid #661199';
    } else if (dragging === 1) {
        bgColor = '#4444aa';
    }
    return (
        <Box
            className={compact ? 'crumb-box-compact' : 'crumb-box'}
            onMouseOver={() => {
                setHovering(true);
                if (dragging && setMoveDest) {
                    setMoveDest(label);
                }
            }}
            onMouseLeave={() => {
                setHovering(false);
                if (dragging && setMoveDest) {
                    setMoveDest('');
                }
            }}
            onMouseUp={(e) => {
                onMouseUp();
                setMoveDest('');
            }}
            onClick={onClick}
            style={{
                outline: outline,
                backgroundColor: bgColor,
                flexShrink: 1,
                minWidth: 0,
            }}
        >
            <Text
                className="crumb-text"
                truncate="end"
                style={{
                    fontSize: fontSize,
                    width: 'max-content',
                    maxWidth: '100%',
                }}
            >
                {label}
            </Text>
        </Box>
    );
};

// The crumb concatenator, the Crumbcatenator
const Crumbcatenator = ({ crumb, index, squished, setWidth }) => {
    const [crumbRef, setCrumbRef] = useState(null);
    const size = useResize(crumbRef);
    useEffect(() => {
        setWidth(index, size.width);
    }, [size]);

    if (squished >= index + 1) {
        return null;
    }

    return (
        <Box
            ref={setCrumbRef}
            style={{
                display: 'flex',
                alignItems: 'center',
                width: 'max-content',
            }}
        >
            {size.width > 10 && index !== 0 && (
                <IconChevronRight style={{ width: '20px', minWidth: '20px' }} />
            )}
            {crumb}
        </Box>
    );
};

export const StyledLoaf = ({ crumbs }) => {
    const [widths, setWidths] = useState(new Array(crumbs.length));
    const [squished, setSquished] = useState(0);
    const [crumbsRef, setCrumbRef] = useState(null);
    const size = useResize(crumbsRef);

    useEffect(() => {
        if (widths.length !== crumbs.length) {
            const newWidths = [...widths.slice(0, crumbs.length)];
            setWidths(newWidths);
        }
    }, [crumbs.length]);

    useEffect(() => {
        if (!widths || widths[0] == undefined) {
            return;
        }
        let total = widths.reduce((acc, v) => acc + v);
        let squishCount;

        // - 20 to account for width of ... text
        for (squishCount = 0; total > size.width - 20; squishCount++) {
            total -= widths[squishCount];
        }
        setSquished(squishCount);
    }, [size, widths]);

    return (
        <Box ref={setCrumbRef} className="loaf">
            {squished !== 0 && <Text className="crumb-text">...</Text>}

            {crumbs.map((c, i) => (
                <Crumbcatenator
                    key={i}
                    crumb={c}
                    index={i}
                    squished={squished}
                    setWidth={(index, width) =>
                        setWidths((p) => {
                            p[index] = width;
                            return [...p];
                        })
                    }
                />
            ))}
        </Box>
    );
};

const Crumbs = memo(
    ({
        finalFile,
        parents,
        moveSelectedTo,
        navOnLast,
        setMoveDest,
        dragging,
    }: {
        finalFile: FileInfoT;
        parents: FileInfoT[];
        navOnLast: boolean;
        moveSelectedTo?: (folderId: string) => void;
        setMoveDest?: (itemName: string) => void;
        dragging?: number;
    }) => {
        const navigate = useNavigate();
        const { usr }: UserContextT = useContext(userContext);

        const loaf = useMemo(() => {
            if (!usr || !finalFile?.id) {
                return null;
            }

            const parentsIds = parents.map((p) => p.id);
            if (parentsIds.includes('shared')) {
                let sharedRoot = getBlankFile();
                sharedRoot.filename = 'Shared';
                parents.unshift(sharedRoot);
            } else if (
                finalFile.id === usr.trashId ||
                parentsIds.includes(usr.trashId)
            ) {
                if (parents[0]?.id === usr.homeId) {
                    parents.shift();
                }
                if (
                    parents[0]?.id === usr.trashId &&
                    parents[0].filename !== 'Trash'
                ) {
                    parents[0].filename = 'Trash';
                }
                if (
                    finalFile.id === usr.trashId &&
                    finalFile.filename !== 'Trash'
                ) {
                    finalFile.filename = 'Trash';
                }
            } else if (
                finalFile.id === usr.homeId ||
                parentsIds.includes(usr.homeId)
            ) {
                if (
                    parents[0]?.id === usr.homeId &&
                    parents[0].filename !== 'Home'
                ) {
                    parents[0].filename = 'Home';
                }
                if (
                    finalFile.id === usr.homeId &&
                    finalFile.filename !== 'Home'
                ) {
                    finalFile.filename = 'Home';
                }
            } else if (finalFile.id == 'EXTERNAL_ROOT') {
                finalFile.filename = 'External';
                finalFile.id = 'external';
            } else if (parents[0]?.id == 'EXTERNAL_ROOT') {
                parents[0].filename = 'External';
                parents[0].id = 'external';
            }

            const crumbs = parents.map((parent) => {
                return (
                    <StyledBreadcrumb
                        key={parent.id}
                        label={parent.filename}
                        onClick={(e) => {
                            e.stopPropagation();
                            navigate(`/files/${parent.id}`);
                        }}
                        dragging={dragging}
                        onMouseUp={() => {
                            if (dragging !== 0) {
                                moveSelectedTo(parent.id);
                            }
                        }}
                        setMoveDest={setMoveDest}
                    />
                );
            });

            crumbs.push(
                <StyledBreadcrumb
                    key={finalFile.id}
                    label={
                        finalFile.id === usr.homeId
                            ? 'Home'
                            : finalFile.filename
                    }
                    onClick={(e) => {
                        e.stopPropagation();
                        if (!navOnLast) {
                            return;
                        }
                        navigate(
                            `/files/${
                                finalFile.parentFolderId === usr.homeId
                                    ? 'home'
                                    : finalFile.parentFolderId
                            }`
                        );
                    }}
                    setMoveDest={setMoveDest}
                />
            );

            return <StyledLoaf crumbs={crumbs} />;
        }, [
            parents,
            finalFile,
            moveSelectedTo,
            navOnLast,
            dragging,
            navigate,
            usr,
            setMoveDest,
        ]);

        return loaf;
    },
    (prev, next) => {
        return (
            (prev.dragging === next.dragging &&
                prev.parents === next.parents &&
                prev.finalFile === next.finalFile) ||
            next.parents === null ||
            !next.finalFile?.id
        );
    }
);

export default Crumbs;
