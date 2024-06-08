export interface ShareDataT {
    Accessors: string[]
    Expires: string
    Public: boolean
    shareId: string
    fileId: string
    ShareName: string
    Wormhole: boolean
}

export class WeblensShare {
    private data: ShareDataT

    constructor(init: ShareDataT) {
        if (!init) {
            console.error('Attempt to init share with no data')
            return
        }
        this.data = init
    }

    Id(): string {
        return this.data.shareId
    }

    IsPublic() {
        return this.data.Public
    }

    IsWormhole() {
        return this.data.Wormhole
    }

    GetFileId(): string {
        return this.data.fileId
    }

    GetAccessors(): string[] {
        return this.data.Accessors
    }

    GetPublicLink(): string {
        return `${window.location.origin}/files/share/${this.data.shareId}`
    }
}
