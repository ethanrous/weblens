import SharesApi from '@weblens/api/SharesApi'
import {
    PermissionsInfo,
    PermissionsParams,
    ShareInfo,
    UserInfo,
} from '@weblens/api/swag'

export class WeblensShare {
    shareId: string
    accessors: UserInfo[] = []
    private _permissions: Record<string, PermissionsInfo> = {}
    expires: number
    private _public: boolean
    fileId: string
    shareName: string
    wormhole: boolean
    owner: string

    constructor(init?: ShareInfo) {
        this.assign(init)
    }

    private assign(init: ShareInfo) {
        if (!init) {
            return
        }

        this.shareId = init.shareId
        this.fileId = init.fileId
        this.shareName = init.shareName
        this.expires = init.expires
        this._public = init.public
        this.wormhole = init.wormhole
        this.owner = init.owner

        if (init.accessors) {
            this.accessors = init.accessors
        }

        if (init.permissions) {
            this._permissions = init.permissions
        }
    }

    Id(): string {
        return this.shareId
    }

    IsPublic() {
        return this._public
    }

    public get public(): boolean {
        return this._public
    }

    public get permissions(): Record<string, PermissionsParams> {
        return this._permissions
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

    private async createShare() {
        if (this.shareId) {
            return
        }

        const { data: shareInfo } = await SharesApi.createFileShare({
            fileId: this.fileId,
            public: this._public,
            wormhole: this.wormhole,
        })

        this.assign(shareInfo)
    }

    public checkPermission(
        username: string,
        permission: keyof PermissionsParams
    ): boolean {
        if (this.owner === username) {
            return true
        }

        if (!this._permissions[username]) {
            return false
        }

        const perms = this._permissions[username]
        return !!perms[permission]
    }

    public async addAccessor(username: string) {
        await this.createShare()

        return await SharesApi.addUserToShare(this.shareId, {
            username: username,
        })
    }

    public async removeAccessor(username: string) {
        return await SharesApi.removeUserFromShare(this.shareId, username)
    }

    public async setPublic(isPublic: boolean) {
        await this.createShare()

        if (this._public === isPublic) {
            return
        }

        this._public = isPublic
        return await SharesApi.setSharePublic(this.shareId, isPublic)
    }

    public async updateAccessorPerms(user: string, perms: PermissionsParams) {
        await SharesApi.updateShareAccessorPermissions(
            this.shareId,
            user,
            perms
        )
    }
}
