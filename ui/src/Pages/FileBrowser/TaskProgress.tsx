import { useMemo } from "react";
import { FBDispatchT, ScanMeta } from "../../types/Types";
import { nsToHumanTime } from "../../util";
import { Box, Text } from "@mantine/core";
import { ColumnBox, RowBox } from "./FileBrowserStyles";
import { IconX } from "@tabler/icons-react";
import { WeblensProgress } from "../../components/WeblensProgress";

export enum TaskStage {
    Queued,
    InProgress,
    Complete,
}

export class TaskProgress {
    taskId: string;
    taskType: string;
    target: string;
    workingOn: string;
    note: string;

    timeNs: number;
    progressPercent: number;
    tasksComplete: number;
    tasksTotal: number;

    stage: TaskStage;

    hidden: boolean;

    constructor(taskId: string, taskType: string, target: string) {
        console.log("Creating task prog...");

        if (!taskId || !target || !taskType) {
            console.error(
                "TaskId:",
                taskId,
                "Target:",
                target,
                "Task Type:",
                taskType
            );
            throw new Error("Empty prop in TaskProgress constructor");
        }

        this.taskId = taskId;
        this.taskType = taskType;
        this.target = target;

        this.timeNs = 0;
        this.progressPercent = 0;

        this.stage = TaskStage.Queued;

        this.hidden = false;
    }

    GetTaskId(): string {
        return this.taskId;
    }

    GetTaskType(): string {
        switch (this.taskType) {
            case "scan_directory":
                return "Scan Folder";
        }
    }

    getTime(): string {
        return nsToHumanTime(this.timeNs);
    }

    getProgress(): number {
        if (this.stage === TaskStage.Complete) {
            return 100;
        }
        return this.progressPercent;
    }

    hide(): void {
        this.hidden = true;
    }
}

export const TasksDisplay = ({
    scanProgress,
    dispatch,
}: {
    scanProgress: TaskProgress[];
    dispatch: FBDispatchT;
}) => {
    if (scanProgress.length == 0) {
        return null;
    }
    const cards = useMemo(() => {
        return scanProgress.map((sp) => {
            if (sp.hidden) {
                return null;
            }
            return (
                <TaskProgCard key={sp.taskId} prog={sp} dispatch={dispatch} />
            );
        });
    }, [scanProgress]);
    return <ColumnBox style={{ height: "max-content" }}>{cards}</ColumnBox>;
};

const TaskProgCard = ({
    prog,
}: {
    prog: TaskProgress;
    dispatch: FBDispatchT;
}) => {
    console.log(prog.stage);
    return (
        <Box className="task-progress-box">
            <RowBox style={{ height: "max-content" }}>
                <Box style={{ width: "100%" }}>
                    <Text size="12px" style={{ userSelect: "none" }}>
                        {prog.GetTaskType()}
                    </Text>
                    <Text size="16px" fw={600} style={{ userSelect: "none" }}>
                        {prog.target}
                    </Text>
                </Box>
                <IconX
                    size={20}
                    cursor={"pointer"}
                    onClick={
                        () => prog.hide()
                        // dispatch({
                        //     type: "remove_task_progress",
                        //     taskId: prog.taskId,
                        // })
                    }
                />
            </RowBox>
            <Box
                style={{ height: 25, flexShrink: 0, width: "100%", margin: 10 }}
            >
                <WeblensProgress
                    value={prog.getProgress()}
                    complete={prog.stage === TaskStage.Complete}
                    loading={prog.stage === TaskStage.Queued}
                />
            </Box>
            {prog.stage !== TaskStage.Complete && (
                <RowBox
                    style={{
                        justifyContent: "space-between",
                        height: "max-content",
                        gap: 10,
                    }}
                >
                    <Text
                        size="10px"
                        truncate="end"
                        style={{ userSelect: "none" }}
                    >
                        {prog.workingOn}
                    </Text>
                    {prog.tasksTotal > 0 && (
                        <Text size="10px" style={{ userSelect: "none" }}>
                            {prog.tasksComplete}/{prog.tasksTotal}
                        </Text>
                    )}
                </RowBox>
            )}
            <RowBox
                style={{
                    justifyContent: "space-between",
                    height: "max-content",
                    gap: 10,
                }}
            >
                {prog.stage === TaskStage.Complete && (
                    <Text
                        size="10px"
                        style={{ width: "max-content", userSelect: "none" }}
                    >
                        Finished{" "}
                        {prog.timeNs !== 0 ? `in ${prog.getTime()}` : ""}
                    </Text>
                )}
                {prog.stage === TaskStage.Queued && (
                    <Text
                        size="10px"
                        style={{ width: "max-content", userSelect: "none" }}
                    >
                        Queued...
                    </Text>
                )}
                {/* <Text
                    size="10px"
                    style={{
                        width: "max-content",
                        textWrap: "nowrap",
                        userSelect: "none",
                    }}
                >
                    {prog.note}
                </Text> */}
            </RowBox>
        </Box>
    );
};
