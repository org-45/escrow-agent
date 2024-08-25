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
        const token = localStorage.getItem('jwt');
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

    const handleLoginSuccess = (token: string) => {
        localStorage.setItem('jwt', token);
        setJwt(token);
        setCurrentPage('home');
        router.push('/');
    };

    if (currentPage === 'signup') {
        return <Signup onSignupSuccess={handleSignupSuccess} />;
    }

    if (currentPage === 'login') {
        return <Login onLoginSuccess={handleLoginSuccess} />;
    }

    if (currentPage === 'home') {
        return <Home />;
    }

    return null;
};

export default Main;
