import {
    IconArrowRight,
    IconCheck,
    IconClipboard,
    IconClockHour4,
    IconFile,
    IconReload,
    IconRestore,
    IconServer,
    IconTrash,
    IconX,
} from '@tabler/icons-react'
import { TowersApi } from '@weblens/api/ServersApi'
import { TowerInfo } from '@weblens/api/swag'
import LoaderDots from '@weblens/lib/LoaderDots'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import WeblensProgress from '@weblens/lib/WeblensProgress'
import WeblensTooltip from '@weblens/lib/WeblensTooltip'
import { useTimer } from '@weblens/lib/hooks'
import {
    BackupProgressT,
    RestoreProgress,
} from '@weblens/pages/Backup/BackupLogic'
import { historyDateTime } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import { TaskStageT } from '@weblens/store/TaskStateControl'
import { ErrorHandler } from '@weblens/types/Types'
import { humanFileSize, nsToHumanTime } from '@weblens/util'
import { useEffect, useMemo, useState } from 'react'

import WeblensLoader from './Loading'
import { useSessionStore } from './UserInfo'
import { WebsocketStatusCard } from './filebrowser/websocketStatus'

export default function RemoteStatus({
    remoteInfo,
    refetchRemotes,
    restoreProgress,
    backupProgress,
    setBackupProgress,
}: {
    remoteInfo: TowerInfo
    refetchRemotes: () => void
    restoreProgress: RestoreProgress
    backupProgress: BackupProgressT
    setBackupProgress: (p: BackupProgressT) => void
}) {
    const [restoring, setRestoring] = useState(false)
    const [newApiKey, setNewApiKey] = useState('')
    const [restoreUrl, setRestoreUrl] = useState(remoteInfo.coreAddress)
    const { elapsedTime } = useTimer(restoreProgress?.timestamp)

    const serverInfo = useSessionStore((state) => state.server)

    const roleMismatch =
        remoteInfo.role === 'core' &&
        (!remoteInfo.reportedRole ||
            remoteInfo.reportedRole !== remoteInfo.role)
    const canSync =
        remoteInfo.online &&
        !roleMismatch &&
        (!backupProgress || backupProgress.error !== null)

    const backupHeaderText = useMemo(() => {
        if (backupProgress) {
            if (backupProgress.error) {
                return (
                    <div>
                        <h4 className="text-red-500">Backup Failed</h4>
                        <p className="text-red-500">{backupProgress.error}</p>
                    </div>
                )
            } else if (backupProgress.totalTime) {
                return (
                    <div className="flex items-center gap-2">
                        <h4>Backup Complete</h4>
                        <p className="text-green-500">
                            {nsToHumanTime(backupProgress.totalTime)}
                        </p>
                        <div className="flex grow" />
                        <WeblensButton
                            Left={IconX}
                            tooltip="Close"
                            size="small"
                            flavor="outline"
                            onClick={() => setBackupProgress(null)}
                        />
                    </div>
                )
            } else {
                return (
                    <div className="flex items-center gap-2">
                        <h4 className="text-white">Backup In Progress</h4>
                        <WeblensLoader />
                    </div>
                )
            }
        }
        return ''
    }, [backupProgress?.error, backupProgress?.totalTime])

    return (
        <div
            key={remoteInfo.id}
            className="bg-background-secondary data-[restoring=true]-h-[400px] flex w-full flex-col rounded-md border p-2 transition"
            data-restoring={restoring}
            data-backup={backupProgress !== null}
        >
            <div className="mb-2 flex flex-row items-center justify-between">
                <span className="flex w-max items-center gap-1 text-lg">
                    <IconServer />
                    {remoteInfo.name}
                </span>
                {remoteInfo.id !== serverInfo.id && (
                    <WebsocketStatusCard ready={remoteInfo.online ? 1 : -1} />
                )}
                {remoteInfo.id === serverInfo.id && (
                    <div className="bg-amethyst-600 flex w-max items-center gap-2 rounded p-1">
                        <span className="text-xs">This Tower</span>
                    </div>
                )}
            </div>

            <div className="flex w-full gap-6">
                <div className="flex flex-col">
                    <span className="text-text-secondary flex items-center gap-1 text-nowrap">
                        <IconClockHour4 size={14} />
                        Last Backup
                    </span>
                    <div className="flex h-max gap-1">
                        <span className="theme-text-dark-bg text-nowrap">
                            {remoteInfo.lastBackup
                                ? historyDateTime(remoteInfo.lastBackup)
                                : 'Never'}
                        </span>
                    </div>
                </div>

                {remoteInfo.backupSize != -1 && (
                    <div className="flex w-max flex-col">
                        <span className="text-text-secondary text-nowrap">
                            Backup size
                        </span>
                        <span className="text-white">
                            {humanFileSize(remoteInfo.backupSize)}
                        </span>
                    </div>
                )}

                <div className="mr-6 flex flex-col">
                    <span className="text-text-secondary">Tower Id</span>
                    <WeblensTooltip
                        label={
                            <div className="flex flex-row">
                                <IconClipboard />
                                <p>{remoteInfo.id}</p>
                            </div>
                        }
                        className="w-max"
                    >
                        <h5
                            className="cursor-pointer truncate text-center leading-none select-none"
                            onClick={() => {
                                navigator.clipboard
                                    .writeText(remoteInfo.id)
                                    .catch(ErrorHandler)
                            }}
                        >
                            {remoteInfo.id}
                        </h5>
                    </WeblensTooltip>
                </div>

                <div className="flex w-full flex-row">
                    {roleMismatch && (
                        <div className="flex flex-col items-center gap-1">
                            <span className="text-red-500">
                                Server is not initialized
                            </span>
                            <p>Restore recommended</p>
                            <div className="flex gap-2">
                                <WeblensInput
                                    value={newApiKey}
                                    placeholder="New Api Key"
                                    valueCallback={setNewApiKey}
                                />
                                <WeblensButton
                                    Left={IconArrowRight}
                                    onClick={() => {
                                        TowersApi.updateRemote(remoteInfo.id, {
                                            usingKey: newApiKey,
                                        }).catch(ErrorHandler)
                                        setNewApiKey('')
                                    }}
                                />
                            </div>
                        </div>
                    )}
                    <div className="ml-auto flex max-w-full flex-row gap-2">
                        <WeblensButton
                            squareSize={40}
                            labelOnHover
                            tooltip={canSync ? 'Sync now' : 'Sync unavailable'}
                            Left={IconReload}
                            disabled={!canSync}
                            onClick={async () => {
                                return TowersApi.launchBackup(
                                    remoteInfo.id
                                ).catch(ErrorHandler)
                            }}
                        />
                        {remoteInfo.role === 'core' && (
                            <WeblensButton
                                tooltip={
                                    canSync
                                        ? 'Restore'
                                        : 'Server is already initialized'
                                }
                                requireConfirm
                                Left={IconRestore}
                                squareSize={40}
                                disabled={canSync}
                                onClick={() => setRestoring((p) => !p)}
                            />
                        )}
                        <WeblensButton
                            Left={IconTrash}
                            danger
                            requireConfirm
                            tooltip="Delete remote"
                            onClick={() => {
                                TowersApi.deleteRemote(remoteInfo.id)
                                    .then(() => refetchRemotes())
                                    .catch(ErrorHandler)
                            }}
                        />
                    </div>
                </div>
            </div>
            {restoring && (
                <div className="restore-dialogue">
                    <div className="flex w-[50%] flex-col items-center justify-around">
                        <div className="flex w-[50%] flex-col">
                            <p className="m-2 text-2xl">Restore Target</p>
                            <WeblensInput
                                value={restoreUrl}
                                squareSize={40}
                                valueCallback={setRestoreUrl}
                            />
                        </div>
                        <WeblensButton
                            label={`Restore ${remoteInfo.name}`}
                            Left={IconRestore}
                            disabled={canSync}
                            onClick={() =>
                                TowersApi.restoreCore(remoteInfo.id, {
                                    restoreId: remoteInfo.id,
                                    restoreUrl: restoreUrl,
                                })
                            }
                        />
                    </div>
                    {restoreProgress && (
                        <div className="w-[50%]">
                            <p className="text-nowrap">
                                {restoreProgress.stage}
                            </p>
                            {restoreProgress.error && (
                                <p className="text-red-500">
                                    {restoreProgress.error}
                                </p>
                            )}
                            {!restoreProgress.error &&
                                restoreProgress.timestamp && (
                                    <p>
                                        {nsToHumanTime(elapsedTime * 1000000)}
                                    </p>
                                )}
                            {restoreProgress.progressTotal && (
                                <WeblensProgress
                                    value={
                                        (restoreProgress.progressCurrent /
                                            restoreProgress.progressTotal) *
                                        100
                                    }
                                />
                            )}
                        </div>
                    )}
                </div>
            )}
            {backupProgress && (
                <div className="border-wl-outline-subtle mt-4 flex flex-col gap-2 border-l-2 pl-4">
                    <div className="mb-2 flex flex-col gap-1">
                        {backupHeaderText}
                    </div>
                    {backupProgress.stages?.map((s) => (
                        <StageDisplay
                            key={s.name}
                            stage={s}
                            error={backupProgress.error}
                        />
                    ))}
                    {backupProgress.progressTotal && (
                        <div className="m-1 flex w-full items-center gap-2">
                            <WeblensProgress
                                value={
                                    (backupProgress.progressCurrent /
                                        backupProgress.progressTotal) *
                                    100
                                }
                            />
                            <p className="p-1 text-nowrap text-white">
                                {backupProgress.progressCurrent} /{' '}
                                {backupProgress.progressTotal} files
                            </p>
                        </div>
                    )}
                    {backupProgress.files.size > 0 && (
                        <div className="my-2 flex flex-col gap-1">
                            {Array.from(backupProgress.files).map(
                                ([name, { start }]) => (
                                    <BackupFile
                                        key={`backup-file-${name}`}
                                        name={name}
                                        start={start}
                                    />
                                )
                            )}
                        </div>
                    )}
                </div>
            )}
        </div>
    )
}

function StageDisplay({ stage, error }: { stage: TaskStageT; error?: string }) {
    const [startTime, setStartTime] = useState<Date>()
    const { elapsedTime, handlePause, isRunning, handleStart } = useTimer(
        startTime,
        true
    )

    useEffect(() => {
        if (stage.started && !startTime?.getTime()) {
            setStartTime(new Date(stage.started))
        }
    }, [stage.started])

    useEffect(() => {
        if (
            isRunning &&
            (!stage.started || (stage.started && stage.finished))
        ) {
            handlePause()
        } else if (!isRunning && startTime && !stage.finished) {
            handleStart()
        }
    }, [startTime, stage.finished, isRunning])

    return (
        <div
            className="flex flex-row items-center gap-2 rounded-md border p-2"
            data-complete={Boolean(stage.finished)}
            data-failed={Boolean(stage.started && !stage.finished && error)}
        >
            {Boolean(stage.finished) && <IconCheck />}
            <p className="truncate">{stage.name}</p>
            {Boolean(stage.started) && !stage.finished && !error && (
                <>
                    <LoaderDots />
                    <p className="text-nowrap">
                        {nsToHumanTime(elapsedTime * 1000000)}
                    </p>
                </>
            )}
            {Boolean(stage.started) && !stage.finished && error && (
                <p className="text-nowrap">Failed</p>
            )}
            {Boolean(stage.finished) && (
                <p className="ml-auto truncate">
                    {nsToHumanTime((stage.finished - stage.started) * 1000000)}
                </p>
            )}
        </div>
    )
}

function BackupFile({ name, start }: { name: string; start: Date }) {
    const { elapsedTime } = useTimer(start)
    return (
        <div className="flex flex-row gap-2">
            <IconFile />
            <p className="min-w-48 text-white">
                {nsToHumanTime(elapsedTime * 1000000)}
            </p>
            <p className="text-white">{name}</p>
        </div>
    )
}
