import axios from 'axios';
import {MessagePlugin} from 'tdesign-vue-next';
import {useAuthStore} from '@/store';
import {Api, type ErrorResponse} from './api';

export const client = new Api({
    baseURL: '/',
});

client.instance.interceptors.request.use((config) => {
    const authStore = useAuthStore();
    if (authStore.isAuthenticated && authStore.token) {
        config.headers.Authorization = `Token ${authStore.token}`;
    }

    return config;
});

client.instance.interceptors.response.use(
    (response) => response,
    (error: unknown) => {
        if (axios.isAxiosError<ErrorResponse>(error)) {
            if (error.response?.status === 401) {
                useAuthStore().cleanToken();
                window.location.reload();
                return Promise.reject(error);
            }

            MessagePlugin.error(error.response?.data?.msg || error.message || 'Unknown error');
        } else {
            MessagePlugin.error('Unknown error');
        }

        return Promise.reject(error);
    },
);
