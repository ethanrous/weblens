import { useQuery } from '@tanstack/react-query'
import { useFileBrowserStore } from '@weblens/store/FBStateControl'
import { ErrorHandler } from '@weblens/types/Types'
import WeblensFile from '@weblens/types/files/File'
import { WeblensShare } from '@weblens/types/share/share'

export default function useShare(target?: WeblensFile) {
    const folderInfo = useFileBrowserStore((state) => state.folderInfo)

    let _target = target
    if (!_target) {
        _target = folderInfo
    }

    const {
        data: share,
        refetch,
        isLoading,
    } = useQuery<WeblensShare>({
        queryKey: ['share', _target.Id()],
        queryFn: async () => {
            console.log('Fetching share for', _target.Id())
            const share = await _target.GetShare(true).catch(ErrorHandler)
            if (!share) {
                return null
            }

            return share
        },
    })

    return {
        share,
        refetchShare: refetch,
        shareLoading: isLoading,
    }
}
