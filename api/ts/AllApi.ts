import type { Configuration } from "./generated/configuration";

import {
  APIKeysApiFactory,
  FeatureFlagsApiFactory,
  FilesApiFactory,
  FolderApiFactory,
  MediaApiFactory,
  ShareApiFactory,
  TowersApiFactory,
  UsersApiFactory,
} from "./generated/api";

export * from "./generated/api";

export type WLAPI = {
  MediaAPI: ReturnType<typeof MediaApiFactory>;
  FilesAPI: ReturnType<typeof FilesApiFactory>;
  FoldersAPI: ReturnType<typeof FolderApiFactory>;
  TowersAPI: ReturnType<typeof TowersApiFactory>;
  SharesAPI: ReturnType<typeof ShareApiFactory>;
  UsersAPI: ReturnType<typeof UsersApiFactory>;
  APIKeysAPI: ReturnType<typeof APIKeysApiFactory>;
  FeatureFlagsAPI: ReturnType<typeof FeatureFlagsApiFactory>;
};

export function WeblensAPIFactory(apiEndpoint: string): WLAPI {
  return {
    MediaAPI: MediaApiFactory({} as Configuration, apiEndpoint),
    FilesAPI: FilesApiFactory({} as Configuration, apiEndpoint),
    FoldersAPI: FolderApiFactory({} as Configuration, apiEndpoint),
    TowersAPI: TowersApiFactory({} as Configuration, apiEndpoint),
    SharesAPI: ShareApiFactory({} as Configuration, apiEndpoint),
    UsersAPI: UsersApiFactory({} as Configuration, apiEndpoint),
    APIKeysAPI: APIKeysApiFactory({} as Configuration, apiEndpoint),
    FeatureFlagsAPI: FeatureFlagsApiFactory({} as Configuration, apiEndpoint),
  };
}
