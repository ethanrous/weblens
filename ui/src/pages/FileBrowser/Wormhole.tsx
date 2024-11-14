import './style/fileBrowserStyle.scss'

// const UploadPlaque = ({ wormholeId }: { wormholeId: string }) => {
//     return (
//         <div className="h-[45vh]">
//             <FileButton
//                 onChange={(files) => {
//                     HandleUploadButton(files, wormholeId, true, wormholeId)
//                 }}
//                 accept="file"
//                 multiple
//             >
//                 {(props) => {
//                     return (
//                         <div className="flex bg-bottom-grey h-[20vh] w-[20vw] p-3 rounded justify-center">
//                             <div
//                                 className="cursor-pointer h-max w-max"
//                                 onClick={() => {
//                                     props.onClick()
//                                 }}
//                             >
//                                 <IconUpload
//                                     size={100}
//                                     style={{ padding: '10px' }}
//                                 />
//                                 <Text
//                                     size="20px"
//                                     className="select-none font-semibold"
//                                 >
//                                     Upload
//                                 </Text>
//                                 <Space h={4}></Space>
//                                 <Text size="12px" className="select-none">
//                                     Click or Drop
//                                 </Text>
//                             </div>
//                         </div>
//                     )
//                 }}
//             </FileButton>
//         </div>
//     )
// }

// const WormholeWrapper = ({
//     file,
//     children,
// }: {
//     file: WeblensFile
//     children: ReactNode
// }) => {
//     const [dragging, setDragging] = useState(0)
//     // const [dropSpotRef, setDropSpotRef] = useState<HTMLDivElement>(null)
//     // const handleDrag = useCallback(
//     //     (e: MouseEvent) => {
//     //         e.preventDefault()
//     //         if (e.type === 'dragenter' || e.type === 'dragover') {
//     //             if (!dragging) {
//     //                 setDragging(2)
//     //             }
//     //         } else if (dragging) {
//     //             setDragging(0)
//     //         }
//     //     },
//     //     [dragging]
//     // )
//
//     return (
//         <div className="wormhole-wrapper">
//             <div
//                 // ref={setDropSpotRef}
//                 style={{ position: 'relative', width: '98%', height: '98%' }}
//                 //                    See DirViewWrapper \/
//                 onMouseMove={() => {
//                     if (dragging) {
//                         setTimeout(() => setDragging(0), 10)
//                     }
//                 }}
//             >
//                 <DropSpot
//                     parent={file}
//                     // onDrop={(e: MouseEvent) =>
//                     //     HandleDrop(
//                     //         e.dataTransfer.items,
//                     //         fileId,
//                     //         [],
//                     //         true,
//                     //         wormholeId,
//                     //     )
//                     // }
//                     // dropSpotTitle={wormholeName}
//                     // stopDragging={() => setDragging(DraggingStateT.NoDrag)}
//                     // dropAllowed={validWormhole}
//                     // handleDrag={handleDrag}
//                     // wrapperRef={dropSpotRef}
//                 />
//                 {/* <div className="justify-center" onDragOver={handleDrag}> */}
//                 <div className="justify-center">{children}</div>
//             </div>
//         </div>
//     )
// }

// export default function Wormhole() {
//     const wormholeId = useParams()['*']
//
//     const [file, setFile] = useState<WeblensFile>(null)
//
//     useEffect(() => {
//         if (wormholeId !== '') {
//             GetWormholeInfo(wormholeId)
//                 .then((res) => {
//                     if (res.status !== 200) {
//                         return Promise.reject(new Error(res.statusText))
//                     }
//                     return res.json()
//                 })
//                 .then((v: ShareInfo) => FileApi.getFile(v.fileId))
//                 .then((res) => setFile(new WeblensFile(res.data)))
//                 .catch((r) => {
//                     notifications.show({
//                         title: 'Failed to get wormhole info',
//                         message: String(r),
//                         color: 'red',
//                     })
//                 })
//         }
//     }, [wormholeId])
//     // const valid = Boolean(wormholeInfo)
//
//     return (
//         <div>
//             <UploadStatus />
//             <WormholeWrapper
//                 file={file}
//                 // wormholeId={wormholeId}
//                 // wormholeName={wormholeInfo?.shareName}
//                 // fileId={wormholeInfo?.fileId}
//                 // validWormhole={valid}
//             >
//                 <div className="flex flex-row h-[20vh] w-max items-center">
//                     <div className="h-max w-max">
//                         <Text size="40" style={{ lineHeight: '40px' }}>
//                             {valid ? 'Wormhole to' : 'Wormhole not found'}
//                         </Text>
//                         {!valid && (
//                             <Text size="20" style={{ lineHeight: '40px' }}>
//                                 {'Wormhole does not exist or was closed'}
//                             </Text>
//                         )}
//                     </div>
//                     {valid && (
//                         <IconFolder size={40} style={{ marginLeft: '7px' }} />
//                     )}
//                     <Text
//                         fw={700}
//                         size="40"
//                         style={{ lineHeight: '40px', marginLeft: 3 }}
//                     >
//                         {wormholeInfo?.shareName}
//                     </Text>
//                 </div>
//                 {valid && <UploadPlaque wormholeId={wormholeId} />}
//             </WormholeWrapper>
//         </div>
//     )
// }
