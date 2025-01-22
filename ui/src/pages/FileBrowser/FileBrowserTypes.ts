export interface FbViewOptsT {
    dirViewMode: DirViewModeT
    sortDirection: number // 1 or -1
    sortFunc: string
}

export enum DirViewModeT {
    Grid,
    List,
    Columns,
}

export const enum FbActionT {
    FileCreate = 'fileCreate',
    FileRestore = 'fileRestore',
    FileMove = 'fileMove',
    FileDelete = 'fileDelete',
    FileSizeChange = 'fileSizeChange',
}
