import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import axios from 'axios';
import { API_BASE_URL } from '@/services/api';

const Login: React.FC = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState<string | null>(null);
    const router = useRouter();

    useEffect(() => {
        if (typeof window !== 'undefined') {
            const token = localStorage.getItem('jwt');
            if (token) {
                router.push('/');
            }
        }
    }, [router]);

    const onLoginSuccess = (token: string) => {
        if (typeof window !== 'undefined') {
            localStorage.setItem('jwt', token); 
            router.push('/');
        }
    };

    const handleLogin = async (event: React.FormEvent) => {
        event.preventDefault();
        try {
            const response = await axios.post(`${API_BASE_URL}/login`, {
                username,
                password,
            });
            const { token } = response.data;
            onLoginSuccess(token);
        } catch (err) {
            setError('Login failed. Please check your username and password.');
            console.error(err);
        }
    };

    return (
        <div>
            <h1>Login</h1>
            {error && <p style={{ color: 'red' }}>{error}</p>}
            <form onSubmit={handleLogin}>
                <div>
                    <label>Username:</label>
                    <input type="text" value={username} onChange={(e) => setUsername(e.target.value)} required />
                </div>
                <div>
                    <label>Password:</label>
                    <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
                </div>
                <button type="submit">Login</button>
            </form>
        </div>
    );
};

export default Login;
