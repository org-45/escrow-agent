import React, {useState} from 'react';
import axios from 'axios';
import {API_BASE_URL} from '@/services/api';

interface SignupProps {
    onSignupSuccess: () => void;
}

const Signup: React.FC<SignupProps> = ({onSignupSuccess}) => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [role, setRole] = useState('');
    const [error, setError] = useState<string | null>(null);

    const handleSignup = async (event: React.FormEvent) => {
        event.preventDefault();
        try {
            const response = await axios.post(`${API_BASE_URL}/register`, {
                username,
                password,
                role,
            });

            if (response.status === 201) {
                onSignupSuccess();
            }
        } catch (err: any) {
            if (err.response && err.response.status === 409) {
                setError('Username already exists. Please choose a different username.');
            } else {
                setError('Signup failed. Please try again.');
            }
            console.error(err);
        }
    };

    return (
        <div>
            <h1>Signup</h1>
            {error && <p style={{color: 'red'}}>{error}</p>}
            <form onSubmit={handleSignup}>
                <div>
                    <label>Username:</label>
                    <input type="text" value={username} onChange={e => setUsername(e.target.value)} required />
                </div>
                <div>
                    <label>Password:</label>
                    <input type="password" value={password} onChange={e => setPassword(e.target.value)} required />
                </div>
                <div>
                    <label>Role:</label>
                    <input type="text" value={role} onChange={e => setRole(e.target.value)} required />
                </div>

                <button type="submit">Signup</button>
            </form>
        </div>
    );
};

export default Signup;
