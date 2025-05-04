import { FbModeT, useFileBrowserStore } from '@weblens/store/FBStateControl'

import API_ENDPOINT from './ApiEndpoint.js'
import { WsAction, WsSendFunc, WsSubscriptionType, useWebsocketStore } from './Websocket'
import {
	FilesApiAxiosParamCreator,
	FilesApiFactory,
	FolderApiFactory,
	FolderInfo,
} from './swag/api.js'
import WeblensFile from '@weblens/types/files/File.js'

export const FileApi = FilesApiFactory(null, API_ENDPOINT)
export const FolderApi = FolderApiFactory(null, API_ENDPOINT)

export function SubToFolder(subId: string, shareId: string, wsSend: WsSendFunc) {
	if (subId === '') {
		console.trace('Empty subId')
		return
	} else if (subId === 'shared') {
		return
	}

	wsSend({
		action: WsAction.Subscribe, subscriptionType: WsSubscriptionType.Folder, subscribeKey: subId, content: {
			shareId: shareId,
		}
	})
}

export function SubToTask(
	taskId: string,
	lookingFor: string[],
) {
	const wsSend = useWebsocketStore.getState().wsSend

	wsSend({
		action: WsAction.Subscribe, subscriptionType: WsSubscriptionType.Task, subscribeKey: taskId, content: {
			lookingFor: lookingFor,
		}
	})
}

export function ScanDirectory(directory: WeblensFile) {
	const wsSend = useWebsocketStore.getState().wsSend
	const shareId = useFileBrowserStore.getState().shareId

	wsSend({ action: WsAction.ScanDirectory, content: { folderId: directory.Id(), shareId: shareId } })
}

export function CancelTask(taskId: string) {
	const wsSend = useWebsocketStore.getState().wsSend

	wsSend({ action: WsAction.CancelTask, content: { taskId: taskId } })
}

export function UnsubFromFolder(subId: string, wsSend: WsSendFunc) {
	if (!subId || useWebsocketStore.getState().readyState < 1) {
		return
	}
	wsSend({ action: WsAction.Unsubscribe, content: { subscribeKey: subId } })
}

export async function GetFolderData(
	folderId: string,
	fbMode: FbModeT,
	shareId?: string,
	viewingTime?: Date
): Promise<FolderInfo> {
	if (fbMode === FbModeT.share && !shareId) {
		const res = await FileApi.getSharedFiles()
		return res.data
	}
	if (fbMode === FbModeT.external) {
		console.error('External files not implemented')
	}

	const res = await FolderApi.getFolder(
		folderId,
		shareId ? shareId : undefined,
		viewingTime?.getTime(),
		{ withCredentials: true }
	)
	return res.data
}

export async function downloadSingleFile(
	fileId: string,
	filename: string,
	isZip: boolean,
	shareId: string,
	format: 'webp' | 'jpeg' | null
) {
	const a = document.createElement('a')
	const paramCreator = FilesApiAxiosParamCreator()
	const args = await paramCreator.downloadFile(fileId, shareId, format, isZip)
	const url = API_ENDPOINT + args.url

	if (isZip) {
		filename = 'weblens_download_' + filename
	}

	a.href = url
	a.download = filename
	a.click()
}
