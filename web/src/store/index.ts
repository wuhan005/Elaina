import {createPinia} from 'pinia';
import useAuthStore from './modules/auth';

const pinia = createPinia();
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'

pinia.use(piniaPluginPersistedstate)

export {useAuthStore};
export default pinia;
