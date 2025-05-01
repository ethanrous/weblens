import { UserInfo } from '@weblens/api/swag'

export enum UserPermissions {
	Public = 0,
	Basic = 1,
	Admin = 2,
	Owner = 3,
	System = 4,
}

export default class User implements UserInfo {
	fullName: string;
	homeId: string;
	homeSize: number;
	permissionLevel: number;
	token?: string;
	trashId: string;
	trashSize: number;
	username: string;
	activated: boolean;

	isLoggedIn: boolean

	constructor(info?: UserInfo, isLoggedIn?: boolean) {
		if (info) {
			Object.assign(this, info)
		}

		this.isLoggedIn = isLoggedIn || false
	}
}
