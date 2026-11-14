import {createPinia} from 'pinia';
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'
import useAuthStore from './modules/auth';

const pinia = createPinia();
pinia.use(piniaPluginPersistedstate)

export {useAuthStore};
export default pinia;
