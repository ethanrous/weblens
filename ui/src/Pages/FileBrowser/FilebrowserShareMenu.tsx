import { Autocomplete, Box, Space, Text } from '@mantine/core';
import {
    AutocompleteUsers,
    ShareFiles,
    UpdateFileShare,
} from '../../api/FileBrowserApi';
import {
    IconLink,
    IconUser,
    IconUserCancel,
    IconUsersGroup,
    IconX,
} from '@tabler/icons-react';
import { WeblensButton } from '../../components/WeblensButton';

import { useCallback, useContext, useEffect, useState } from 'react';
import { userContext } from '../../Context';
import { ColumnBox, RowBox } from './FilebrowserStyles';
import { AuthHeaderT, FileInfoT, shareData } from '../../types/Types';

export function ShareInput({
    isPublic,
    sharedUsers,
    setSharedUsers,
}: {
    isPublic: boolean;
    sharedUsers: string[];
    setSharedUsers: (v) => void;
}) {
    const { usr, authHeader }: UserContextT = useContext(userContext);
    const [userSearchResult, setUserSearch] = useState([]);
    const [search, setSearch] = useState('');

    const searchUsers = useCallback(
        async (query: string) => {
            if (query.length < 2) {
                setUserSearch([]);
            }

            let users: string[] = await AutocompleteUsers(query, authHeader);
            const selfIndex = users.indexOf(usr.username);
            if (selfIndex !== -1) {
                users.splice(selfIndex, 1);
            }
            setUserSearch(users);
        },
        [usr.username, authHeader]
    );

    const renderAutocompleteOption = useCallback(
        ({ option }) => {
            return (
                <Box
                    style={{ height: '100%', width: '100%' }}
                    onClick={(e) => {
                        e.stopPropagation();
                        setSharedUsers((v) => {
                            v.push(option.value);
                            return [...v];
                        });
                    }}
                >
                    <Text
                        style={{
                            color: sharedUsers.includes(option.value)
                                ? '#555555'
                                : 'white',
                        }}
                    >
                        {option.value}
                    </Text>
                </Box>
            );
        },
        [sharedUsers, setSharedUsers]
    );

    return (
        <ColumnBox style={{ width: 300 }}>
            <Autocomplete
                disabled={isPublic}
                data={userSearchResult.filter((v) => !sharedUsers.includes(v))}
                value={search}
                renderOption={renderAutocompleteOption}
                onChange={(s) => {
                    setSearch(s);
                    searchUsers(s);
                }}
                comboboxProps={{ dropdownPadding: 0 }}
                placeholder="Add people"
                style={{ width: '100%', padding: 8 }}
            />
            <Text
                c={isPublic ? '#777777' : 'white'}
                style={{
                    width: '100%',
                    textAlign: 'center',
                    marginTop: 10,
                    userSelect: 'none',
                }}
            >
                Shared with
            </Text>
            <ColumnBox
                style={{
                    height: 'max-content',
                    minHeight: '33px',
                    backgroundColor: '#00000044',
                    marginTop: 5,
                    paddingTop: '10px',
                    paddingBottom: '10px',
                }}
            >
                {sharedUsers.map((v) => {
                    return (
                        <RowBox
                            key={v}
                            style={{
                                width: '90%',
                                height: '33px',
                                padding: 10,
                            }}
                        >
                            <IconUser color={isPublic ? '#777777' : 'white'} />
                            <Space w={10} />
                            <Text
                                c={isPublic ? '#777777' : 'white'}
                                style={{ userSelect: 'none' }}
                                size="20px"
                            >
                                {v}
                            </Text>
                            <Space style={{ display: 'flex', flexGrow: 1 }} />
                            <Box
                                className="xBox"
                                style={{
                                    pointerEvents: isPublic ? 'none' : 'all',
                                }}
                                onClick={() => {
                                    setSharedUsers((u) => {
                                        u.splice(u.indexOf(v), 1);
                                        return [...u];
                                    });
                                }}
                            >
                                <IconX
                                    scale={'3px'}
                                    color={isPublic ? '#777777' : 'white'}
                                />
                            </Box>
                        </RowBox>
                    );
                })}
                {sharedUsers.length === 0 && (
                    <Text style={{ height: '100%', userSelect: 'none' }}>
                        Not shared
                    </Text>
                )}
            </ColumnBox>
        </ColumnBox>
    );
}

export function ShareBox({
    candidates,
    authHeader,
}: {
    candidates: FileInfoT[];
    authHeader: AuthHeaderT;
}) {
    const [sharedUsers, setSharedUsers] = useState([]);
    const [pub, setPublic] = useState(false);
    const [shareData, setShareData]: [
        shareData: shareData[],
        setShareData: (v) => void
    ] = useState(null);
    useEffect(() => {
        if (
            !candidates ||
            candidates.length === 0 ||
            candidates[0].shares.length === 0
        ) {
            return;
        }
        setShareData(candidates[0].shares.filter((s) => !s.Wormhole));
        // GetFileShare(candidates[0].id, authHeader).then((v: shareData[]) => ).catch(r => notifications.show({ title: "Failed to get share data", message: String(r), color: 'red' }))
    }, [candidates, authHeader]);

    useEffect(() => {
        if (!shareData || shareData.length === 0) {
            return;
        }
        setPublic(shareData[0].Public);
        if (shareData[0].Accessors) {
            setSharedUsers(shareData[0].Accessors);
        }
    }, [shareData]);

    const shareOrUpdate = useCallback(async () => {
        if (!shareData || shareData.length === 0) {
            const res = await ShareFiles(
                candidates.map((v) => v.id),
                pub,
                sharedUsers,
                authHeader
            );
            return res.shareData.shareId;
        } else {
            await UpdateFileShare(
                shareData[0].shareId,
                pub,
                sharedUsers,
                authHeader
            );
            return shareData[0].shareId;
        }
    }, [candidates, shareData, pub, sharedUsers, authHeader]);

    return (
        <ColumnBox style={{ gap: 4 }}>
            <ShareInput
                isPublic={pub}
                sharedUsers={sharedUsers}
                setSharedUsers={setSharedUsers}
            />
            <WeblensButton
                toggleOn={pub}
                onClick={() => setPublic(!pub)}
                label={pub ? 'Public' : 'Private'}
                postScript={
                    pub
                        ? 'Anyone with link can access'
                        : 'Only shared users can access'
                }
                Left={pub ? <IconUsersGroup /> : <IconUserCancel />}
            />

            <RowBox
                style={{ justifyContent: 'space-between', maxWidth: '95%' }}
            >
                <WeblensButton
                    label="Copy link"
                    showSuccess
                    Left={<IconLink />}
                    onClick={(e) => {
                        e.stopPropagation();
                        shareOrUpdate().then((v) => {
                            navigator.clipboard.writeText(
                                `${window.location.origin}/share/${v}`
                            );
                        });
                        return true;
                    }}
                    style={{ width: '50%' }}
                />
                <WeblensButton
                    label="Save"
                    showSuccess
                    onClick={(e) => {
                        e.stopPropagation();
                        shareOrUpdate();
                        return true;
                    }}
                    style={{ width: '50%' }}
                />
            </RowBox>
        </ColumnBox>
    );
}
