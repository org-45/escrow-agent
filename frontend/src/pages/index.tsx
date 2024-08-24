import React, {useEffect, useState} from 'react';
import {getAllPendingEscrows, createEscrow, EscrowAPI} from '../services/api';

const Home: React.FC = () => {
    const [pendingEscrows, setPendingEscrows] = useState<EscrowAPI[]>([]);
    const [error, setError] = useState<string | null>(null);

    const [newEscrow, setNewEscrow] = useState({
        BuyerID: '',
        SellerID: '',
        Amount: 0,
        Description: '',
    });

    useEffect(() => {
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
    }, []);

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

    return (
        <div>
            <h1>Pending Escrows</h1>
            {error && <p>{error}</p>}

            <ul>
                {pendingEscrows.map(escrow => (
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
        </div>
    );
};

export default Home;
