import { useEffect, useMemo, useState } from "react";
import * as d3 from "d3";
import { humanFileSize } from "../../util";
import { Box, Text } from "@mantine/core";
import { WeblensButton } from "../../components/WeblensButton";
import { useResize } from "../../components/hooks";
import { GetFileInfo, getFilesystemStats } from "../../api/FileBrowserApi";
import { useNavigate } from "react-router-dom";
import { AuthHeaderT, FileInfoT } from "../../types/Types";
import { IconFolder } from "@tabler/icons-react";

export type TreeNode = {
    type: "node";
    value: number;
    name: string;
    children: Tree[];
};
export type TreeLeaf = {
    type: "leaf";
    name: string;
    value: number;
};

type Tree = TreeNode | TreeLeaf;
type extSize = {
    name: string;
    value: number;
};

export const StatTree = ({
    folderInfo,
    authHeader,
}: {
    folderInfo: FileInfoT;
    authHeader: AuthHeaderT;
}) => {
    const nav = useNavigate();
    const [stats, setStats]: [stats: extSize[], setStats: (s) => void] =
        useState([]);
    const [boxRef, setBoxRef] = useState(null);
    const size = useResize(boxRef);
    const [statFilter, setStatsFilter] = useState([]);

    useEffect(() => {
        if (!folderInfo.id) {
            return;
        }
        getFilesystemStats(folderInfo.id, authHeader).then((s) =>
            setStats(s.sizesByExtension)
        );
    }, [folderInfo.id]);

    if (!folderInfo) {
        return null;
    }

    return (
        <Box
            style={{
                display: "flex",
                flexDirection: "column",
                justifyContent: "center",
                width: "100%",
                height: "100%",
            }}
        >
            <Box
                style={{
                    display: "flex",
                    flexDirection: "row",
                    alignItems: "center",
                    height: "58px",
                    width: "100%",
                    padding: 5,
                }}
            >
                <Box
                    style={{
                        display: "flex",
                        flexDirection: "row",
                        alignItems: "center",
                        cursor: "pointer",
                    }}
                    onClick={() => nav(`/files/${folderInfo.id}`)}
                >
                    <IconFolder
                        size={"28px"}
                        style={{ marginLeft: 10, marginRight: 4 }}
                    />
                    <Text className="crumb-text">{folderInfo.filename}</Text>
                </Box>
                <Box style={{ flexGrow: 1 }} />
                {stats.length !== 0 && (
                    <WeblensButton
                        onClick={() => setStatsFilter([])}
                        disabled={statFilter.length === 0}
                        width={150}
                        label={`Clear Filter`}
                        postScript={
                            statFilter.length === 0
                                ? "Right click to hide a block"
                                : `${statFilter.length} blocks hidden`
                        }
                    />
                )}
            </Box>
            {stats.length !== 0 && (
                <Box
                    ref={setBoxRef}
                    style={{
                        display: "flex",
                        backgroundColor: "#222222",
                        justifyContent: "center",
                        alignItems: "center",
                        height: "100%",
                        borderRadius: 4,
                        margin: 10,
                    }}
                >
                    <Tree
                        height={size.height}
                        width={size.width}
                        stats={stats.filter((s) => {
                            return !statFilter.includes(s.name);
                        })}
                        doSearch={(name) =>
                            nav(`/files/search/${folderInfo.id}?filter=${name}`)
                        }
                        statFilter={statFilter}
                        setStatsFilter={setStatsFilter}
                    />
                </Box>
            )}
        </Box>
    );
};

export const Tree = ({
    height,
    width,
    stats,
    doSearch,
    statFilter,
    setStatsFilter,
}: {
    height;
    width;
    stats: extSize[];
    doSearch: (name: string) => void;
    statFilter;
    setStatsFilter;
}) => {
    const [hovering, setHovering] = useState(null);
    const hierarchy = useMemo(() => {
        if (!stats) {
            return;
        }
        let total = 0;
        stats.map((s) => {
            total += s.value;
        });

        const data: Tree = {
            type: "node",
            name: "FileType",
            value: 0,
            children: stats
                .filter((k) => {
                    return k.value != 0 && (k.value / total) * 100 >= 1;
                })
                .map((k: extSize) => {
                    return {
                        type: "leaf",
                        name: k.name,
                        value: k.value,
                    };
                }),
        };
        return d3.hierarchy(data).sum((d) => d.value);
    }, [stats, statFilter]);

    const root = useMemo(() => {
        if (!hierarchy) {
            return;
        }
        const treeGenerator = d3
            .treemap<Tree>()
            .size([width, height])
            .padding(4);
        return treeGenerator(hierarchy);
    }, [hierarchy, width, height]);

    if (!hierarchy?.children) {
        return (
            <svg
                width={width}
                height={height}
                onContextMenu={(e) => e.preventDefault()}
            />
        );
    }

    const firstLevelGroups = hierarchy.children.map((child) => child.data.name);
    var colorScale = d3.scaleOrdinal().domain(firstLevelGroups).range([
        // '#ff4444',
        // '#44ff44',
        // '#4444ff',
        "#264653",
        "#2A9D8F",
        "#E9C46A",
        "#F4A261",
        "#E76F51",
        "#ea2f86",
        "#f09c0a",
        "#fae000",
        "#93e223",
        "#4070d3",
        "#493c9e",
    ]);

    const allShapes = root.leaves().map((leaf) => {
        const color = colorScale(leaf.data.name);
        const [size, units] = humanFileSize(leaf.data.value);
        const sizeStr = `${size}${units}`;
        const textWidth =
            leaf.data.name.length > sizeStr.length
                ? leaf.data.name.length
                : sizeStr.length;

        return (
            <g
                key={leaf.data.name}
                style={{
                    cursor: "pointer",
                    opacity:
                        hovering && hovering !== leaf.data.name
                            ? "20%"
                            : "100%",
                }}
                className="file-type-block"
                onContextMenu={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    setStatsFilter((p) => {
                        p.push(leaf.data.name);
                        return [...p];
                    });
                }}
                onMouseOver={() => setHovering(leaf.data.name)}
                onMouseLeave={() => setHovering(null)}
                onClick={() => doSearch(leaf.data.name)}
            >
                <rect
                    key={`${leaf.data.name}-rect`}
                    x={leaf.x0}
                    y={leaf.y0}
                    rx={2}
                    width={leaf.x1 - leaf.x0}
                    height={leaf.y1 - leaf.y0}
                    fill={color}
                />
                <rect
                    key={`${leaf.data.name}-text-back`}
                    x={leaf.x0 - (textWidth * 11) / 2 + (leaf.x1 - leaf.x0) / 2}
                    y={leaf.y0 - 25 + (leaf.y1 - leaf.y0) / 2}
                    width={textWidth * 11}
                    height={50}
                    rx={4}
                    fill="#00000044"
                />
                <text
                    key={`${leaf.data.name}-name`}
                    x={leaf.x0 + (leaf.x1 - leaf.x0) / 2}
                    y={leaf.y0 - 9 + (leaf.y1 - leaf.y0) / 2}
                    fontSize={16}
                    textAnchor="middle"
                    dominantBaseline="middle"
                    fill={"white"}
                    fontWeight={600}
                >
                    {leaf.data.name}
                </text>
                <text
                    key={`${leaf.data.name}-size`}
                    x={leaf.x0 + (leaf.x1 - leaf.x0) / 2}
                    y={leaf.y0 + 9 + (leaf.y1 - leaf.y0) / 2}
                    fontSize={16}
                    textAnchor="middle"
                    dominantBaseline="middle"
                    fill={"white"}
                >
                    {sizeStr}
                </text>
            </g>
        );
    });

    return (
        <svg width={width} height={height}>
            {allShapes}
        </svg>
    );
};
