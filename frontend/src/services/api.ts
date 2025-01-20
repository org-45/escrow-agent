import axios from 'axios';

export const API_BASE_URL = 'http://localhost:8080';

const getAuthToken = () => {
    return localStorage.getItem('escrow-agent-client-jwt');
};

export interface Escrow {
    id: string;
    buyer_id: string;
    seller_id: string;
    amount: number;
    status: string;
    created_at: string;
    released_at?: string;
    disputed_at?: string;
    description: string;
}

export interface EscrowAPI {
    ID: string;
    BuyerID: string;
    SellerID: string;
    Amount: number;
    Status: string;
    CreatedAt: string;
    ReleasedAt?: string;
    DisputedAt?: string;
    Description: string;
}

export interface User{
	username:string;
	id:number;
	role:string;
	created_at:string
}

export const getUserDetails = async () => {
    try {
        const token = getAuthToken();
        const response = await axios.get(`${process.env.NEXT_PUBLIC_API_BASE_URL}/profile`, {
            headers: {
                accept: 'application/json',
                Authorization: `Bearer ${token}`,
            },
        });
        return response.data;
    } catch (error) {
        console.error('Error creating escrow:', error);
        throw error;
    }
};

export const createEscrow = async (escrowData: Omit<EscrowAPI, 'ID' | 'Status' | 'CreatedAt'>): Promise<EscrowAPI> => {
    try {
        const token = getAuthToken();
        const response = await axios.post<any>(`${API_BASE_URL}/escrow`, escrowData, {
            headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${token}`,
            },
        });
        return response.data;
    } catch (error) {
        console.error('Error creating escrow:', error);
        throw error;
    }
};

export const getAllPendingEscrows = async (): Promise<EscrowAPI[]> => {
    const token = getAuthToken();

    try {
        const response = await axios.get<EscrowAPI[]>(`${API_BASE_URL}/escrow/pending`, {
            headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${token}`,
            },
        });
        return response.data;
    } catch (error) {
        console.error('Error fetching pending escrows:', error);
        throw error;
    }
};

export const releaseFunds = async (escrowId: string): Promise<void> => {
    const token = getAuthToken();
    try {
        await axios.post(`${API_BASE_URL}/escrow/${escrowId}/release`, {
            headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${token}`,
            },
        });
    } catch (error) {
        console.error('Error releasing funds:', error);
        throw error;
    }
};

export const disputeEscrow = async (escrowId: string): Promise<void> => {
    const token = getAuthToken();
    try {
        await axios.post(`${API_BASE_URL}/escrow/${escrowId}/dispute`, {
            headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${token}`,
            },
        });
    } catch (error) {
        console.error('Error disputing escrow:', error);
        throw error;
    }
};
