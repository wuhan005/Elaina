import axios from 'axios';
import type {AxiosResponse} from 'axios';
import {MessagePlugin} from 'tdesign-vue-next';
import {useAuthStore} from '@/store';

export interface HttpResponse<T = unknown> {
    msg: string;
    data: T;
}

axios.interceptors.request.use(
    (config) => {
        const authStore = useAuthStore();
        const token = authStore.token
        if (authStore.isAuthenticated && token) {
            config.headers.Authorization = `Token ${token}`;
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

axios.interceptors.response.use(
    (response: AxiosResponse) => {
        const res = response.data;
        const statusCode = response.status
        if (statusCode / 100 !== 2) {
            MessagePlugin.error(res.msg || 'Unknown error')
        }
        return res.data;
    },
    (error) => {
        const statusCode = error.response.status
        if (statusCode === 401) {
            const authStore = useAuthStore();
            authStore.cleanToken();
            window.location.reload();
            return
        }

        const msg = error.response.data.msg
        MessagePlugin.error(msg || 'Unknown error')
        return Promise.reject(error);
    }
);
