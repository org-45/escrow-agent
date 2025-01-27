import React, {ChangeEvent, FormEvent, useEffect, useState} from 'react';
import { createEscrow, getTransactions, getUserDetails, User} from '../services/api';
import {useRouter} from 'next/router';
import axios from 'axios';

interface LogoutProps {
    onLogout: () => void;
}

const Home: React.FC<LogoutProps> = ({onLogout}) => {
    const [user, setUser] = useState<User>({
        id: 1,
        username: 'chauchausoup',
        role: 'buyer, seller, admin',
        created_at: '2023-10-01T15:23:45Z',
    });
    const [error, setError] = useState<string | null>(null);
    const router = useRouter();


    const [file, setFile] = useState<File | null>(null);
    const [uploadMessage, setUploadMessage] = useState<string>('');

    useEffect(() => {
        const token = localStorage.getItem('escrow-agent-client-jwt');
        if (!token) {
            router.push('/login');
        } else {

            const fetchUserDetails = async () => {
                try {
                    const u = await getUserDetails();
                    setUser(u);
                    localStorage.setItem('escrow-agent-client-role', u.role);
                } catch (err) {
                    setError('Failed to fetch user details');
                    console.error(err);
                }
            };
            fetchUserDetails();
        }
    }, [router]);

    const handleFileChange = (event: ChangeEvent<HTMLInputElement>) => {
        if (event.target.files && event.target.files.length > 0) {
            setFile(event.target.files[0]);
        }
    };

    const handleFileUpload = async (event: FormEvent) => {
        event.preventDefault();

        if (!file) {
            setUploadMessage('Please select a file to upload.');
            return;
        }

        const token = localStorage.getItem('escrow-agent-client-jwt');

        if (!token) {
            setUploadMessage('You must be logged in to upload files.');
            return;
        }

        const formData = new FormData();
        formData.append('file', file);

        try {
            const response = await axios.post(`${process.env.NEXT_PUBLIC_API_BASE_URL}/upload`, formData, {
                headers: {
                    'Content-Type': 'multipart/form-data',
                    Authorization: `Bearer ${token}`,
                },
            });
            setUploadMessage(`File uploaded successfully: ${response.data.file_url}`);
        } catch (error: any) {
            setUploadMessage(`Failed to upload file: ${error.message}`);
        }
    };

    return (
        <div>
            <h3>User details</h3>
            <p>
                {user.username} {user.id} {user.role} {user.created_at}
            </p>

            <button onClick={onLogout}>Logout</button>

            {localStorage.getItem('escrow-agent-client-role') == 'buyer' ? <CreateTransaction /> : null}

						<TransactionsTable/>
						
            <h4>File upload</h4>
            <form onSubmit={handleFileUpload}>
                <div>
                    <input type="file" onChange={handleFileChange} />
                </div>
                <button type="submit">Upload File</button>
            </form>
            {uploadMessage && <p>{uploadMessage}</p>}
        </div>
    );
};

export default Home;

function CreateTransaction() {
    const [formData, setFormData] = useState({
        seller_id: '',
        amount: '',
        status: 'pending',
    });

    const handleChange = (e: any) => {
        const {name, value, type} = e.target;
        const normalizedValue = type === 'number' ? (value === '' ? '' : Number(value)) : value;

        setFormData({
            ...formData,
            [name]: normalizedValue,
        });
    };

    const handleSubmit = async (e: any) => {
        e.preventDefault();
        console.log('Form Submitted:', formData);
        const token = localStorage.getItem('escrow-agent-client-jwt');

        try {
            await axios.post(`${process.env.NEXT_PUBLIC_API_BASE_URL}/transactions`, formData, {
                headers: {
                    Accept: 'application/json',
                    Authorization: `Bearer ${token}`,
                },
            });
        } catch (error: any) {
            console.log(error);
        }
    };

    return (
        <form onSubmit={handleSubmit}>
            <label htmlFor="seller_id">Seller ID:</label>
            <input
                type="number"
                id="seller_id"
                name="seller_id"
                value={formData.seller_id}
                onChange={handleChange}
                required
                placeholder="Enter seller ID"
            />

            <label htmlFor="amount">Amount:</label>
            <input
                type="number"
                id="amount"
                name="amount"
                value={formData.amount}
                onChange={handleChange}
                required
                placeholder="Enter amount"
            />

            <label htmlFor="status">Status:</label>
            <select id="status" name="status" value={formData.status} onChange={handleChange} required>
                <option value="pending">Pending</option>
                <option value="held">Held</option>
                <option value="released">Released</option>
                <option value="cancelled">Cancelled</option>
            </select>

            <button type="submit">Submit</button>
        </form>
    );
}


const TransactionsTable = () => {
    const [transactions, setTransactions] = useState([
        {
            transaction_id: 2,
            buyer_id: 4,
            seller_id: 2,
            amount: 25,
            status: 'pending',
            created_at: '2025-01-20T11:21:28.624658Z',
            updated_at: '2025-01-20T11:21:28.624658Z',
        },
    ]); // State for transactions
    const [loading, setLoading] = useState(true); // State for loading indicator
    const [error, setError] = useState(null); // State for error handling

    // Fetch transactions from an API or a mock function
    useEffect(() => {
        const fetchTransactions = async () => {
            try {
                setLoading(true);
                const token = localStorage.getItem('escrow-agent-client-jwt');
                const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/transactions`, {
                    headers: {
                        Accept: 'application/json',
                        Authorization: `Bearer ${token}`,
                    },
                });
                if (!response.ok) {
                    throw new Error('Failed to fetch transactions');
                }
                const data = await response.json();
                setTransactions(data);
            } catch (err:any) {
                setError(err.message);
            } finally {
                setLoading(false);
            }
        };

        fetchTransactions();
    }, []);

    if (loading) return <p>Loading...</p>;
    if (error) return <p>Error: {error}</p>;

    return (
        <div>
            <h4>Transactions</h4>
            <table style={{width: '50%', borderCollapse: 'collapse'}}>
                <thead>
                    <tr>
                        <th>Transaction ID</th>
                        <th>Buyer ID</th>
                        <th>Seller ID</th>
                        <th>Amount</th>
                        <th>Status</th>
                        <th>Created At</th>
                        <th>Updated At</th>
                    </tr>
                </thead>
                <tbody>
                    {transactions?.map(transaction => (
                        <tr key={transaction.transaction_id}>
                            <td>{transaction.transaction_id}</td>
                            <td>{transaction.buyer_id}</td>
                            <td>{transaction.seller_id}</td>
                            <td>{transaction.amount}</td>
                            <td>{transaction.status}</td>
                            <td>{new Date(transaction.created_at).toLocaleString()}</td>
                            <td>{new Date(transaction.updated_at).toLocaleString()}</td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );
};
