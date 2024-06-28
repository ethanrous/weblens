export interface ShareDataT {
    id: string
    accessors: string[]
    expires: string
    public: boolean
    fileId: string
    shareName: string
    wormhole: boolean
}

export class WeblensShare {
    private data: ShareDataT

    constructor(init: ShareDataT) {
        if (!init) {
            console.error('Attempt to init share with no data')
            return
        }
        this.data = init
        if (!this.data.accessors) {
            this.data.accessors = []
        }
    }

    Id(): string {
        return this.data.id
    }

    IsPublic() {
        return this.data.public
    }

    IsWormhole() {
        return this.data.wormhole
    }

    GetFileId(): string {
        return this.data.fileId
    }

    GetAccessors(): string[] {
        return this.data.accessors
    }

    GetPublicLink(): string {
        return `${window.location.origin}/files/share/${this.data.id}/${this.data.fileId}`
    }
}
