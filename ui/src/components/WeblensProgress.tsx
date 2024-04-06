import { Box, CSSProperties } from "@mantine/core";
import { memo } from "react";
import "../style/weblensProgress.css";

type progressProps = {
    value: number;
    color: string;
    orientation?: "horizontal" | "vertical";
    style?: CSSProperties;
};

export const WeblensProgress = memo(
    ({ value, color, orientation = "horizontal", style }: progressProps) => {
        return (
            <Box
                className="weblens-progress"
                style={{
                    justifyContent: orientation === "horizontal" ? "flex-start" : "flex-end",
                    flexDirection: orientation === "horizontal" ? "row" : "column",
                    ...style,
                }}
            >
                <Box
                    className="weblens-progress-bar"
                    style={{
                        height: orientation === "horizontal" ? "100%" : `${value}%`,
                        width: orientation === "horizontal" ? `${value}%` : "100%",
                        backgroundColor: color,
                    }}
                ></Box>
            </Box>
        );
    },
    (prev, next) => {
        if (prev.value !== next.value) {
            return false;
        }
        if (prev.color !== next.color) {
            return false;
        }
        if (prev.orientation !== next.orientation) {
            return false;
        }
        return true;
    },
);
