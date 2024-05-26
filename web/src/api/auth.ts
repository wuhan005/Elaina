import axios from 'axios';

export interface SignInForm {
    password: string;
}

export function signIn(data: SignInForm) {
    return axios.post<string>('/api/auth/sign-in', data)
}

export interface ProfileResp {

}

export function profile() {
    return axios.get<ProfileResp>('/api/auth/profile')
}
