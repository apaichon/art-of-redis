// Default values for local development
const DEFAULT_WS_PORT = 9003;
const DEFAULT_WS_HOST = 'localhost';
const DEFAULT_WS_PROTOCOL = window.location.protocol === 'https:' ? 'wss:' : 'ws:';

// Environment variables from Vite
const WS_PORT = import.meta.env.VITE_WS_PORT || DEFAULT_WS_PORT;
const WS_HOST = import.meta.env.VITE_WS_HOST || DEFAULT_WS_HOST;
const WS_PROTOCOL = import.meta.env.VITE_WS_PROTOCOL || DEFAULT_WS_PROTOCOL;

export const WS_URL = `${WS_PROTOCOL}//${WS_HOST}:${WS_PORT}/ws`;

// You can add more configuration variables here
export const API_CONFIG = {
  baseUrl: `http://${WS_HOST}:${WS_PORT}`,
  endpoints: {
    sales: '/api/sales'
  }
};