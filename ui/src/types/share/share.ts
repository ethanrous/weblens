import SharesApi from '@weblens/api/SharesApi'
import { ShareInfo, UserInfo } from '@weblens/api/swag'

import { ErrorHandler } from '../Types'

export class WeblensShare {
    shareId: string
    accessors: UserInfo[]
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

    GetAccessors(): UserInfo[] {
        return this.accessors
    }

    GetPublicLink(): string {
        return `${window.location.origin}/files/share/${this.shareId}/${this.fileId}`
    }

    async UpdateShare(isPublic: boolean, accessors: string[]) {
        if (!this.Id()) {
            throw new Error('Attempt to update share with no id')
        }

        if (isPublic !== this.public) {
            await SharesApi.setSharePublic(this.Id(), isPublic)
                .then(() => {
                    this.public = isPublic
                })
                .catch(ErrorHandler)
            this.public = isPublic
        }

        const add = accessors.filter(
            (x) => !this.accessors.find((u) => u.username === x)
        )
        const remove = this.accessors.filter(
            (x) => !accessors.includes(x.username)
        )

        if (remove.length !== 0 || add.length !== 0) {
            await SharesApi.setShareAccessors(this.Id(), {
                addUsers: add,
                removeUsers: remove.map((u) => u.username),
            })
                .then((res) => {
                    this.accessors = res.data.accessors
                })
                .catch(ErrorHandler)
        }
    }
}
