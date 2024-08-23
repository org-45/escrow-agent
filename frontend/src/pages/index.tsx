import React, {useEffect, useState} from 'react';
import {getAllPendingEscrows, Escrow, EscrowAPI} from '../services/api';

const Home: React.FC = () => {
    const [pendingEscrows, setPendingEscrows] = useState<EscrowAPI[]>([]);
    const [error, setError] = useState<string | null>(null);

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
        </div>
    );
};

export default Home;
