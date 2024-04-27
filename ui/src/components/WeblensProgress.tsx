import { Box, CSSProperties } from "@mantine/core";
import { memo } from "react";
import "../style/weblensProgress.scss";

type progressProps = {
    value: number;
    complete?: boolean;
    orientation?: "horizontal" | "vertical";
    loading?: boolean;
    style?: CSSProperties;
};

export const WeblensProgress = memo(
    ({
        value,
        complete = false,
        orientation = "horizontal",
        loading = false,
        style,
    }: progressProps) => {
        return (
            <Box
                className="weblens-progress"
                mod={{
                    "data-loading": Boolean(loading).toString(),
                    "data-complete": Boolean(complete),
                }}
                style={{
                    justifyContent:
                        orientation === "horizontal"
                            ? "flex-start"
                            : "flex-end",
                    flexDirection:
                        orientation === "horizontal" ? "row" : "column",
                    ...style,
                }}
            >
                <Box
                    className="weblens-progress-bar"
                    mod={{
                        "data-complete": Boolean(complete),
                    }}
                    style={{
                        height:
                            orientation === "horizontal" ? "100%" : `${value}%`,
                        width:
                            orientation === "horizontal" ? `${value}%` : "100%",
                    }}
                />
            </Box>
        );
    },
    (prev, next) => {
        if (prev.value !== next.value) {
            return false;
        }
        if (prev.complete !== next.complete) {
            return false;
        }
        if (prev.orientation !== next.orientation) {
            return false;
        }
        return true;
    }
);
