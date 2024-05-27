<template>
  <t-head-menu v-model="menuValue" theme="light" @change="onChange">
    <template #logo>
      <div class="logo">Elaina</div>
    </template>

    <t-menu-item value="dashboard">
      <template #icon>
        <t-icon name="dashboard"/>
      </template>
      Dashboard
    </t-menu-item>
    <t-menu-item value="template">
      <template #icon>
        <t-icon name="template"/>
      </template>
      Template
    </t-menu-item>
    <t-menu-item value="sandbox">
      <template #icon>
        <t-icon name="system-components"/>
      </template>
      Sandbox
    </t-menu-item>
    <template #operations>
      <t-dropdown :options="[{ content: 'Sign Out', value: 'sign-out' }]" @click="onClickDropDown">
        <t-button variant="text" shape="square">
          <template #icon>
            <t-icon name="ellipsis"/>
          </template>
        </t-button>
      </t-dropdown>
    </template>
  </t-head-menu>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {MenuValue} from "tdesign-vue-next";
import {DropdownOption} from "tdesign-vue-next";
import {useAuthStore} from "@/store";

const route = useRoute()
const router = useRouter()
const menuValue = ref(route.name)
const authStore = useAuthStore()

const onChange = (value: MenuValue) => {
  router.push({name: value.toString()})
}

const onClickDropDown = (dropdownItem: DropdownOption, context: { e: MouseEvent }) => {
  switch (dropdownItem.value) {
    case 'sign-out':
      authStore.cleanToken()
      router.push({name: 'signIn'})
      break
  }
}
</script>

<style scoped lang="scss">
.logo {

}
</style>
