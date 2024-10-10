import {
    IconClockHour4,
    IconFile,
    IconReload,
    IconRestore,
    IconTrash,
} from '@tabler/icons-react'
import { deleteRemote, doBackup } from '@weblens/api/ApiFetch'
import WeblensButton from '@weblens/lib/WeblensButton'
import { WebsocketStatus } from '@weblens/pages/FileBrowser/FileBrowserMiscComponents'
import { ServerInfoT } from '@weblens/types/Types'
import './remoteStatus.scss'
import { useState } from 'react'
import WeblensInput from '@weblens/lib/WeblensInput'
import { launchRestore } from '@weblens/api/SystemApi'
import {
    BackupProgress,
    RestoreProgress,
} from '@weblens/pages/Backup/BackupLogic'
import WeblensProgress from '@weblens/lib/WeblensProgress'
import { useTimer } from '@weblens/components/hooks'
import { humanFileSize, nsToHumanTime } from '@weblens/util'
import './theme.scss'
import { historyDate } from '@weblens/pages/FileBrowser/FileBrowserLogic'
import WeblensTooltip from '@weblens/lib/WeblensTooltip'

export default function RemoteStatus({
    remoteInfo,
    refetchRemotes,
    restoreProgress,
    backupProgress,
}: {
    remoteInfo: ServerInfoT
    refetchRemotes: () => void
    restoreProgress: RestoreProgress
    backupProgress: BackupProgress
}) {
    const [restoring, setRestoring] = useState(false)
    const [restoreUrl, setRestoreUrl] = useState(remoteInfo.coreAddress)
    const { elapsedTime } = useTimer(restoreProgress?.timestamp)

    const roleMismatch =
        remoteInfo.role === 'core' &&
        (remoteInfo.reportedRole === '' ||
            remoteInfo.reportedRole !== remoteInfo.role)
    const canSync = remoteInfo.online && !roleMismatch && (!backupProgress || backupProgress.error !== null)

    return (
        <div
            key={remoteInfo.id}
            className="remote-box-container"
            data-restoring={restoring}
            data-backup={backupProgress !== null}
        >
            <div className="remote-box">
                <div className="flex flex-col gap-1">
                    <div className="flex flex-row items-center gap-1">
                        <WeblensTooltip
                            label={`Remote Id (Click to Copy): ${remoteInfo.id}`}
                        >
                            <p
                                className="theme-text-dark-bg font-semibold select-none cursor-pointer truncate"
                                onClick={() =>
                                    navigator.clipboard.writeText(remoteInfo.id)
                                }
                            >
                                {remoteInfo.name} ({remoteInfo.role})
                            </p>
                        </WeblensTooltip>
                        <WebsocketStatus ready={remoteInfo.online ? 1 : -1} />
                    </div>
                    <div className="flex gap-1 h-max">
                        <IconClockHour4 />
                        <p className="theme-text-dark-bg">
                            {remoteInfo.lastBackup
                                ? historyDate(remoteInfo.lastBackup)
                                : 'Never'}
                        </p>
                        <div className='w-[1px] bg-[--wl-outline-subtle] h-min-1 m-1'/>
                        <p>{humanFileSize(remoteInfo.backupSize) }</p>
                    </div>
                </div>
                {roleMismatch && (
                    <p className="text-red-500">
                        Server is not initialized - Recommend restore
                    </p>
                )}
                <div className="flex flex-row max-w-full">
                    <WeblensButton
                        label="Sync now"
                        squareSize={40}
                        labelOnHover
                        tooltip={!canSync ? 'Sync unavailable' : ''}
                        Left={IconReload}
                        disabled={!canSync}
                        onClick={async () => {
                            const res = await doBackup(remoteInfo.id)
                            return res.status === 200
                        }}
                    />
                    {remoteInfo.role === 'core' && (
                        <WeblensButton
                            label="Restore"
                            tooltip={
                                canSync ? 'Server is already initialized' : ''
                            }
                            labelOnHover
                            Left={IconRestore}
                            squareSize={40}
                            disabled={canSync}
                            onClick={() => setRestoring((p) => !p)}
                        />
                    )}
                    <WeblensButton
                        Left={IconTrash}
                        danger
                        onClick={async () => {
                            deleteRemote(remoteInfo.id).then(() =>
                                refetchRemotes()
                            )
                        }}
                    />
                </div>
            </div>
            {restoring && (
                <div className="restore-diologue">
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
                                launchRestore(remoteInfo, restoreUrl)
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
                            {restoreProgress.progress_total && (
                                <WeblensProgress
                                    value={
                                        (restoreProgress.progress_current /
                                            restoreProgress.progress_total) *
                                        100
                                    }
                                />
                            )}
                        </div>
                    )}
                </div>
            )}
            {backupProgress && backupProgress.error && (
                <div className="pt-4">
                    <h4 className="text-lg font-semibold text-red-500">
                        Backup Failed
                    </h4>
                    <p className="p-2 theme-text-dark-bg outline outline-red-500 rounded">
                        {backupProgress.error}
                    </p>
                </div>
            )}
            {backupProgress && !backupProgress.error && (
                <div className="pt-4">
                    <p>{backupProgress.stage}</p>
                    {backupProgress.progress_total && (
                        <WeblensProgress
                            value={
                                (backupProgress.progress_current /
                                    backupProgress.progress_total) *
                                100
                            }
                        />
                    )}
                    {backupProgress.files.size > 0 && (
                        <div className="flex flex-col gap-1 my-2">
                            {Array.from(backupProgress.files).map(
                                ([name, { start }]) => (
                                    <BackupFile
                                        key={name}
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

function BackupFile({ name, start }) {
    const { elapsedTime } = useTimer(start)
    return (
        <div className="flex flex-row gap-2">
            <IconFile />
            <p className="min-w-48">{nsToHumanTime(elapsedTime * 1000000)}</p>
            <p>{name}</p>
        </div>
    )
}
