import { useCallback, useContext, useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { createUser, login } from '../../api/ApiFetch';
import { UserContext } from '../../Context';
import { notifications } from '@mantine/notifications';
import { Input, Space, Tabs } from '@mantine/core';
import WeblensButton from '../../components/WeblensButton';
import { useKeyDown } from '../../components/hooks';
import { UserContextT } from '../../types/Types';
import WeblensInput from '../../components/WeblensInput';

async function CheckCreds(username: string, password: string, setCookie, nav) {
    if (!username || !password) {
        return false;
    }
    return await login(username, password)
        .then(res => {
            if (res.status !== 200) {
                return Promise.reject('Incorrect username or password');
            } else {
                return res.json();
            }
        })
        .then(data => {
            setCookie('weblens-username', username, { sameSite: 'strict' });
            setCookie('weblens-login-token', data.token, {
                sameSite: 'strict',
            });
            nav('/');
            return true;
        })
        .catch(r => {
            notifications.show({ message: String(r), color: 'red' });
            return false;
        });
}

async function CreateUser(username: string, password: string): Promise<boolean> {
    return await createUser(username, password)
        .then(x => {
            return true;
        })
        .catch(r => {
            console.error(r);
            return false;
        });
}

export const useKeyDownLogin = login => {
    const onKeyDown = useCallback(
        event => {
            if (event.key === 'Enter') {
                login();
            }
        },
        [login],
    );

    useEffect(() => {
        document.addEventListener('keydown', onKeyDown);
        return () => {
            document.removeEventListener('keydown', onKeyDown);
        };
    }, [onKeyDown]);
};

const Login = () => {
    const [userInput, setUserInput] = useState('');
    const [passInput, setPassInput] = useState('');
    const [tab, setTab] = useState('login');
    const nav = useNavigate();
    const loc = useLocation();
    const { authHeader, setCookie }: UserContextT = useContext(UserContext);

    useEffect(() => {
        if (loc.state == null && authHeader.Authorization !== '') {
            nav('/');
        }
    }, [authHeader, loc.state, nav]);

    const [buttonRef, setButtonRef] = useState(null);
    useKeyDown('Enter', e => {
        if (buttonRef) {
            buttonRef.click();
        }
    });
    const badUsername = userInput[0] === '.' || userInput.includes('/');

    return (
        <div
            className="flex flex-row h-screen w-screen items-center justify-center"
            style={{
                background: 'linear-gradient(45deg, rgba(2,0,36,1) 0%, rgba(94,43,173,1) 50%, rgba(0,212,255,1) 100%)',
            }}
        >
            {/* <ScatteredPhotos /> */}
            <div className="flex flex-col justify-center items-center shadow-soft bg-bottom-grey outline outline-main-accent rounded-xl p-6 w-[400px] max-w-[600px] max-h-[400px]">
                <Tabs
                    className="flex flex-col w-full h-[90%] justify-center items-center gap-5"
                    value={tab}
                    onChange={setTab}
                    keepMounted={false}
                    variant="pills"
                >
                    <Tabs.List grow className="w-full mb-2">
                        <Tabs.Tab value="login" className="bg-main-accent">
                            Login
                        </Tabs.Tab>
                        <Tabs.Tab value="sign-up" className="bg-main-accent">
                            Sign Up
                        </Tabs.Tab>
                    </Tabs.List>
                    <Tabs.Panel value="login" className="flex flex-col justify-center items-center w-full">
                        <WeblensInput
                            placeholder="Username"
                            value={userInput}
                            autoFocus
                            onComplete={() => {}}
                            valueCallback={setUserInput}
                            height={40}
                        />

                        <WeblensInput
                            placeholder="Password"
                            value={passInput}
                            onComplete={() => {}}
                            valueCallback={setPassInput}
                            height={40}
                        />
                        <Space h={'md'} />
                        <WeblensButton
                            label="Login"
                            fillWidth
                            squareSize={50}
                            disabled={userInput === '' || passInput === ''}
                            centerContent
                            onClick={() => CheckCreds(userInput, passInput, setCookie, nav)}
                            setButtonRef={setButtonRef}
                        />
                    </Tabs.Panel>
                    <Tabs.Panel
                        value="sign-up"
                        style={{
                            display: 'flex',
                            flexDirection: 'column',
                            justifyContent: 'center',
                            alignItems: 'center',
                            width: '100%',
                        }}
                    >
                        <WeblensInput
                            placeholder="Username"
                            value={userInput}
                            autoFocus
                            onComplete={() => {}}
                            valueCallback={setUserInput}
                            height={40}
                        />

                        {badUsername && (
                            <Input.Error style={{ width: '100%' }}>
                                Username must not begin with '.' and cannot include '/'
                            </Input.Error>
                        )}
                        <WeblensInput
                            placeholder="Password"
                            value={passInput}
                            onComplete={() => {}}
                            valueCallback={setPassInput}
                            height={40}
                        />
                        <Space h={'md'} />
                        <WeblensButton
                            label="Sign Up"
                            doSuper
                            squareSize={50}
                            fillWidth
                            disabled={userInput === '' || passInput === '' || badUsername}
                            centerContent
                            onClick={async () => CreateUser(userInput, passInput)}
                            setButtonRef={setButtonRef}
                        />
                    </Tabs.Panel>
                </Tabs>
            </div>
        </div>
    );
};

export default Login;
