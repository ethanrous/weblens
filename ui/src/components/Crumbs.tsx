import { useNavigate } from "react-router-dom";

import { Box, Text } from "@mantine/core";
import { IconChevronRight, IconHome, IconTrash } from "@tabler/icons-react";

import { WeblensFile } from "../classes/File";
import { memo, useContext, useEffect, useMemo, useState } from "react";
import { UserContextT } from "../types/Types";
import { UserContext } from "../Context";
import { useResize } from "./hooks";

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

const CrumbText = ({ label, fontSize }) => {
    if (label === "Home") {
        return (
            <Box style={{ height: 30, width: 30 }}>
                <IconHome style={{ width: "100%", height: "100%" }} />
            </Box>
        );
    }
    if (label === "Trash") {
        return <IconTrash />;
    }

    return (
        <Text
            className="crumb-text"
            truncate="end"
            style={{
                fontSize: fontSize,
                width: "max-content",
                maxWidth: "100%",
            }}
        >
            {label}
        </Text>
    );
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
    let bgColor = "transparent";

    if (alwaysOn) {
        outline = "1px solid #aaaaaa";
        bgColor = "rgba(30, 30, 30, 0.5)";
    } else if (dragging === 1 && hovering) {
        outline = "2px solid #661199";
    } else if (dragging === 1) {
        bgColor = "#4444aa";
    }
    return (
        <Box
            className={compact ? "crumb-box-compact" : "crumb-box"}
            onMouseOver={() => {
                setHovering(true);
                if (dragging && setMoveDest) {
                    setMoveDest(label);
                }
            }}
            onMouseLeave={() => {
                setHovering(false);
                if (dragging && setMoveDest) {
                    setMoveDest("");
                }
            }}
            onMouseUp={(e) => {
                onMouseUp();
                setMoveDest("");
            }}
            onClick={onClick}
            style={{
                outline: outline,
                backgroundColor: bgColor,
                flexShrink: 1,
                minWidth: 0,
            }}
        >
            <CrumbText label={label} fontSize={fontSize} />
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
                display: "flex",
                alignItems: "center",
                width: "max-content",
            }}
        >
            {size.width > 10 && (
                <IconChevronRight style={{ width: "20px", minWidth: "20px" }} />
            )}
            {crumb}
        </Box>
    );
};

export const StyledLoaf = ({ crumbs, postText }) => {
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
            <Box
                ref={setCrumbRef}
                style={{
                    display: "flex",
                    alignItems: "center",
                    width: "max-content",
                }}
            >
                {crumbs[0]}
            </Box>

            {squished !== 0 && (
                <IconChevronRight style={{ width: "20px", minWidth: "20px" }} />
            )}
            {squished !== 0 && <Text className="crumb-text">...</Text>}

            {crumbs.slice(1).map((c, i) => (
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
            <Text
                className="crumb-text"
                truncate="end"
                style={{
                    marginLeft: 20,
                    color: "#c4c4c4",
                    fontSize: 20,
                }}
            >
                {postText}
            </Text>
        </Box>
    );
};

const Crumbs = memo(
    ({
        finalFile,
        postText,
        moveSelectedTo,
        navOnLast,
        setMoveDest,
        dragging,
    }: {
        finalFile: WeblensFile;
        postText?: string;
        navOnLast: boolean;
        moveSelectedTo?: (folderId: string) => void;
        setMoveDest?: (itemName: string) => void;
        dragging?: number;
    }) => {
        const navigate = useNavigate();
        const { usr }: UserContextT = useContext(UserContext);

        const loaf = useMemo(() => {
            if (!usr || !finalFile?.Id()) {
                return <StyledLoaf crumbs={[]} postText={""} />;
            }

            const parents = finalFile.FormatParents();
            const crumbs = parents.map((parent) => {
                return (
                    <StyledBreadcrumb
                        key={parent.Id()}
                        label={parent.GetFilename()}
                        onClick={(e) => {
                            e.stopPropagation();
                            navigate(`/files/${parent.Id()}`);
                        }}
                        dragging={dragging}
                        onMouseUp={() => {
                            if (dragging !== 0) {
                                moveSelectedTo(parent.Id());
                            }
                        }}
                        setMoveDest={setMoveDest}
                    />
                );
            });

            crumbs.push(
                <StyledBreadcrumb
                    key={finalFile.Id()}
                    label={finalFile.GetFilename()}
                    onClick={(e) => {
                        e.stopPropagation();
                        if (!navOnLast) {
                            return;
                        }
                        navigate(
                            `/files/${
                                finalFile.ParentId() === usr.homeId
                                    ? "home"
                                    : finalFile.ParentId()
                            }`
                        );
                    }}
                    setMoveDest={setMoveDest}
                />
            );

            return <StyledLoaf crumbs={crumbs} postText={postText} />;
        }, [
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
        if (
            prev.dragging !== next.dragging ||
            prev.finalFile !== next.finalFile
        ) {
            return false;
        }

        return true;
    }
);

export default Crumbs;
