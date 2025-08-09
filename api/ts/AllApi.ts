import type { Configuration } from './generated/configuration'

import {
    FilesApiFactory,
    FolderApiFactory,
    MediaApiFactory,
    ShareApiFactory,
    TowersApiFactory,
    UsersApiFactory,
} from './generated/api'

export * from './generated/api'

export type WLAPI = {
    MediaApi: ReturnType<typeof MediaApiFactory>
    FilesApi: ReturnType<typeof FilesApiFactory>
    FoldersApi: ReturnType<typeof FolderApiFactory>
    TowersApi: ReturnType<typeof TowersApiFactory>
    SharesApi: ReturnType<typeof ShareApiFactory>
    UsersApi: ReturnType<typeof UsersApiFactory>
}

export function WeblensApiFactory(apiEndpoint: string): WLAPI {
	return {
		MediaApi: MediaApiFactory({} as Configuration, apiEndpoint),
		FilesApi: FilesApiFactory({} as Configuration, apiEndpoint),
		FoldersApi: FolderApiFactory({} as Configuration, apiEndpoint),
		TowersApi: TowersApiFactory({} as Configuration, apiEndpoint),
		SharesApi: ShareApiFactory({} as Configuration, apiEndpoint),
		UsersApi: UsersApiFactory({} as Configuration, apiEndpoint),
	}
}
