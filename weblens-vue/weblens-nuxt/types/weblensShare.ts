import type { PermissionsInfo, PermissionsParams, ShareInfo, UserInfo } from '@ethanrous/weblens-api'
import { useWeblensAPI } from '~/api/AllApi'

export default class WeblensShare implements ShareInfo {
    private _permissions: Record<string, PermissionsInfo> = {}
    private _public: boolean = false

    shareID: string = ''
    isDir: boolean = false
    accessors: UserInfo[] = []
    expires: number = 0
    fileID: string = ''
    shareName: string = ''
    wormhole: boolean = false
    owner: string = ''
    timelineOnly: boolean = false

    constructor(init: ShareInfo) {
        this.assign(init)
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

    private async updateShare(newInfo?: Partial<ShareInfo>) {
        if (!newInfo) {
            const res = await useWeblensAPI().SharesAPI.updateFileShare(this.shareID, this.info)
            newInfo = res.data
        }

        this.assign(newInfo)
    }

    public ID(): string {
        return this.shareID
    }

    public IsPublic() {
        return this._public
    }

    public get public(): boolean {
        return this._public
    }

    public get permissions(): Record<string, PermissionsParams> {
        return this._permissions
    }

    public IsWormhole() {
        return this.wormhole
    }

    public GetFileID(): string {
        return this.fileID
    }

    public GetAccessors(): UserInfo[] {
        return this.accessors
    }

    public GetLink(timeline: boolean = false): string {
        return `${window.location.origin}/files/share/${this.shareID}${timeline ? '?timeline=true' : ''}`
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
        const newInfo = (
            await useWeblensAPI().SharesAPI.addUserToShare(this.shareID, {
                username: username,
            })
        ).data

        if (!newInfo.accessors) {
            return
        }

        await this.updateShare(newInfo)
    }

    public async removeAccessor(username: string) {
        const newInfo = (await useWeblensAPI().SharesAPI.removeUserFromShare(this.shareID, username)).data

        if (!newInfo.accessors) {
            return
        }

        await this.updateShare(newInfo)
    }

    public async updateAccessorPerms(user: string, perms: PermissionsParams) {
        const newShare = await useWeblensAPI().SharesAPI.updateShareAccessorPermissions(this.shareID, user, perms)
        await this.updateShare(newShare.data)
    }

    public async toggleIsPublic() {
        return this.setPublic(!this._public)
    }

    public async setPublic(isPublic: boolean) {
        if (this._public === isPublic) {
            return
        }

        this._public = isPublic
        await this.updateShare()
    }

    public async toggleTimelineOnly() {
        return this.setTimelineOnly(!this.timelineOnly)
    }

    public async setTimelineOnly(timelineOnly: boolean) {
        if (this.timelineOnly === timelineOnly) {
            return
        }

        this.timelineOnly = timelineOnly
        await this.updateShare()
    }
}
