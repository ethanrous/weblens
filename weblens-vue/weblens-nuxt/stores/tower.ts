import { defineStore } from 'pinia'
import { useWeblensAPI } from '~/api/AllApi'

export const useTowerStore = defineStore('tower', () => {
    const { data: towerInfo } = useAsyncData(
        'tower',
        async () => {
            const towerRes = await useWeblensAPI().TowersAPI.getServerInfo()
            return towerRes.data
        },
        { immediate: true, lazy: false },
    )

    return { towerInfo }
})
