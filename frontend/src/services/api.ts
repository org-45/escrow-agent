import axios from 'axios';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';

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

export const createEscrow = async (escrowData: Omit<EscrowAPI, 'id' | 'status' | 'created_at'>): Promise<EscrowAPI> => {
    try {
        const response = await axios.post<EscrowAPI>(`${API_BASE_URL}/escrow`, escrowData);
        return response.data;
    } catch (error) {
        console.error('Error creating escrow:', error);
        throw error;
    }
};

export const getAllPendingEscrows = async (): Promise<EscrowAPI[]> => {
    try {
        const response = await axios.get<EscrowAPI[]>(`${API_BASE_URL}/escrow/pending`);
        return response.data;
    } catch (error) {
        console.error('Error fetching pending escrows:', error);
        throw error;
    }
};

// Add more functions for other API endpoints as needed
