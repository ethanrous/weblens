import {
    setFileSharePublic,
    setShareAccessors,
} from '@weblens/types/share/shareQuery'

export interface ShareInfo {
    id: string
    accessors?: string[]
    expires?: string
    public?: boolean
    fileId?: string
    shareName?: string
    wormhole?: boolean
}

export class WeblensShare {
    shareId: string
    accessors: string[]
    expires: string
    public: boolean
    fileId: string
    shareName: string
    wormhole: boolean

    constructor(init: ShareInfo) {
        if (!init) {
            console.error('Attempt to init share with no data')
            return
        }
        Object.assign(this, init)
        if (!this.accessors) {
            this.accessors = []
        }
    }

    Id(): string {
        return this.shareId
    }

    IsPublic() {
        return this.public
    }

    IsWormhole() {
        return this.wormhole
    }

    GetFileId(): string {
        return this.fileId
    }

    GetAccessors(): string[] {
        return this.accessors
    }

    GetPublicLink(): string {
        return `${window.location.origin}/files/share/${this.shareId}/${this.fileId}`
    }

    async UpdateShare(isPublic: boolean, accessors: string[]) {
        if (isPublic !== this.public) {
            await setFileSharePublic(this.Id(), isPublic)
            this.public = isPublic
        }

        const add = accessors.filter((x) => !this.accessors.includes(x))
        const remove = this.accessors.filter((x) => !accessors.includes(x))

        if (remove.length !== 0 || add.length !== 0) {
            await setShareAccessors(this.Id(), add, remove)
            this.accessors = accessors
        }
    }
}
