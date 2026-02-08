import type { UserInfo } from '@ethanrous/weblens-api'
import { Optional } from '~/util/option'

export enum UserPermissions {
    PUBLIC = 0,
    BASIC = 1,
    ADMIN = 2,
    OWNER = 3,
    SYSTEM = 4,
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
            case UserPermissions.PUBLIC:
                return 'Public'
            case UserPermissions.BASIC:
                return 'Basic'
            case UserPermissions.ADMIN:
                return 'Admin'
            case UserPermissions.OWNER:
                return 'Owner'
            case UserPermissions.SYSTEM:
                return 'System'
            default:
                return 'Unknown'
        }
    }
}
