import React, {ChangeEvent, FormEvent, useEffect, useState} from 'react';
import {getAllPendingEscrows, createEscrow, EscrowAPI, getUserDetails, User} from '../services/api';
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
    const [pendingEscrows, setPendingEscrows] = useState<EscrowAPI[]>([]);
    const [error, setError] = useState<string | null>(null);
    const router = useRouter();

    const [newEscrow, setNewEscrow] = useState({
        BuyerID: '',
        SellerID: '',
        Amount: 0,
        Description: '',
    });

    const [file, setFile] = useState<File | null>(null);
    const [uploadMessage, setUploadMessage] = useState<string>('');

    useEffect(() => {
        const token = localStorage.getItem('escrow-agent-client-jwt');
        if (!token) {
            router.push('/login');
        } else {
            const fetchPendingEscrows = async () => {
                try {
                    const escrows = await getAllPendingEscrows();
                    setPendingEscrows(escrows);
                } catch (err) {
                    setError('Failed to fetch pending escrows');
                    console.error(err);
                }
            };
            fetchPendingEscrows();

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

    const handleCreateEscrow = async (event: React.FormEvent) => {
        event.preventDefault();
        try {
            const escrowData = {
                BuyerID: newEscrow.BuyerID,
                SellerID: newEscrow.SellerID,
                Amount: newEscrow.Amount,
                Description: newEscrow.Description,
            };

            const escrow = await createEscrow(escrowData);
            setPendingEscrows([...pendingEscrows, escrow]);
            setNewEscrow({BuyerID: '', SellerID: '', Amount: 0, Description: ''});
        } catch (err) {
            setError('Failed to create escrow');
            console.error(err);
        }
    };

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

            <h2>Create New Escrow</h2>
            <form onSubmit={handleCreateEscrow}>
                <div>
                    <label>Buyer ID:</label>
                    <input
                        type="text"
                        value={newEscrow.BuyerID}
                        onChange={e => setNewEscrow({...newEscrow, BuyerID: e.target.value})}
                        required
                    />
                </div>
                <div>
                    <label>Seller ID:</label>
                    <input
                        type="text"
                        value={newEscrow.SellerID}
                        onChange={e => setNewEscrow({...newEscrow, SellerID: e.target.value})}
                        required
                    />
                </div>
                <div>
                    <label>Amount:</label>
                    <input
                        type="number"
                        value={newEscrow.Amount}
                        onChange={e => setNewEscrow({...newEscrow, Amount: parseFloat(e.target.value)})}
                        required
                    />
                </div>
                <div>
                    <label>Description:</label>
                    <input
                        type="text"
                        value={newEscrow.Description}
                        onChange={e => setNewEscrow({...newEscrow, Description: e.target.value})}
                        required
                    />
                </div>
                <button type="submit">Create Escrow</button>
            </form>
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
        seller_id: 0,
        amount: 0,
        status: 'pending', // Default status
    });

    const handleChange = (e: any) => {
        const {name, value, type} = e.target;
        setFormData({
            ...formData,
            [name]: type === 'number' ? Number(value) : value,
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
