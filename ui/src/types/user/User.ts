import { UserInfo } from '@weblens/api/swag'

export default class User {
    activated?: boolean
    admin?: boolean
    homeId?: string
    isSystemUser?: boolean
    owner?: boolean
    password?: string
    trashId?: string
    username?: string

    isLoggedIn: boolean

    constructor(info?: UserInfo, isLoggedIn?: boolean) {
        if (info) {
            Object.assign(this, info)
        }

        this.isLoggedIn = isLoggedIn || false
    }
}
