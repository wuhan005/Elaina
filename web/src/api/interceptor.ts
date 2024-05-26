import axios from 'axios';
import type {AxiosResponse} from 'axios';
import {MessagePlugin} from 'tdesign-vue-next';

export interface HttpResponse<T = unknown> {
    msg: string;
    data: T;
}

axios.interceptors.response.use(
    (response: AxiosResponse<HttpResponse>) => {
        const res = response.data;
        const statusCode = response.status
        if (statusCode / 100 !== 2) {
            MessagePlugin.error(res.msg || '未知错误')
        }
        return res.data;
    },
    (error) => {
        const msg = error.response.data.msg
        MessagePlugin.error(msg || '未知错误')
        return Promise.reject(error);
    }
);
