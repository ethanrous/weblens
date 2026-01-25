import type { UserInfo } from '@ethanrous/weblens-api'
import { Optional } from '~/util/option'

export enum UserPermissions {
    Public = 0,
    Basic = 1,
    Admin = 2,
    Owner = 3,
    System = 4,
}

export default class User implements UserInfo {
    fullName: string = ''
    homeID: string = ''
    homeSize: number = 0
    permissionLevel: number = 0
    token?: string
    trashID: string = ''
    trashSize: number = 0
    username: string = ''
    activated: boolean = false
    updatedAt: number = 0

    isLoggedIn: Optional<boolean>

    constructor(info?: UserInfo, isLoggedIn?: boolean) {
        if (info) {
            Object.assign(this, info)
        }

        this.isLoggedIn = new Optional(isLoggedIn)
    }

    public GetPermissionLevel(): UserPermissions {
        return this.permissionLevel as UserPermissions
    }

    public static GetPermissionLevelName(level: UserPermissions): string {
        switch (level) {
            case UserPermissions.Public:
                return 'Public'
            case UserPermissions.Basic:
                return 'Basic'
            case UserPermissions.Admin:
                return 'Admin'
            case UserPermissions.Owner:
                return 'Owner'
            case UserPermissions.System:
                return 'System'
            default:
                return 'Unknown'
        }
    }
}
