import {
    IconClipboard,
    IconClockHour4,
    IconFile,
    IconReload,
    IconRestore,
    IconTrash,
    IconX,
} from '@tabler/icons-react'
import { ServersApi } from '@weblens/api/ServersApi'
import { ServerInfo } from '@weblens/api/swag'
import { useTimer } from '@weblens/components/hooks'
import WeblensButton from '@weblens/lib/WeblensButton'
import WeblensInput from '@weblens/lib/WeblensInput'
import WeblensProgress from '@weblens/lib/WeblensProgress'
import WeblensTooltip from '@weblens/lib/WeblensTooltip'
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
import WebsocketStatus from './filebrowser/websocketStatus'
import './remoteStatus.scss'

export default function RemoteStatus({
    remoteInfo,
    refetchRemotes,
    restoreProgress,
    backupProgress,
    setBackupProgress,
}: {
    remoteInfo: ServerInfo
    refetchRemotes: () => void
    restoreProgress: RestoreProgress
    backupProgress: BackupProgressT
    setBackupProgress: (p: BackupProgressT) => void
}) {
    const [restoring, setRestoring] = useState(false)
    const [restoreUrl, setRestoreUrl] = useState(remoteInfo.coreAddress)
    const { elapsedTime } = useTimer(restoreProgress?.timestamp)

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
                            squareSize={30}
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
            className="remote-box-container"
            data-restoring={restoring}
            data-backup={backupProgress !== null}
        >
            <div className="remote-box">
                <div className="flex flex-col gap-1">
                    <div className="flex items-center gap-1">
                        <WeblensTooltip
                            label={
                                <div className="flex flex-row">
                                    <IconClipboard />
                                    <p>{remoteInfo.id}</p>
                                </div>
                            }
                        >
                            <h3
                                className="theme-text-dark-bg font-semibold select-none cursor-pointer truncate"
                                onClick={() => {
                                    navigator.clipboard
                                        .writeText(remoteInfo.id)
                                        .catch(ErrorHandler)
                                }}
                            >
                                {remoteInfo.name}
                            </h3>
                        </WeblensTooltip>
                        <WebsocketStatus ready={remoteInfo.online ? 1 : -1} />
                    </div>
                    <div className="flex gap-1 h-max">
                        <IconClockHour4 />
                        <p className="theme-text-dark-bg">
                            {remoteInfo.lastBackup
                                ? historyDateTime(remoteInfo.lastBackup)
                                : 'Never'}
                        </p>
                        {remoteInfo.backupSize != -1 && (
                            <div className="flex">
                                <div className="w-[1px] bg-[--wl-outline-subtle] h-min-1 m-1" />
                                <p className="text-white">
                                    {humanFileSize(remoteInfo.backupSize)}
                                </p>
                            </div>
                        )}
                    </div>
                </div>
                {roleMismatch && (
                    <p className="text-red-500">
                        Server is not initialized - Recommend restore
                    </p>
                )}
                <div className="flex flex-row max-w-full">
                    <WeblensButton
                        squareSize={40}
                        labelOnHover
                        tooltip={canSync ? 'Sync now' : 'Sync unavailable'}
                        Left={IconReload}
                        disabled={!canSync}
                        onClick={async () => {
                            return ServersApi.launchBackup(remoteInfo.id).catch(
                                ErrorHandler
                            )
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
                        onClick={() => {
                            ServersApi.deleteRemote(remoteInfo.id)
                                .then(() => refetchRemotes())
                                .catch(ErrorHandler)
                        }}
                    />
                </div>
            </div>
            {restoring && (
                <div className="restore-dialogue">
                    <div className="flex flex-col w-[50%] justify-around items-center">
                        <div className="flex flex-col w-[50%] ">
                            <p className="text-2xl m-2">Restore Target</p>
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
                                ServersApi.restoreCore(remoteInfo.id, {
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
                <div className="flex flex-col mt-4 pl-4 gap-2 border-l-2 border-wl-outline-subtle">
                    <div className="flex flex-col gap-1 mb-2">
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
                        <div className="flex w-full m-1 items-center gap-2">
                            <WeblensProgress
                                value={
                                    (backupProgress.progressCurrent /
                                        backupProgress.progressTotal) *
                                    100
                                }
                            />
                            <p className="text-nowrap text-white p-1">
                                {backupProgress.progressCurrent} /{' '}
                                {backupProgress.progressTotal} files
                            </p>
                        </div>
                    )}
                    {backupProgress.files.size > 0 && (
                        <div className="flex flex-col gap-1 my-2">
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
            className="backup-stage-box"
            data-complete={Boolean(stage.finished)}
            data-failed={Boolean(stage.started && !stage.finished && error)}
        >
            <p className="truncate">{stage.name}</p>
            {Boolean(stage.started) && !stage.finished && !error && (
                <p className="text-nowrap">
                    {nsToHumanTime(elapsedTime * 1000000)}
                </p>
            )}
            {Boolean(stage.started) && !stage.finished && error && (
                <p className="text-nowrap">Failed</p>
            )}
            {Boolean(stage.finished) && (
                <p className="truncate">
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
