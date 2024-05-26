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
      <t-button variant="text" shape="square">
        <template #icon>
          <t-icon name="ellipsis"/>
        </template>
      </t-button>
    </template>
  </t-head-menu>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import {useRouter} from 'vue-router'
import {MenuValue} from "tdesign-vue-next";

const router = useRouter()
const menuValue = ref('dashboard')

const onChange = (value: MenuValue) => {
  router.push({name: value})
}
</script>

<style scoped lang="scss">
.logo {

}
</style>
