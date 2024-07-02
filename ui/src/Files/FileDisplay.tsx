import { Divider, Loader, Text } from '@mantine/core';
import { IconFolder, IconPhoto } from '@tabler/icons-react';
import React, { memo, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { MediaImage } from '../Media/PhotoContainer';
import { FbMenuModeT, GlobalContextType, SelectedState, WeblensFile } from './File';
import { DraggingStateT, FbContext } from './filesContext';
import { useMedia } from '../components/hooks';
import { MoveSelected } from '../Pages/FileBrowser/FileBrowserLogic';
import { UserContext } from '../Context';
import { all } from 'axios';
import WeblensInput from '../components/WeblensInput';

type WrapperProps = {
    itemInfo: WeblensFile;
    fileRef;

    editing: boolean;

    selected: SelectedState;

    width: number;

    dragging: DraggingStateT;

    setSelected: (itemId: string, selected?: boolean) => void;
    doSelectMany: () => void;
    moveSelected: (entryId: string) => void;
    setMoveDest: (itemName: string) => void;

    setDragging: (d: DraggingStateT) => void;
    setHovering: (i: string) => void;

    setMenuMode: (m: FbMenuModeT) => void;
    setMenuPos: ({ x, y }: { x: number; y: number }) => void;
    setMenuTarget: (itemId: string) => void;

    children;
};

type TitleProps = {
    itemId: string;
    itemTitle: string;
    secondaryInfo?: string;
    editing: boolean;
    setEditing: (e: boolean) => void;
    allowEditing: boolean;
    blockFocus: (b: boolean) => void;
    rename: (itemId: string, newName: string) => void;
};

function selectedStyles(selected: SelectedState): {
    backgroundColor: string;
    outline: string;
} {
    let backgroundColor = '#222222';
    let outline;
    if (selected & SelectedState.Hovering) {
        backgroundColor = '#333333';
    }

    if (selected & SelectedState.InRange) {
        backgroundColor = '#373365';
    }

    if (selected & SelectedState.Selected) {
        backgroundColor = '#331177bb';
    }

    if (selected & SelectedState.LastSelected) {
        outline = '2px solid #442299';
    }

    if (selected & SelectedState.Droppable) {
        // $dark-paper
        backgroundColor = '#1c1049';
        outline = '2px solid #4444ff';
    }

    return { backgroundColor, outline };
}

const ItemWrapper = memo(
    ({
        itemInfo: file,
        fileRef,
        width,
        selected,
        setSelected,
        doSelectMany,
        dragging = DraggingStateT.NoDrag,
        setDragging,
        setHovering,
        moveSelected,
        setMenuMode,
        setMenuPos,
        setMenuTarget,
        setMoveDest,
        children,
    }: WrapperProps) => {
        const [mouseDown, setMouseDown] = useState(null);
        const nav = useNavigate();
        const { fbState, fbDispatch } = useContext(FbContext);

        const { outline, backgroundColor } = useMemo(() => {
            return selectedStyles(selected);
        }, [selected]);

        return (
            <div
                className="weblens-file animate-fade"
                ref={fileRef}
                onMouseOver={e => {
                    e.stopPropagation();
                    file.SetHovering(true);
                    setHovering(file.Id());
                    if (dragging && !file.IsSelected() && file.IsFolder()) {
                        setMoveDest(file.GetFilename());
                    }
                }}
                onMouseDown={e => {
                    setMouseDown({ x: e.clientX, y: e.clientY });
                }}
                onMouseMove={e => {
                    if (
                        mouseDown &&
                        !dragging &&
                        (Math.abs(mouseDown.x - e.clientX) > 20 || Math.abs(mouseDown.y - e.clientY) > 20)
                    ) {
                        setSelected(file.Id(), true);
                        setDragging(DraggingStateT.InternalDrag);
                    }
                }}
                onClick={e => {
                    e.stopPropagation();
                    if (e.shiftKey) {
                        doSelectMany();
                    } else {
                        setSelected(file.Id());
                    }
                }}
                onMouseUp={e => {
                    if (dragging !== 0) {
                        if (!file.IsSelected() && file.IsFolder()) {
                            moveSelected(file.Id());
                        }
                        setMoveDest('');
                        setDragging(DraggingStateT.NoDrag);
                    }
                    setMouseDown(null);
                }}
                onDoubleClick={e => {
                    e.stopPropagation();
                    const jump = file.GetVisitRoute(fbState.fbMode, fbState.shareId, fbDispatch);
                    if (jump) {
                        nav(jump);
                    }
                }}
                onContextMenu={e => {
                    e.preventDefault();
                    e.stopPropagation();

                    setMenuTarget(file.Id());
                    setMenuPos({ x: e.clientX, y: e.clientY });
                    if (fbState.menuMode === FbMenuModeT.Closed) {
                        setMenuMode(FbMenuModeT.Default);
                    }
                }}
                onMouseLeave={e => {
                    file.SetHovering(false);
                    setHovering('');
                    if (dragging && file.IsFolder()) {
                        setMoveDest('');
                    }
                    if (mouseDown) {
                        setMouseDown(null);
                    }
                }}
            >
                <div
                    className="flex flex-col items-center justify-center overflow-hidden rounded-md transition-colors"
                    children={children}
                    style={{
                        outline: outline,
                        backgroundColor: backgroundColor,
                        // height: (width - MARGIN * 2) * 1.1,
                        // width: width - MARGIN * 2,
                        cursor: dragging !== 0 && !file.IsFolder() ? 'default' : 'pointer',
                    }}
                />
                {(file.IsSelected() || !file.IsFolder()) && dragging !== 0 && (
                    <div
                        className="no-drop-cover"
                        // style={{
                        //     height: (width - MARGIN * 2) * 1.1,
                        //     width: width - MARGIN * 2,
                        // }}
                        onMouseLeave={e => {
                            file.SetHovering(false);
                            setHovering('');
                        }}
                        onClick={e => e.stopPropagation()}
                    />
                )}
            </div>
        );
    },
    (prev, next) => {
        if (prev.itemInfo !== next.itemInfo) {
            return false;
        } else if (prev.selected !== next.selected) {
            return false;
        } else if (prev.editing !== next.editing) {
            return false;
        } else if (prev.dragging !== next.dragging) {
            return false;
        } else if (prev.width !== next.width) {
            return false;
        } else if (prev.children !== next.children) {
            return false;
        }
        return true;
    },
);

const FileVisualWrapper = ({ file }) => {
    return (
        <div className="w-full p-2 pb-0 aspect-square overflow-hidden">
            <div className="w-full h-full overflow-hidden rounded-md flex justify-center items-center">
                <FileVisual file={file} />
            </div>
        </div>
    );
};

const FileVisual = memo(
    ({ file }: { file: WeblensFile }) => {
        const mediaData = useMedia(file.GetMediaId());

        if (file.IsFolder()) {
            return <IconFolder size={150} />;
        }

        if (mediaData) {
            return <MediaImage media={mediaData} quality="thumbnail" />;
        } else if (file.IsImage()) {
            return <IconPhoto />;
        }

        return null;
    },
    (prev, next) => {
        if (prev.file.GetMediaId() !== next.file.GetMediaId()) {
            return false;
        }
        return true;
    },
);

const useKeyDown = (
    itemId: string,
    oldName: string,
    newName: string,
    editing: boolean,
    setEditing: (b: boolean) => void,
    rename: (itemId: string, newName: string) => void,
) => {
    const onKeyDown = useCallback(
        event => {
            if (!editing) {
                return;
            }
            if (event.key === 'Enter') {
                if (oldName !== newName) {
                    rename(itemId, newName);
                }
                setEditing(false);
            } else if (event.key === 'Escape') {
                setEditing(false);
                // Rename with empty name is a "cancel" to the rename
                rename(itemId, '');
            }
        },
        [itemId, oldName, newName, editing, setEditing, rename],
    );

    useEffect(() => {
        document.addEventListener('keydown', onKeyDown);
        return () => {
            document.removeEventListener('keydown', onKeyDown);
        };
    }, [onKeyDown]);
};

const TextBox = memo(
    ({ itemId, itemTitle, secondaryInfo, editing, setEditing, allowEditing, blockFocus, rename }: TitleProps) => {
        const editRef: React.Ref<HTMLInputElement> = useRef();
        const [renameVal, setRenameVal] = useState(itemTitle);

        const setEditingPlus = useCallback(
            (b: boolean) => {
                setEditing(b);
                setRenameVal(cur => {
                    if (cur === '') {
                        return itemTitle;
                    } else {
                        return cur;
                    }
                });
                blockFocus(b);
            },
            [itemTitle, setEditing, blockFocus],
        );
        // useKeyDown(itemId, itemTitle, renameVal, editing, setEditingPlus, rename);

        return (
            <div className="item-info-box">
                <WeblensInput
                    subtle
                    value={renameVal}
                    valueCallback={setRenameVal}
                    openInput={() => {
                        setEditingPlus(true);
                    }}
                    closeInput={() => {
                        setEditingPlus(false);
                        rename(itemId, '');
                    }}
                    onComplete={v => {
                        rename(itemId, v);
                        setEditingPlus(false);
                    }}
                />
            </div>
        );
    },
    (prev, next) => {
        if (prev.secondaryInfo !== next.secondaryInfo) {
            return false;
        } else if (prev.editing !== next.editing) {
            return false;
        } else if (prev.itemTitle !== next.itemTitle) {
            return false;
        }
        return true;
    },
);

export const FileDisplay = memo(
    ({
        file,
        selected,
        index,
        context,
    }: {
        file: WeblensFile;
        selected: SelectedState;
        index: number;
        context: GlobalContextType;
    }) => {
        const wrapRef = useRef();
        const [editing, setEditing] = useState(false);
        const { fbState, fbDispatch } = useContext(FbContext);
        const [mouseDown, setMouseDown] = useState(null);
        const { authHeader } = useContext(UserContext);
        const nav = useNavigate();

        return (
            <div
                className="weblens-file animate-fade"
                ref={wrapRef}
                data-clickable={!fbState.draggingState || file.IsFolder()}
                data-selected={(selected & SelectedState.Selected) >> 8}
                data-in-range={(selected & SelectedState.InRange) >> 4}
                data-hovering={selected & SelectedState.Hovering}
                data-last-selected={(selected & SelectedState.LastSelected) >> 12}
                data-droppable={selected & SelectedState.Droppable}
                onMouseOver={e => {
                    e.stopPropagation();
                    file.SetHovering(true);
                    fbDispatch({ type: 'set_hovering', hovering: file.Id() });
                    if (fbState.draggingState && !file.IsSelected() && file.IsFolder()) {
                        fbDispatch({ type: 'set_move_dest', fileName: file.GetFilename() });
                    }
                }}
                onMouseDown={e => {
                    setMouseDown({ x: e.clientX, y: e.clientY });
                }}
                onMouseMove={e => {
                    if (
                        mouseDown &&
                        !fbState.draggingState &&
                        (Math.abs(mouseDown.x - e.clientX) > 20 || Math.abs(mouseDown.y - e.clientY) > 20)
                    ) {
                        fbDispatch({ type: 'set_selected', fileId: file.Id(), selected: true });
                        fbDispatch({ type: 'set_dragging', dragging: DraggingStateT.InternalDrag });
                    }
                }}
                onClick={e => {
                    e.stopPropagation();
                    if (e.shiftKey) {
                        fbDispatch({ type: 'set_selected', fileIds: context.doSelectMany() });
                    } else {
                        fbDispatch({ type: 'set_selected', fileId: file.Id() });
                    }
                }}
                onMouseUp={e => {
                    if (fbState.draggingState !== DraggingStateT.NoDrag) {
                        if (!file.IsSelected() && file.IsFolder()) {
                            MoveSelected(fbState.selected, file.Id(), authHeader).then(() =>
                                fbDispatch({ type: 'clear_selected' }),
                            );
                        }
                        fbDispatch({ type: 'set_move_dest', fileName: '' });
                        fbDispatch({ type: 'set_dragging', dragging: DraggingStateT.NoDrag });
                    }
                    setMouseDown(null);
                }}
                onDoubleClick={e => {
                    e.stopPropagation();
                    const jump = file.GetVisitRoute(fbState.fbMode, fbState.shareId, fbDispatch);
                    if (jump) {
                        nav(jump);
                    }
                }}
                onContextMenu={e => {
                    e.preventDefault();
                    e.stopPropagation();

                    fbDispatch({ type: 'set_menu_target', fileId: file.Id() });
                    fbDispatch({ type: 'set_menu_pos', pos: { x: e.clientX, y: e.clientY } });

                    if (fbState.menuMode === FbMenuModeT.Closed) {
                        fbDispatch({ type: 'set_menu_open', menuMode: FbMenuModeT.Default });
                    }
                }}
                onMouseLeave={e => {
                    fbDispatch({ type: 'set_hovering', hovering: '' });
                    if (fbState.draggingState && file.IsFolder()) {
                        fbDispatch({ type: 'set_move_dest', fileName: '' });
                    }
                    if (mouseDown) {
                        setMouseDown(null);
                    }
                }}
            >
                <FileVisualWrapper file={file} />
                <div className="file-size-box">
                    <p>{file.FormatSize()}</p>
                </div>
                <TextBox
                    itemId={file.Id()}
                    itemTitle={file.GetFilename()}
                    secondaryInfo={file.FormatSize()}
                    editing={editing}
                    setEditing={e => {
                        if (!context.allowEditing) {
                            return;
                        }
                        setEditing(e);
                    }}
                    allowEditing={context.allowEditing}
                    blockFocus={context.blockFocus}
                    rename={(id, newName) => {
                        if (newName === file.GetFilename() || (newName === '' && file.Id() !== 'NEW_DIR')) {
                            return;
                        }
                        context.rename(id, newName);
                    }}
                />

                {file.Id() === 'NEW_DIR' && !editing && (
                    <Loader color="white" size={20} style={{ position: 'absolute', top: 20, right: 20 }} />
                )}
            </div>
        );
    },
    (prev, next) => {
        if (prev.file.Id() !== next.file.Id()) {
            return false;
        } else if (prev.context !== next.context) {
            return false;
        } else if (prev.context.itemWidth !== next.context.itemWidth) {
            return false;
        } else if (prev.context.dragging !== next.context.dragging) {
            return false;
        } else if (prev.selected !== next.selected) {
            return false;
        } else if (prev.file.IsHovering() !== next.file.IsHovering()) {
            return false;
        }
        return true;
    },
);
