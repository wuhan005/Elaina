import {defineStore} from "pinia";
import {AuthState} from "./types";

const useAuthStore = defineStore('auth', {
    persist: true,

    state: (): AuthState => ({
        isAuthenticated: false,
        token: '',
    }),

    actions: {
        setToken(token: string) {
            this.token = token;
            this.isAuthenticated = true;
        },

        cleanToken() {
            this.token = '';
            this.isAuthenticated = false;
        },
    }
})

export default useAuthStore;
