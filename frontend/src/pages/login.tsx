// pages/login.tsx
import React, {useState} from 'react';
import axios from 'axios';
import {useRouter} from 'next/router';

const Login: React.FC = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState<string | null>(null);
    const router = useRouter();

    const handleLogin = async (event: React.FormEvent) => {
        event.preventDefault();
        try {
            const response = await axios.post(`${process.env.NEXT_PUBLIC_API_BASE_URL}/login`, {
                username,
                password,
            });
            const {token} = response.data;

            // Store the token in localStorage
            localStorage.setItem('jwt', token);

            // Redirect to a protected page or home page
            router.push('/');
        } catch (err) {
            setError('Login failed. Please check your username and password.');
            console.error(err);
        }
    };

    return (
        <div>
            <h1>Login</h1>
            {error && <p style={{color: 'red'}}>{error}</p>}
            <form onSubmit={handleLogin}>
                <div>
                    <label>Username:</label>
                    <input type="text" value={username} onChange={e => setUsername(e.target.value)} required />
                </div>
                <div>
                    <label>Password:</label>
                    <input type="password" value={password} onChange={e => setPassword(e.target.value)} required />
                </div>
                <button type="submit">Login</button>
            </form>
        </div>
    );
};

export default Login;
