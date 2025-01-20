import React, {ChangeEvent, FormEvent, useEffect, useState} from 'react';
import {getAllPendingEscrows, createEscrow, EscrowAPI} from '../services/api';
import {useRouter} from 'next/router';
import axios from 'axios';

interface LogoutProps {
    onLogout: () => void;
}


const Home: React.FC<LogoutProps> = ({onLogout}) => {
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
            <h1>Pending Escrows</h1>
            <button onClick={onLogout}>Logout</button>

            {error && <p>{error}</p>}

            <ul>
                {pendingEscrows?.map(escrow => (
                    <li key={escrow.ID}>
                        {escrow.BuyerID} owes {escrow.Amount} to {escrow.SellerID}
                    </li>
                ))}
            </ul>

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
