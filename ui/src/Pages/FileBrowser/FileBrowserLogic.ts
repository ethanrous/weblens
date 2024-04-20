import { DragEvent, useCallback, useEffect, useState } from "react";

import { notifications } from "@mantine/notifications";

import Upload, { fileUploadMetadata } from "../../api/Upload";
import {
    FileInfoT,
    FbStateT,
    getBlankFile,
    FileBrowserAction,
    FBDispatchT,
    UserInfoT,
    AuthHeaderT,
    ScanMeta,
} from "../../types/Types";
import {
    CreateFolder,
    DeleteFiles,
    moveFiles,
    RenameFile,
    SubToFolder,
    UnsubFromFolder,
    downloadSingleFile,
    requestZipCreate,
} from "../../api/FileBrowserApi";
import { humanFileSize } from "../../util";
import useWeblensSocket from "../../api/Websocket";
import { useNavigate } from "react-router-dom";
import { ItemProps } from "../../components/ItemDisplay";

const handleSelect = (state: FbStateT, action: FileBrowserAction): FbStateT => {
    let numSelected = state.selected.size;
    if (state.holdingShift && numSelected > 0 && state.lastSelected !== "") {
        const dirList = Array.from(state.dirMap.values());
        let startIndex = dirList.findIndex((val) => val.id === state.lastSelected);
        let endIndex = dirList.findIndex((val) => val.id === action.fileId);

        if (endIndex < startIndex) {
            [startIndex, endIndex] = [endIndex, startIndex];
        }

        for (const val of dirList.slice(startIndex, endIndex + 1)) {
            state.selected.set(val.id, true);
        }

        return { ...state, lastSelected: action.fileId };
    } else {
        // If action.selected is undefined, i.e. not passed to the request,
        // we treat that as a request to toggle the selection
        if (action.selected === undefined) {
            if (state.selected.get(action.fileId)) {
                state.selected.delete(action.fileId);
            } else {
                state.selected.set(action.fileId, true);
                return { ...state, lastSelected: action.fileId };
            }
        }
        // state.selected.get returns undefined if not selected,
        // so we not (!) it to make boolean, and again to match... yay javascript :/
        else if (!!state.selected.get(action.fileId) === action.selected) {
            // If the file is already in the correct state, we do nothing.
            // Specifically, we do not overwrite lastSelected
        } else {
            if (action.selected) {
                state.selected.set(action.fileId, true);
            } else {
                state.selected.delete(action.fileId);
            }
        }

        if (state.selected.size === 0) {
            return { ...state, lastSelected: "" };
        }
    }
    return { ...state };
};

function createFile(dirMap: Map<string, FileInfoT>, user, newData: FileInfoT, currentFolderId) {
    if (dirMap.get(newData.id)) {
        console.warn("Taking no action creating file that already exists", newData.id);
        return;
    }

    if (newData.id === user.trashFolderId) {
        return;
    }

    if (newData.parentFolderId === currentFolderId || user.username !== newData.owner) {
        dirMap.set(newData.id, newData);
    }
}

function updateFile(state: FbStateT, user: UserInfoT, existingId: string, newData: FileInfoT) {
    let existingFile: FileInfoT = state.dirMap.get(existingId);
    if (!newData) {
        return;
    }

    if (newData.id === user.trashId) {
        if (state.folderInfo.id === user.trashId) {
            return {
                ...state,
                folderInfo: newData,
                trashDirSize: newData.size,
            };
        }

        return { ...state, trashDirSize: newData.size };
    }

    if (newData.id === user.homeId) {
        if (newData.id === state.folderInfo.id) {
            return { ...state, folderInfo: newData, homeDirSize: newData.size };
        }
        return { ...state, homeDirSize: newData.size };
    }

    if (newData.id === "EXTERNAL") {
        newData.id = newData.filename
        newData.filename = "External"
    }

    if (newData.id === state.folderInfo.id) {
        return { ...state, folderInfo: newData };
    }

    if (!existingFile && newData.parentFolderId !== state.contentId) {
        console.warn("Not upserting file not in view", existingId);
        return { ...state };
    }

    if (existingFile && existingFile.id !== newData.id) {
        state.dirMap.delete(existingFile.id);
    }

    if (newData.parentFolderId === state.folderInfo.id || user.username !== newData.owner) {
        state.dirMap.set(newData.id, newData);
    }

    return { ...state };
}

export const fileBrowserReducer = (state: FbStateT, action: FileBrowserAction): FbStateT => {
    switch (action.type) {
        case "create_file": {
            for (const fileInfo of action.fileInfos) {
                if (fileInfo.filename === state.waitingForNewName) {
                    state.dirMap.delete("NEW_DIR");
                    state.dirMap.set(fileInfo.id, fileInfo);
                    return { ...state, waitingForNewName: "" };
                }
                createFile(state.dirMap, action.user, fileInfo, state.folderInfo.id);
            }
            return { ...state, dirMap: new Map(state.dirMap) };
        }

        case "update_file": {
            for (const fileInfo of action.fileInfos) {
                state = updateFile(state, action.user, fileInfo.id, fileInfo);
            }

            return { ...state, dirMap: new Map(state.dirMap) };
        }

        case "replace_file": {
            state.dirMap.delete(action.fileId)
            const sel = state.selected.delete(action.fileId)
            if (action.fileInfo.parentFolderId !== state.folderInfo.id) {
                return { ...state, dirMap: new Map(state.dirMap), selected: new Map(state.selected) };
            }
            state.dirMap.set(action.fileInfo.id, action.fileInfo)
            if (sel) {
                state.selected.set(action.fileInfo.id, true)
            }

            return { ...state, dirMap: new Map(state.dirMap), selected: new Map(state.selected) };
        }

        case "update_many": {
            const filesMeta: { fileId: string; updateInfo: FileInfoT }[] = action.files;
            for (const fileMeta of filesMeta) {
                createFile(state.dirMap, action.user, fileMeta.updateInfo, state.folderInfo.id);
            }
            return { ...state, dirMap: new Map(state.dirMap) };
        }

        case "set_folder_info": {
            if (!action.fileInfo || !action.user) {
                console.error("Trying to set undefined file info or user");
                return { ...state };
            }
            if (action.fileInfo.id === action.user.homeId) {
                action.fileInfo.filename = "Home"
            }
            return { ...state, folderInfo: action.fileInfo };
        }

        case "set_parents_info": {
            return { ...state, parents: action.parents };
        }

        case "add_loading": {
            const newLoading = state.loading.filter((v) => v !== action.loading);
            newLoading.push(action.loading);
            return {
                ...state,
                loading: newLoading,
            };
        }

        case "remove_loading": {
            const newLoading = state.loading.filter((v) => v !== action.loading);
            // (action.loading)
            return {
                ...state,
                loading: newLoading,
            };
        }

        case "scan_complete": {
            let index = state.scanProgress.findIndex((s, i, a) => s.taskId === action.taskId);
            let sm: ScanMeta;
            if (index === -1) {
                sm = {
                    taskId: action.taskId,
                    taskType: action.taskType,
                    target: action.target,
                    tasksTotal: 0,
                    tasksComplete: 0,
                    progress: 0,
                    mostRecent: "",
                    complete: false,
                    time: 0,
                    note: "",
                };
                index = state.scanProgress.length;
                state.scanProgress.push(sm);
            }
            state.scanProgress[index].complete = true;
            state.scanProgress[index].time = action.time;
            if (action.note) {
                sm.note = action.note;
            }
            return {
                ...state,
                scanProgress: [...state.scanProgress],
            };
        }

        case "update_scan_progress": {
            let index = state.scanProgress.findIndex((s, i, a) => s.taskId === action.taskId);
            let sm: ScanMeta;
            if (index === -1) {
                sm = {
                    taskId: action.taskId,
                    taskType: action.taskType,
                    target: action.target,
                    tasksTotal: 0,
                    tasksComplete: 0,
                    progress: 0,
                    mostRecent: "",
                    complete: false,
                    time: 0,
                    note: "",
                };
                index = state.scanProgress.length;
                state.scanProgress.push(sm);
            } else {
                sm = state.scanProgress[index];
            }
            sm.progress = action.progress;
            sm.mostRecent = action.fileName;
            sm.tasksComplete = action.tasksComplete;
            sm.tasksTotal = action.tasksTotal;
            if (action.note) {
                sm.note = action.note;
            }
            state.scanProgress[index] = sm;

            return {
                ...state,
                scanProgress: [...state.scanProgress],
            };
        }

        case "remove_task_progress": {
            state.scanProgress = state.scanProgress.filter((p) => Boolean(p.taskId));
            let index = state.scanProgress.findIndex((s, i, a) => s.taskId === action.taskId);
            state.scanProgress.splice(index, 1);

            return { ...state, scanProgress: [...state.scanProgress] };
        }

        case "set_search": {
            return {
                ...state,
                searchContent: action.search,
            };
        }

        case "set_dragging": {
            let dragging: number;

            if (!action.dragging) {
                dragging = 0;
            } else if (action.dragging && !action.external) {
                dragging = 1;
            } else if (action.dragging && action.external) {
                dragging = 2;
            }

            return {
                ...state,
                draggingState: dragging,
            };
        }

        case "set_hovering": {
            return { ...state, hovering: action.hovering };
        }

        case "new_dir": {
            let newDir: FileInfoT = getBlankFile();
            newDir.id = "NEW_DIR";
            newDir.filename = "New Folder";
            newDir.isDir = true;
            newDir.modifiable = true;
            newDir.parentFolderId = state.folderInfo.id;
            state.dirMap.set(newDir.id, newDir);
            console.log("HERE!", state.dirMap)
            return { ...state, dirMap: new Map(state.dirMap) };
        }

        case "set_selected": {
            state = handleSelect(state, action);
            return { ...state, selected: new Map(state.selected) };
        }

        case "select_all": {
            for (const fileId of state.filesList) {
                state.selected.set(fileId, true);
            }
            return { ...state, menuOpen: false, selected: new Map(state.selected) };
        }

        case "select_ids": {
            for (const id of action.fileIds) {
                state.selected.set(id, true);
            }
            return { ...state, selected: new Map(state.selected) };
        }

        case "set_block_focus": {
            return { ...state, blockFocus: action.block };
        }

        case "clear_files": {
            state.dirMap.clear();
            state.selected.clear();

            return {
                ...state,
                folderInfo: getBlankFile(),
                parents: [],
                lastSelected: "",
            };
        }

        case "clear_selected": {
            if (state.selected.size === 0) {
                return state;
            }

            return {
                ...state,
                lastSelected: "",
                selected: new Map<string, boolean>(),
            };
        }

        case "delete_from_map": {
            for (const fileId of action.fileIds) {
                state.dirMap.delete(fileId);
                state.selected.delete(fileId);
            }

            return { ...state, dirMap: new Map(state.dirMap), selected: new Map(state.selected) };
        }

        case "holding_shift": {
            return {
                ...state,
                holdingShift: action.shift,
            };
        }

        case "stop_presenting":
        case "set_presentation": {
            if (action.presentingId) {
                state.selected.clear();
                state.selected.set(action.presentingId, true);
            }
            return {
                ...state,
                presentingId: action.presentingId,
            };
        }

        case "set_col_count": {
            return { ...state, numCols: action.numCols };
        }

        case "set_files_list": {
            return { ...state, filesList: [...action.fileIds] };
        }

        case "set_menu_open": {
            return {
                ...state,
                menuOpen: action.open,
                menuTargetId: action.open ? state.menuTargetId : "",
            };
        }

        case "set_menu_target": {
            return { ...state, menuTargetId: action.fileId };
        }

        case "set_menu_pos": {
            return { ...state, menuPos: action.pos };
        }

        case "presentation_next": {
            const index = state.filesList.indexOf(state.lastSelected);
            let lastSelected = state.lastSelected;
            if (index + 1 < state.filesList.length) {
                state.selected.clear();
                lastSelected = state.filesList[index + 1];
                state.selected.set(lastSelected, true);
            }
            return {
                ...state,
                lastSelected: lastSelected,
                presentingId: lastSelected,
            };
        }

        case "presentation_previous": {
            const index = state.filesList.indexOf(state.lastSelected);
            let lastSelected = state.lastSelected;
            if (index - 1 >= 0) {
                state.selected.clear();
                lastSelected = state.filesList[index - 1];
                state.selected.set(lastSelected, true);
            }
            return {
                ...state,
                lastSelected: lastSelected,
                presentingId: lastSelected,
            };
        }

        case "move_selection": {
            if (state.presentingId) {
                return { ...state };
            }
            let lastSelected = state.lastSelected;
            const prevIndex = state.lastSelected ? state.filesList.indexOf(state.lastSelected) : -1;
            let finalIndex = -1;
            if (action.direction === "ArrowDown") {
                if (prevIndex === -1) {
                    finalIndex = 0;
                } else if (prevIndex + state.numCols < state.filesList.length) {
                    finalIndex = prevIndex + state.numCols;
                }
            } else if (action.direction === "ArrowUp") {
                if (prevIndex === -1) {
                    finalIndex = state.filesList.length - 1;
                } else if (prevIndex - state.numCols >= 0) {
                    finalIndex = prevIndex - state.numCols;
                }
            } else if (action.direction === "ArrowLeft") {
                if (prevIndex === -1) {
                    finalIndex = state.filesList.length - 1;
                }
                if (prevIndex - 1 >= 0 && prevIndex % state.numCols !== 0) {
                    finalIndex = prevIndex - 1;
                }
            } else if (action.direction === "ArrowRight") {
                if (prevIndex === -1) {
                    finalIndex = 0;
                } else if (prevIndex + 1 < state.filesList.length && prevIndex % state.numCols !== state.numCols - 1) {
                    finalIndex = prevIndex + 1;
                }
            }

            if (finalIndex !== -1) {
                if (!state.holdingShift) {
                    state.selected.clear();
                    state.selected.set(state.filesList[finalIndex], true);
                } else {
                    if (prevIndex < finalIndex) {
                        for (const file of state.filesList.slice(prevIndex, finalIndex + 1)) {
                            state.selected.set(file, true);
                        }
                    } else {
                        for (const file of state.filesList.slice(finalIndex, prevIndex + 1)) {
                            state.selected.set(file, true);
                        }
                    }
                }
                lastSelected = state.filesList[finalIndex];
            }

            return {
                ...state,
                lastSelected: lastSelected,
                presentingId: state.presentingId ? lastSelected : "",
                selected: new Map(state.selected),
            };
        }

        case "paste_image": {
            return { ...state, pasteImg: action.img };
        }

        case "set_scroll_to": {
            return { ...state, scrollTo: action.fileId };
        }

        // When we are waiting for a new file to be created, we don't know the id
        // so we wait to see the file with the right name to be created
        case "set_waiting_for": {
            return { ...state, waitingForNewName: action.fileName };
        }

        case "set_move_dest": {
            return { ...state, moveDest: action.fileName };
        }

        case "set_location_state": {
            return { ...state, contentId: action.realId, fbMode: action.mode, shareId: action.shareId };
        }

        case "set_search_results": {
            return { ...state, searchResults: action.fileInfos };
        }

        case "set_sort": {
            if (action.sortType) {
                return { ...state, sortFunc: action.sortType };
            } else if (action.sortDirection) {
                return { ...state, sortDirection: action.sortDirection };
            } else {
                return { ...state };
            }
        }

        case "set_past_time": {
            return {...state, viewingPast: action.past}
        }

        case "set_file_info_menu": {
            return {...state, fileInfoMenu: action.open}
        }

        default: {
            console.error("Got unexpected dispatch type: ", action.type);
            notifications.show({
                title: "Unexpected filebrowser dispatch",
                message: action.type,
                color: "red",
            });
            return { ...state };
        }
    }
};

export function getSortFunc(sortType: string, sortDirection: number) {
    switch (sortType) {
        case "Name":
            return (a: ItemProps, b: ItemProps) => a.itemTitle.localeCompare(b.itemTitle) * sortDirection;
        case "Date Modified":
            return (a: ItemProps, b: ItemProps) => {
                return (b.modifyDate.getTime() - a.modifyDate.getTime()) * sortDirection;
            };
        case "Size":
            return (a: ItemProps, b: ItemProps) => (b.itemSizeBytes - a.itemSizeBytes) * sortDirection;
    }
}

export const useSubscribe = (cId: string, sId: string, mode: string, usr: UserInfoT, dispatch: FBDispatchT, authHeader: AuthHeaderT) => {
    const { wsSend, lastMessage, readyState } = useWeblensSocket();

    useEffect(() => {
        if (readyState === 1 && cId != null && cId !== "shared" && usr.isLoggedIn) {
            if (sId !== cId) {
                return;
            }
            if (cId === usr.homeId) {
                SubToFolder(usr.trashId, false, wsSend);
                return () => UnsubFromFolder(usr.trashId, wsSend);
            }
            SubToFolder(cId, false, wsSend);
            return () => UnsubFromFolder(cId, wsSend);
        }
    }, [readyState, sId, cId, usr.homeId, usr.trashId, wsSend]);

    // Subscribe to the home folder if we aren't in it, to be able to update the total disk usage
    useEffect(() => {
        if (!usr.isLoggedIn || readyState !== 1) {
            return;
        }
        SubToFolder(usr.homeId, false, wsSend);
        return () => UnsubFromFolder(usr.homeId, wsSend);
    }, [usr.homeId, readyState]);

    // Listen for incoming websocket messages
    useEffect(() => {
        HandleWebsocketMessage(lastMessage, cId, mode, usr, dispatch, authHeader);
    }, [lastMessage, usr, authHeader]);

    return {wsSend, readyState};
};

export const getRealId = async (
    contentId: string,
    mode: string,
    usr: UserInfoT,
    authHeader: AuthHeaderT,
) => {

    if (mode === "stats" && contentId === "external") {
        return "EXTERNAL"
    }

    if (contentId === "home") {
        return usr.homeId;
    } else if (contentId === "trash") {
        return usr.trashId;
    } else if (!contentId) {
        return ''
    } else {
        return contentId;
    }
};

export const handleDragOver = (event: DragEvent, dispatch: FBDispatchT, dragging: number) => {
    event.preventDefault();
    event.stopPropagation();

    if (event.type === "dragenter" || event.type === "dragover") {
        !dragging &&
            dispatch({
                type: "set_dragging",
                dragging: true,
                external: Boolean(event.dataTransfer.types.length),
            });
    } else {
        dispatch({ type: "set_dragging", dragging: false });
    }
};

export const handleRename = (
    itemId: string,
    newName: string,
    folderId: string,
    selectedCount: number,
    dispatch: FBDispatchT,
    authHeader: AuthHeaderT,
) => {
    // When we are creating a new folder, the id is initially ""
    if (itemId === "NEW_DIR") {
        // If we do not get a new name, the rename is canceled
        if (newName === "") {
            dispatch({ type: "delete_from_map", fileIds: ["NEW_DIR"] });
        } else {
            dispatch({ type: "add_loading", loading: "newFolder" });
            dispatch({ type: "set_waiting_for", fileName: newName });

            CreateFolder(folderId, newName, false, "", authHeader).then((d) => {
                if (selectedCount === 0) {
                    dispatch({ type: "set_selected", fileId: d });
                }
                dispatch({ type: "remove_loading", loading: "newFolder" });
            });
        }
    } else {
        dispatch({ type: "add_loading", loading: "renameFile" });
        RenameFile(itemId, newName, authHeader).then((_) =>
            dispatch({ type: "remove_loading", loading: "renameFile" }),
        );
    }
};

async function getFile(file): Promise<File> {
    try {
        const f = await file.getAsFile();
        return f;
        // return new Promise((resolve, reject) => file.file(resolve, reject));
    } catch (err) {
        return await new Promise((resolve, reject) => file.file(resolve, reject));

        // return new Promise((resolve, reject) => file)
    }
}

async function addDir(
    fsEntry,
    parentFolderId: string,
    topFolderKey: string,
    rootFolderId: string,
    isPublic: boolean,
    shareId: string,
    authHeader: AuthHeaderT,
): Promise<any> {
    return await new Promise(
        async (resolve: (value: fileUploadMetadata[]) => void, reject): Promise<fileUploadMetadata[]> => {
            if (fsEntry.isDirectory === true) {
                const folderId = await CreateFolder(parentFolderId, fsEntry.name, isPublic, shareId, authHeader);
                if (!folderId) {
                    reject();
                }
                let e: fileUploadMetadata = null;
                if (!topFolderKey) {
                    topFolderKey = folderId;
                    e = {
                        file: fsEntry,
                        isDir: true,
                        folderId: folderId,
                        parentId: rootFolderId,
                        isTopLevel: true,
                        topLevelParentKey: null,
                    };
                }

                let dirReader = fsEntry.createReader();
                // addDir(entry, parentFolderId, topFolderKey, rootFolderId, authHeader)
                const entriesPromise = new Promise((resolve: (value: any[]) => void, reject) => {
                    let allEntries = [];

                    const reader = (callback) => (entries) => {
                        if (entries.length === 0) {
                            resolve(allEntries);
                            return;
                        }

                        for (const entry of entries) {
                            allEntries.push(entry);
                        }

                        if (entries.length !== 100) {
                            resolve(allEntries);
                            return;
                        }
                        dirReader.readEntries(callback(callback));
                    };

                    dirReader.readEntries(reader(reader));
                });

                let allResults = [];
                if (e !== null) {
                    allResults.push(e);
                }
                for (const entry of await entriesPromise) {
                    allResults.push(
                        ...(await addDir(entry, folderId, topFolderKey, rootFolderId, isPublic, shareId, authHeader)),
                    );
                }
                resolve(allResults);
            } else {
                if (fsEntry.name === ".DS_Store") {
                    resolve([]);
                    return;
                }
                const f = await getFile(fsEntry);
                let e: fileUploadMetadata = {
                    file: f,
                    parentId: parentFolderId,
                    isDir: false,
                    isTopLevel: parentFolderId === rootFolderId,
                    topLevelParentKey: topFolderKey,
                };
                resolve([e]);
            }
        },
    );
}

export async function HandleDrop(
    entries,
    rootFolderId: string,
    conflictNames: string[],
    isPublic: boolean,
    shareId: string,
    authHeader: AuthHeaderT,
    uploadDispatch,
    wsSend: (action: string, content: any) => void,
) {
    let files: fileUploadMetadata[] = [];
    let topLevels = [];
    if (entries) {
        // Handle Directory
        for (const entry of entries) {
            if (!entry) {
                console.error("Upload entry does not exist or is not a file");
                continue;
            }
            const file = entry.webkitGetAsEntry();
            if (!file) {
                console.error("Drop is not a file");
                continue;
            }
            if (conflictNames.includes(file.name)) {
                notifications.show({
                    title: `Cannot upload "${file.name}"`,
                    message: "A file or folder with that name already exists in this folder",
                    autoClose: 10000,
                    color: "red",
                });
                continue;
            }
            topLevels.push(
                addDir(file, rootFolderId, null, rootFolderId, isPublic, shareId, authHeader)
                    .then((newFiles) => {
                        files.push(...newFiles);
                    })
                    .catch((r) => {
                        notifications.show({
                            message: String(r),
                            color: "red",
                        });
                    }),
            );
        }
    }

    await Promise.all(topLevels);

    if (files.length !== 0) {
        Upload(files, isPublic, shareId, rootFolderId, authHeader, uploadDispatch, wsSend);
    }
}

export function HandleUploadButton(
    files: File[],
    parentFolderId: string,
    isPublic: boolean,
    shareId: string,
    authHeader: AuthHeaderT,
    uploadDispatch,
    wsSend: (action: string, content: any) => void,
) {
    let uploads: fileUploadMetadata[] = [];
    for (const f of files) {
        uploads.push({
            file: f,
            parentId: parentFolderId,
            isDir: false,
            isTopLevel: true,
            topLevelParentKey: parentFolderId,
        });
    }

    if (uploads.length !== 0) {
        Upload(uploads, isPublic, shareId, parentFolderId, authHeader, uploadDispatch, wsSend);
    }
}

export function downloadSelected(
    files: FileInfoT[],
    dispatch: FBDispatchT,
    wsSend: (action: string, content: any) => void,
    authHeader: AuthHeaderT,
    shareId?: string,
) {
    let taskId: string = "";

    if (files.length === 1 && !files[0].isDir) {
        downloadSingleFile(files[0].id, authHeader, dispatch, files[0].filename, undefined, shareId);
        return;
    }

    requestZipCreate(
        files.map((f) => f.id),
        shareId,
        authHeader,
    )
        .then(({ json, status }) => {
            if (status === 200) {
                downloadSingleFile(json.takeoutId, authHeader, dispatch, undefined, "zip", shareId);
            } else if (status === 202) {
                taskId = json.taskId;
                notifications.show({
                    id: `zip_create_${taskId}`,
                    message: `Requesting zip...`,
                    autoClose: false,
                });
                wsSend("subscribe", {
                    subscribeType: "task",
                    subscribeMeta: JSON.stringify({
                        lookingFor: ["takeoutId"],
                    }),
                    subscribeKey: taskId,
                });
            } else if (status !== 0) {
                notifications.show({
                    title: "Failed to request takeout",
                    message: String(json.error),
                    color: "red",
                });
            }
            dispatch({ type: "remove_loading", loading: "zipCreate" });
        })
        .catch((r) => console.error(r));
}

export function HandleWebsocketMessage(
    lastMessage,
    contentId: string,
    fbMode: string,
    usr: UserInfoT,
    dispatch: FBDispatchT,
    authHeader: AuthHeaderT,
) {
    if (lastMessage) {
        let msgData = JSON.parse(lastMessage.data);
        console.log("WSRecv", msgData);
        switch (msgData.eventTag) {
            case "file_created": {
                dispatch({
                    type: "create_file",
                    fileInfos: msgData.content.map((v) => v.fileInfo),
                    user: usr,
                });
                return;
            }
            case "file_updated": {
                dispatch({
                    type: "update_file",
                    fileInfos: msgData.content.map((v) => v.fileInfo),
                    user: usr,
                });
                return;
            }
            // moved is different from updated because the Id of the file will change
            case "file_moved": {
                msgData.content.map(m => {
                    dispatch({
                        type: "replace_file",
                        fileId: m.oldId,
                        fileInfo: m.newFile,
                    });
                })
                return;
            }
            case "file_deleted": {
                dispatch({
                    type: "delete_from_map",
                    fileIds: msgData.content.map((v) => v.fileId),
                });
                return;
            }
            case "scan_complete": {
                dispatch({
                    type: "scan_complete",
                    taskId: msgData.subscribeKey,
                    time: msgData.content[0].execution_time,
                    note: msgData.content[0].note,
                });
                return;
            }
            case "sub_task_complete": {
                const jobName = msgData.content[0].task_job_name
                    .replace("_", " ")
                    .split(" ")
                    .map((s) => {
                        return s.charAt(0).toUpperCase() + s.slice(1);
                    })
                    .join(" ")
                    .replace("Directory", "Folder");

                dispatch({
                    type: "update_scan_progress",
                    progress: msgData.content[0].percent_progress,
                    taskId: msgData.subscribeKey,
                    taskType: jobName,
                    target: msgData.content[0].task_job_target,
                    fileName: msgData.content[0].filename,
                    tasksComplete: msgData.content[0].tasks_complete,
                    tasksTotal: msgData.content[0].tasks_total,
                    note: msgData.content[0].note,
                });
                return;
            }
            case "scan_directory_progress": {
                dispatch({
                    type: "set_scan_progress",
                    progress: (1 - msgData.content["remainingTasks"] / msgData.content["totalTasks"]) * 100,
                });
                return;
            }
            case "create_zip_progress": {
                let content = msgData.content[0].result;
                let [speed, units] = humanFileSize(content["speedBytes"]);
                notifications.update({
                    id: `zip_create_${msgData.subscribeKey}`,
                    message: `Compressing ${content["totalFiles"]} files: ${content["completedFiles"]} / ${content["totalFiles"]} (${speed}${units}/s)`,
                    autoClose: false,
                    loading: true,
                });
                dispatch({
                    type: "set_scan_progress",
                    progress: (content["completedFiles"] / content["totalFiles"]) * 100,
                });
                return;
            }
            case "zip_complete": {
                notifications.hide(`zip_create_${msgData.subscribeKey}`);
                downloadSingleFile(
                    msgData.content[0].result["takeoutId"],
                    authHeader,
                    dispatch,
                    undefined,
                    "zip",
                    fbMode === "share" ? contentId : ""

                );
                return;
            }
            case "error": {
                notifications.show({
                    title: "Websocket error",
                    message: msgData.error,
                    color: "red",
                });
                return;
            }
            default: {
                console.error("Could not parse websocket message type: ", msgData);
                return;
            }
        }
    }
}

export const useKeyDownFileBrowser = (
    fbState: FbStateT,
    searchQuery: string,
    usr,
    dispatch: (action: FileBrowserAction) => void,
    authHeader,
    wsSend,
    searchRef,
) => {
    const nav = useNavigate();
    useEffect(() => {
        const onKeyDown = (event) => {
            if (!fbState.blockFocus) {
                if (document.activeElement === searchRef.current) {
                    if (event.key === "Enter") {
                        console.log("HERE")
                        if (!Boolean(fbState.searchContent)) {
                            if (fbState.fbMode === "search") {
                                nav(`/files/${fbState.contentId}`);
                            }
                            return
                        }
                        nav(`/files/search/${fbState.contentId}?query=${fbState.searchContent}`, {
                            replace: Boolean(searchQuery),
                        });
                    } else if (event.key === "Escape") {
                        searchRef.current.blur();
                    } else {
                        if (event.metaKey && event.key === "a") {
                            event.stopPropagation();
                        }
                    }

                    return;
                }

                if (event.metaKey && event.key === "a") {
                    event.preventDefault();
                    dispatch({ type: "select_all" });
                } else if (
                    !event.metaKey &&
                    (event.key === "ArrowUp" ||
                        event.key === "ArrowDown" ||
                        event.key === "ArrowLeft" ||
                        event.key === "ArrowRight")
                ) {
                    event.preventDefault();
                    dispatch({
                        type: "move_selection",
                        direction: event.key,
                    });
                } else if (
                    !event.metaKey &&
                    ((event.which >= 49 && event.which <= 90) ||
                        (event.key === "Backspace" && Boolean(fbState.searchContent)))
                ) {
                    searchRef.current.focus();
                } else if (event.key === "Escape") {
                    event.preventDefault();
                    if (fbState.pasteImg) {
                        dispatch({ type: "paste_image", img: null });
                    } else {
                        dispatch({ type: "clear_selected" });
                    }
                } else if (event.key === "Shift") {
                    dispatch({ type: "holding_shift", shift: true });
                } else if (event.key === "Enter" && fbState.pasteImg) {
                    if (fbState.folderInfo.id === "shared" || fbState.folderInfo.id === usr.trashFolderId) {
                        notifications.show({
                            title: "Paste blocked",
                            message: "This folder does not allow paste-to-upload",
                            color: "red",
                        });
                        return;
                    }
                    uploadViaUrl(fbState.pasteImg, fbState.folderInfo.id, fbState.dirMap, authHeader, dispatch, wsSend);
                } else if (event.key === " ") {
                    event.preventDefault();
                    if (fbState.lastSelected && !fbState.presentingId) {
                        dispatch({
                            type: "set_presentation",
                            presentingId: fbState.lastSelected,
                        });
                    } else if (fbState.presentingId) {
                        dispatch({ type: "stop_presenting" });
                    }
                }
            }
        };

        const onKeyUp = (event) => {
            if (!fbState.blockFocus) {
                if (event.key === "Shift") {
                    dispatch({ type: "holding_shift", shift: false });
                }
            }
        };

        document.addEventListener("keydown", onKeyDown);
        document.addEventListener("keyup", onKeyUp);
        return () => {
            document.removeEventListener("keydown", onKeyDown);
            document.removeEventListener("keyup", onKeyUp);
        };
    }, [
        fbState.blockFocus,
        fbState.searchContent,
        searchQuery,
        fbState.pasteImg,
        dispatch,
        searchRef,
        fbState.presentingId,
        fbState.lastSelected,
    ]);
};

export const useMousePosition = () => {
    const [mousePosition, setMousePosition] = useState({ x: null, y: null });

    useEffect(() => {
        const updateMousePosition = (ev) => {
            setMousePosition({ x: ev.clientX, y: ev.clientY });
        };
        window.addEventListener("mousemove", updateMousePosition);
        return () => {
            window.removeEventListener("mousemove", updateMousePosition);
        };
    }, []);
    return mousePosition;
};

export const usePaste = (
    folderId: string,
    usr: UserInfoT,
    searchRef,
    blockFocus: boolean,
    dispatch: (Action: FileBrowserAction) => void,
) => {
    const handlePaste = useCallback(
        async (e) => {
            if (blockFocus) {
                return;
            }
            e.preventDefault();
            e.stopPropagation();

            const clipboardItems =
                typeof navigator?.clipboard?.read === "function"
                    ? await navigator.clipboard.read().catch((v) => {
                          console.error(v);
                          notifications.show({
                              title: "Could not paste",
                              message: "Does your browser block clipboard for Weblens?",
                              color: "red",
                          });
                      })
                    : e.clipboardData?.files;
            if (!clipboardItems) {
                return;
            }
            for (const item of clipboardItems) {
                for (const mime of item.types) {
                    if (mime.startsWith("image/")) {
                        if (folderId === "shared" || folderId === usr.trashId) {
                            notifications.show({
                                title: "Paste blocked",
                                message: "This folder does not allow paste-to-upload",
                                color: "red",
                            });
                            return;
                        }
                        const img = await item.getType(mime);
                        dispatch({ type: "paste_image", img: img });
                    } else if (mime === "text/plain") {
                        const text = await (await item.getType("text/plain"))?.text();
                        if (!text) {
                            continue;
                        }
                        searchRef.current.focus();
                        dispatch({ type: "set_search", search: text });
                    } else {
                        console.error("Unknown mime", mime);
                    }
                }
            }
        },
        [folderId, blockFocus],
    );

    useEffect(() => {
        window.addEventListener("paste", handlePaste);
        return () => {
            window.removeEventListener("paste", handlePaste);
        };
    }, [handlePaste]);
};

export function deleteSelected(
    selectedMap: Map<string, boolean>,
    dirMap: Map<string, FileInfoT>,
    authHeader: AuthHeaderT,
) {
    const fileIds = Array.from(selectedMap.keys());
    DeleteFiles(fileIds, authHeader);
}

export function MoveSelected(selectedMap: Map<string, boolean>, destinationId: string, authHeader: AuthHeaderT) {
    return moveFiles(Array.from(selectedMap.keys()), destinationId, authHeader).catch((r) =>
        notifications.show({
            title: "Failed to move files",
            message: String(r),
            color: "red",
        }),
    );
}

export async function uploadViaUrl(
    img: ArrayBuffer,
    folderId: string,
    dirMap: Map<string, FileInfoT>,
    authHeader: AuthHeaderT,
    dispatch: (Action: FileBrowserAction) => void,
    wsSend,
) {
    const names = Array.from(dirMap.values()).map((v) => v.filename);
    let imgNumber = 1;
    let imgName = `image${imgNumber}.jpg`;
    while (names.includes(imgName)) {
        imgNumber++;
        imgName = `image${imgNumber}.jpg`;
    }

    const meta: fileUploadMetadata = {
        file: new File([img], imgName),
        isDir: false,
        parentId: folderId,
        topLevelParentKey: "",
        isTopLevel: true,
    };
    await Upload([meta], false, "", folderId, authHeader, () => {}, wsSend);
    dispatch({ type: "paste_image", img: null });
}

export function selectedMediaIds(dirMap: Map<string, FileInfoT>, selectedIds: string[]): string[] {
    return selectedIds.map((id) => dirMap.get(id)?.mediaData?.mediaId).filter((v) => Boolean(v));
}

export function selectedFolderIds(dirMap: Map<string, FileInfoT>, selectedIds: string[]): string[] {
    return selectedIds.filter((id) => dirMap.get(id).isDir);
}

export function SetFileData(
    data: {
        self?: FileInfoT;
        children?: FileInfoT[];
        parents?: FileInfoT[];
        error?: any;
    },
    dispatch,
    usr,
) {
    if (!data) {
        console.error("Trying to set null file data");
        return;
    }
    if (data.error) {
        console.error(data.error);
        return;
    }
    let children = data.children?.map((val: FileInfoT) => {
        return { fileId: val.id, updateInfo: val };
    });
    if (!children) {
        children = [];
    }
    let parents;
    if (!data.parents) {
        parents = [];
    } else {
        parents = data.parents.reverse();
    }

    dispatch({ type: "set_folder_info", fileInfo: data.self, user: usr });
    dispatch({ type: "update_many", files: children, user: usr });
    dispatch({ type: "set_parents_info", parents: parents });
}

export function getVisitRoute(fbMode: string, cId: string, sId: string, isDir: boolean, isDisplayable: boolean, dispatch: FBDispatchT) {
    if (fbMode === 'share' && !sId) {

    } else if (fbMode === 'share') {
        return `/files/shared/${sId}/${cId}`
    } else if (fbMode === 'external') {
        return `/files/external/${cId}`
    } else if (isDir) {
        return cId
    } else if (isDisplayable) {
        dispatch({
            type: 'set_presentation',
            presentingId: cId,
        });
    }
}