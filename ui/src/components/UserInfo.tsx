import { useEffect, useState } from 'react';
import { useCookies } from 'react-cookie';
import API_ENDPOINT from '../api/ApiEndpoint';
import { notifications } from '@mantine/notifications';
import { UserContextT, UserInfoT } from '../types/Types';

const useR = () => {
    const [cookies, setCookie, removeCookie] = useCookies([
        'weblens-username',
        'weblens-login-token',
    ]);
    const [authHeader, setAuthHeader] = useState<{ Authorization: string }>({
        Authorization: '',
    });
    const [usr, setUserInfo]: [usr: UserInfoT, setUserInfo: any] = useState({
        homeId: '',
        trashId: '',
        username: '',
        admin: false,
        owner: false,
        activated: false,
        isLoggedIn: undefined,
    });

    const clear = () => {
        setAuthHeader({ Authorization: '' });
        setUserInfo({
            admin: false,
            homeId: '',
            trashId: '',
            username: '',
            activated: false,
            owner: false,
            isLoggedIn: false,
        } as UserInfoT);
        removeCookie('weblens-username');
        removeCookie('weblens-login-token');
    };

    useEffect(() => {
        if (
            authHeader.Authorization === '' &&
            cookies['weblens-username'] &&
            cookies['weblens-login-token']
        ) {
            // Auth header unset, but the cookies are ready
            const loginStr = `${cookies['weblens-username']}:${cookies['weblens-login-token']}`;
            // console.log(loginStr.replace(/-/g, '+').replace(/_/g, '/'));
            const login64 = window.btoa(loginStr);
            setAuthHeader({
                Authorization: `Basic ${login64.toString()}`,
            });
        } else if (
            authHeader.Authorization !== '' &&
            (usr.username === '' || Object.keys(usr).length === 0)
        ) {
            // Auth header set, but no user data, go get the user data

            let url = new URL(`${API_ENDPOINT}/user`);
            fetch(url.toString(), { headers: authHeader })
                .then((res) => res.json())
                .then((json) => {
                    if (!json) {
                        return Promise.reject('Invalid user data');
                    }
                    setUserInfo({ ...json, isLoggedIn: true });
                })
                .catch((r) => notifications.show({ message: String(r) }));
        } else if (authHeader.Authorization === '') {
            setUserInfo((p) => {
                p.isLoggedIn = false;
                return { ...p };
            });
        }
    }, [authHeader, cookies]);

    return { authHeader, usr, setCookie, clear } as UserContextT;
};

export default useR;
