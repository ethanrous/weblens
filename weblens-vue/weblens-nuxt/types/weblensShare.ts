import type { PermissionsInfo, PermissionsParams, ShareInfo, UserInfo } from '@ethanrous/weblens-api'
import { useWeblensAPI } from '~/api/AllApi'

export default class WeblensShare implements ShareInfo {
    shareID: string = ''
    isDir: boolean = false
    accessors: UserInfo[] = []
    private _permissions: Record<string, PermissionsInfo> = {}
    expires: number = 0
    private _public: boolean = false
    fileID: string = ''
    shareName: string = ''
    wormhole: boolean = false
    owner: string = ''
    timelineOnly: boolean = false

    constructor(init: ShareInfo) {
        this.assign(init)
    }

    clone(): WeblensShare {
        return new WeblensShare({
            shareID: this.shareID,
            isDir: this.isDir,
            accessors: [...this.accessors],
            permissions: { ...this._permissions },
            expires: this.expires,
            public: this._public,
            timelineOnly: this.timelineOnly,
            fileID: this.fileID,
            shareName: this.shareName,
            wormhole: this.wormhole,
            owner: this.owner,
        })
    }

    private assign(init: ShareInfo) {
        if (!init) {
            return
        }

        this.shareID = init.shareID || ''
        this.fileID = init.fileID || ''
        this.shareName = init.shareName || ''
        this.isDir = init.isDir ?? false
        this.expires = init.expires || 0
        this._public = init.public ?? false
        this.wormhole = init.wormhole ?? false
        this.owner = init.owner || ''
        this.timelineOnly = init.timelineOnly ?? false

        if (init.accessors) {
            this.accessors = init.accessors
        }

        if (init.permissions) {
            this._permissions = init.permissions
        }
    }

    ID(): string {
        return this.shareID
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

    GetFileID(): string {
        return this.fileID
    }

    GetAccessors(): UserInfo[] {
        return this.accessors
    }

    GetLink(): string {
        return `${window.location.origin}/files/share/${this.shareID}${this.timelineOnly ? '?timeline=true' : ''}`
    }

    private get info(): ShareInfo {
        return {
            shareID: this.shareID,
            fileID: this.fileID,
            shareName: this.shareName,
            isDir: this.isDir,
            expires: this.expires,
            public: this._public,
            wormhole: this.wormhole,
            owner: this.owner,
            timelineOnly: this.timelineOnly,
            accessors: this.accessors,
            permissions: this._permissions,
        }
    }

    private async createShareIfNeeded() {
        if (this.shareID) {
            return
        }

        const { data: shareInfo } = await useWeblensAPI().SharesAPI.createFileShare({
            fileID: this.fileID,
            public: this._public,
            wormhole: this.wormhole,
        })

        this.assign(shareInfo)
    }

    public checkPermission(permission: keyof PermissionsParams, username?: string): boolean {
        if (!username) {
            username = useUserStore().getActiveUsername()
        }

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
        await this.createShareIfNeeded()

        const newInfo = (
            await useWeblensAPI().SharesAPI.addUserToShare(this.shareID, {
                username: username,
            })
        ).data

        if (!newInfo.accessors) {
            return
        }

        this.accessors = newInfo.accessors
    }

    public async removeAccessor(username: string) {
        const newInfo = (await useWeblensAPI().SharesAPI.removeUserFromShare(this.shareID, username)).data

        if (!newInfo.accessors) {
            return
        }

        this.accessors = newInfo.accessors
    }

    public async toggleIsPublic() {
        return this.setPublic(!this._public)
    }

    public async toggleTimelineOnly() {
        return this.setTimelineOnly(!this.timelineOnly)
    }

    public async setPublic(isPublic: boolean) {
        await this.createShareIfNeeded()

        if (this._public === isPublic) {
            return
        }

        this._public = isPublic
        await useWeblensAPI().SharesAPI.updateFileShare(this.shareID, this.info)
    }

    public async setTimelineOnly(timelineOnly: boolean) {
        await this.createShareIfNeeded()

        if (this.timelineOnly === timelineOnly) {
            return
        }

        this.timelineOnly = timelineOnly
        await useWeblensAPI().SharesAPI.updateFileShare(this.shareID, this.info)
    }

    public async updateAccessorPerms(user: string, perms: PermissionsParams) {
        const newShare = await useWeblensAPI().SharesAPI.updateShareAccessorPermissions(this.shareID, user, perms)
        this.assign(newShare.data)
    }
}
