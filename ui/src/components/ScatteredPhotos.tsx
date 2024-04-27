import { Box } from "@mantine/core";
import { getRandomInt } from "../util";
import { useEffect, useMemo, useState } from "react";
import { WeblensMedia } from "../classes/Media";
import { getRandomThumbs, newApiKey } from "../api/ApiFetch";
import { useResize } from "./hooks";
import { MediaImage } from "./PhotoContainer";

const ScatteredPhoto = ({
    media,
    attribute,
}: {
    media: WeblensMedia;
    attribute: attribute;
}) => {
    const [mData, setMData] = useState(media);

    return (
        <Box
            className="scattered-photo"
            style={{
                width: attribute.width,
                // transform: `translate(${attribute.top}px, ${attribute.left}px)`,
                left: attribute.left,
                top: attribute.top,
                rotate: `${attribute.rotate}deg`,
                zIndex: 60 - Math.floor(attribute.blur * 10),
            }}
        >
            <MediaImage
                media={mData}
                quality="thumbnail"
                imgStyle={{ filter: `blur(${attribute.blur}px)` }}
                doPublic
                setMediaCallback={(id, quality, data) => {
                    setMData((p) => {
                        if (quality == "thumbnail") {
                            p.thumbnail = data;
                        } else {
                            p.fullres = data;
                        }
                        return { ...p };
                    });
                }}
            />
        </Box>
    );
};

type attribute = {
    width: number;
    left: number;
    top: number;
    edgeRight: number;
    edgeBottom: number;
    blur: number;
    rotate: number;
};

const doesCollide = (box1: attribute, box2: attribute) => {
    let collideH: boolean = false;
    let collideV: boolean = false;
    if (
        (box1.left >= box2.left && box1.left <= box2.edgeRight) ||
        (box2.left >= box1.left && box2.left <= box1.edgeRight)
    ) {
        collideH = true;
    }

    if (
        (box1.top >= box2.top && box1.top <= box2.edgeBottom) ||
        (box2.top >= box1.top && box2.top <= box1.edgeBottom)
    ) {
        collideV = true;
    }
    return collideH && collideV;
};

export const ScatteredPhotos = () => {
    const [medias, setMedias]: [medias: WeblensMedia[], setMedias: any] =
        useState([]);
    useEffect(() => {
        getRandomThumbs().then((r) => setMedias(r.medias));
    }, []);
    const [pageRef, setPageRef] = useState(null);
    const pageSize = useResize(pageRef);

    const attributes = useMemo(() => {
        let attributes: attribute[] = [];
        attributes.push({
            blur: 0,
            rotate: 0,
            width: 500,
            left: pageSize.width / 2 - 300,
            top: pageSize.height / 2 - 300,
            edgeRight: pageSize.width / 2 + 300,
            edgeBottom: pageSize.height / 2 + 300,
        });
        for (const m of medias) {
            // const blur = getRandomInt(1, 3);
            // const rotate = 15;
            const rotate = getRandomInt(0, 20);

            const newAttr: attribute = {
                blur: 0,
                rotate: rotate - 10,
                width: 0,
                left: 0,
                top: 0,
                edgeRight: 0,
                edgeBottom: 0,
            };
            const longOrig =
                m.mediaWidth > m.mediaHeight ? m.mediaWidth : m.mediaHeight;
            const shortOrig =
                m.mediaWidth > m.mediaHeight ? m.mediaHeight : m.mediaWidth;
            let longRatio;
            let maxSize = 500;
            while (maxSize > 150) {
                longRatio = getRandomInt(150, maxSize);
                const shortRatio = (longRatio / longOrig) * shortOrig;

                newAttr.width =
                    m.mediaWidth > m.mediaHeight ? longRatio : shortRatio;
                const height =
                    m.mediaWidth < m.mediaHeight ? longRatio : shortRatio;

                newAttr.left = getRandomInt(0, pageSize.width - newAttr.width);
                newAttr.top = getRandomInt(0, pageSize.height - height);
                newAttr.edgeRight = newAttr.left + newAttr.width;
                newAttr.edgeBottom = newAttr.top + height;

                let colision: attribute;
                for (const past of attributes) {
                    if (doesCollide(past, newAttr)) {
                        colision = past;
                        break;
                    }
                }
                if (!colision) {
                    break;
                }

                if (colision.width === 500) {
                    continue;
                }
                maxSize--;
            }
            newAttr.blur = ((longRatio - 149) / 250) * (4 - 1);

            attributes.push(newAttr);
        }
        return attributes;
    }, [medias]);

    return (
        <Box
            ref={setPageRef}
            style={{
                position: "absolute",
                width: "100vw",
                height: "100vh",
                zIndex: 0,
            }}
        >
            {medias.map((m, i) => {
                return (
                    <ScatteredPhoto
                        key={m.mediaId}
                        media={m}
                        attribute={attributes[i + 1]}
                    />
                );
            })}
        </Box>
    );
};
