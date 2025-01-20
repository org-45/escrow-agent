// pages/index.tsx
import React, {useState, useEffect} from 'react';
import {useRouter} from 'next/router';
import Signup from './signup';
import Login from './login';
import Home from './home'; 

const Main: React.FC = () => {
    const [jwt, setJwt] = useState<string | null>(null);
    const [currentPage, setCurrentPage] = useState<'signup' | 'login' | 'home'>('signup');
    const router = useRouter();

    useEffect(() => {
        const token = localStorage.getItem('escrow-agent-client-jwt');
        if (token) {
            setJwt(token);
            setCurrentPage('home');
        } else {
            setCurrentPage('signup');
        }
    }, []);

    const handleSignupSuccess = () => {
        setCurrentPage('login');
    };

    const onLogout = () => {
        // Clear JWT from local storage
        localStorage.removeItem('escrow-agent-client-jwt');
        setJwt(null);
        setCurrentPage('login');
        router.push('/login');
    };

    if (currentPage === 'signup') {
        return <Signup onSignupSuccess={handleSignupSuccess} />;
    }

    if (currentPage === 'login') {
        return <Login/>;
    }

    if (currentPage === 'home') {
        return <Home onLogout={onLogout}/>;
    }

    return null;
};

export default Main;
