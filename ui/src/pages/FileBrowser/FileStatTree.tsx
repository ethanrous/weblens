import { IconFolder } from '@tabler/icons-react'
import { getFilesystemStats } from '@weblens/api/FileBrowserApi'
import { useResize } from '@weblens/components/hooks'
import WeblensButton from '@weblens/lib/WeblensButton'
import { WeblensFile } from '@weblens/types/files/File'
import { AuthHeaderT } from '@weblens/types/Types'
// import * as d3 from 'd3';
import { humanFileSize } from '@weblens/util'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'

export type TreeNode = {
    type: 'node'
    value: number
    name: string
    children: Tree[]
}
export type TreeLeaf = {
    type: 'leaf'
    name: string
    value: number
}

type Tree = TreeNode | TreeLeaf
type extSize = {
    name: string
    value: number
}

export const StatTree = ({
    folderInfo,
    authHeader,
}: {
    folderInfo: WeblensFile
    authHeader: AuthHeaderT
}) => {
    const nav = useNavigate()
    const [stats, setStats]: [stats: extSize[], setStats: (s) => void] =
        useState([])
    const [boxRef, setBoxRef] = useState(null)
    const size = useResize(boxRef)
    const [statFilter, setStatsFilter] = useState([])

    useEffect(() => {
        if (!folderInfo.Id()) {
            return
        }
        getFilesystemStats(folderInfo.Id(), authHeader).then((s) =>
            setStats(s.sizesByExtension)
        )
    }, [folderInfo.Id()])

    if (!folderInfo) {
        return null
    }

    return (
        <div className="flex flex-col justify-center w-full h-full">
            <div className="flex flex-row items-center justify-between w-full h-14 p-1">
                <div
                    className="flex flex-row items-center cursor-pointer"
                    onClick={() => nav(`/files/${folderInfo.Id()}`)}
                >
                    <IconFolder
                        size={'28px'}
                        style={{ marginLeft: 10, marginRight: 4 }}
                    />
                    <p className="crumb-text">{folderInfo.GetFilename()}</p>
                </div>

                <p className="flex w-max text-nowrap text-xl font-bold">
                    Folder content statistics
                </p>

                <WeblensButton
                    onClick={() => setStatsFilter([])}
                    disabled={statFilter.length === 0}
                    squareSize={50}
                    label={`Clear Filter`}
                />
            </div>
            <div
                ref={setBoxRef}
                className="flex justify-center items-center h-full rounded m-5"
            >
                {stats.length === 0 && (
                    <p className="flex w-max text-nowrap text-xl font-bold">
                        No content
                    </p>
                )}
                {stats.length !== 0 && (
                    <Tree
                        height={size.height}
                        width={size.width}
                        stats={stats.filter((s) => {
                            return !statFilter.includes(s.name)
                        })}
                        doSearch={(name) =>
                            nav(
                                `/files/search/${folderInfo.Id()}?filter=${name}`
                            )
                        }
                        statFilter={statFilter}
                        setStatsFilter={setStatsFilter}
                    />
                )}
            </div>
        </div>
    )
}

const Block = ({
    hovering,
    leaf,
    colorScale,
    setStatsFilter,
    doSearch,
    setHovering,
}: {
    hovering: boolean
    leaf
    colorScale
    setStatsFilter
    doSearch
    setHovering
}) => {
    if (!leaf) {
        return null
    }
    const color = colorScale(leaf.data.name)
    const [size, units] = humanFileSize(leaf.data.value)
    const sizeStr = `${size}${units}`
    const textWidth =
        leaf.data.name.length > sizeStr.length
            ? leaf.data.name.length
            : sizeStr.length

    const tooSmall = !(
        leaf.y1 - leaf.y0 > 50 && leaf.x1 - leaf.x0 > textWidth * 11
    )

    return (
        <g
            key={leaf.data.name}
            style={{
                cursor: 'pointer',
                opacity: hovering !== null && !hovering ? '20%' : '100%',
            }}
            className="file-type-block"
            onContextMenu={(e) => {
                e.preventDefault()
                e.stopPropagation()
                setHovering(null)
                setStatsFilter((p) => {
                    p.push(leaf.data.name)
                    return [...p]
                })
            }}
            onMouseOver={() => setHovering(leaf)}
            onMouseLeave={() => setHovering(null)}
            onClick={() => doSearch(leaf.data.name)}
            xmlns="http://www.w3.org/2000/svg"
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
                className="leaf-text-back"
                x={leaf.x0 - (textWidth * 11) / 2 + (leaf.x1 - leaf.x0) / 2}
                y={leaf.y0 - 25 + (leaf.y1 - leaf.y0) / 2}
                width={textWidth * 11}
                height={50}
                rx={4}
                fill={tooSmall && !hovering ? '#00000000' : '$dark-paper'}
            />
            <text
                className="leaf-text"
                key={`${leaf.data.name}-name`}
                x={leaf.x0 + (leaf.x1 - leaf.x0) / 2}
                y={leaf.y0 - 9 + (leaf.y1 - leaf.y0) / 2}
                fontSize={16}
                textAnchor="middle"
                dominantBaseline="middle"
                fill={'white'}
                fontWeight={600}
                opacity={tooSmall && !hovering ? '0%' : '100%'}
            >
                {leaf.data.name}
            </text>
            <text
                className="leaf-text"
                key={`${leaf.data.name}-size`}
                x={leaf.x0 + (leaf.x1 - leaf.x0) / 2}
                y={leaf.y0 + 9 + (leaf.y1 - leaf.y0) / 2}
                fontSize={16}
                textAnchor="middle"
                dominantBaseline="middle"
                fill={'white'}
                opacity={tooSmall && !hovering ? '0%' : '100%'}
            >
                {sizeStr}
            </text>
        </g>
    )
}

export const Tree = ({
    height,
    width,
    stats,
    doSearch,
    statFilter,
    setStatsFilter,
}: {
    height
    width
    stats: extSize[]
    doSearch: (name: string) => void
    statFilter
    setStatsFilter
}) => {
    return <p>broken right now</p>
    // const [hovering, setHovering] = useState(null);
    // const hierarchy = useMemo(() => {
    //     if (!stats) {
    //         return;
    //     }
    //     let total = 0;
    //     stats.map(s => {
    //         total += s.value;
    //     });
    //
    //     const data: Tree = {
    //         type: 'node',
    //         name: 'FileType',
    //         value: 0,
    //         children: stats
    //             .filter(k => {
    //                 return k.value != 0 && (k.value / total) * 100 >= 1;
    //             })
    //             .map((k: extSize) => {
    //                 return {
    //                     type: 'leaf',
    //                     name: k.name,
    //                     value: k.value,
    //                 };
    //             }),
    //     };
    //     // return d3.hierarchy(data).sum(d => d.value);
    // }, [stats, statFilter]);
    //
    // const root = useMemo(() => {
    //     if (!hierarchy) {
    //         return;
    //     }
    //     // const treeGenerator = d3.treemap<Tree>().size([width, height]).padding(4);
    //     // return treeGenerator(hierarchy);
    // }, [hierarchy, width, height]);
    //
    // if (!hierarchy?.children) {
    //     return <svg width={width} height={height} onContextMenu={e => e.preventDefault()} />;
    // }
    //
    // const firstLevelGroups = hierarchy.children.map(child => child.data.name);
    // const colorScale = d3
    //     .scaleOrdinal()
    //     .domain(firstLevelGroups)
    //     .range([
    //         '#264653',
    //         '#2A9D8F',
    //         '#E9C46A',
    //         '#F4A261',
    //         '#E76F51',
    //         '#ea2f86',
    //         '#f09c0a',
    //         '#fae000',
    //         '#93e223',
    //         '#4070d3',
    //         '#493c9e',
    //     ]);
    //
    // const leaves = root.leaves().sort((a, b) => {
    //     if (a === hovering) {
    //         return 1;
    //     } else if (b === hovering) {
    //         return -1;
    //     } else {
    //         return 0;
    //     }
    // });
    //
    // const allShapes = leaves.map(leaf => {
    //     return (
    //         <Block
    //             key={leaf.data.name}
    //             leaf={leaf}
    //             hovering={hovering === null ? null : hovering === leaf}
    //             colorScale={colorScale}
    //             setStatsFilter={setStatsFilter}
    //             doSearch={doSearch}
    //             setHovering={setHovering}
    //         />
    //     );
    // });
    //
    // return (
    //     <svg width={width} height={height} overflow={'visible'}>
    //         {allShapes}
    //     </svg>
    // );
}
