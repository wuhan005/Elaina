<template>
  <div>
    <t-form ref="form" label-align="top" :label-width="100" :rules="FORM_RULES" :data="formData" @submit="onSubmit">
      <div class="form-basic-container">
        <div class="form-basic-item">
          <div class="form-basic-container-title">{{ mode === 'create' ? 'Create Template' : 'Modify Template' }}</div>

          <t-row class="row-gap" :gutter="[32,24]">
            <t-col :span="12">
              <t-alert theme="info" message="Templates control execution time, WASM linear memory and network permission. Server configuration caps memory, global concurrency, input and workspace usage. Per-task CPU quotas and custom DNS are unavailable."/>
            </t-col>
            <t-col v-if="legacyMaxCpus !== 0" :span="12">
              <t-alert theme="warning" :message="`This template has a legacy CPU quota of ${legacyMaxCpus} and cannot run in WASM. Saving removes this unsupported quota; no per-task CPU quota will be enforced.`"/>
            </t-col>
            <t-col :span="12">
              <t-form-item label="Name" name="name">
                <t-input v-model="formData.name"/>
              </t-form-item>
            </t-col>
            <t-col :span="12">
              <t-form-item label="Programming Languages" name="language" help="WASM: C (Clang 22), Go 1.27.3, Python 3.12, PHP 8.2, QuickJS (not Node.js). Native extensions and subprocesses are unavailable.">
                <t-checkbox-group v-model="formData.language">
                  <t-checkbox v-for="language in LANGUAGES" v-bind:key="language.value" :label="language.label"
                              :value="language.value">
                  </t-checkbox>
                </t-checkbox-group>
              </t-form-item>
            </t-col>
            <t-col :span="12">
              <t-form-item label="Execution Timeout (s)" name="timeout" help="0 uses the 10-second default. Includes compilation and startup; 60 seconds is recommended for C/Go cold starts.">
                <t-input-number v-model="formData.timeout" :min="0" :max="60" :decimal-places="0"/>
              </t-form-item>
            </t-col>
            <t-col :span="12">
              <t-form-item label="WASM Linear Memory (MiB)" name="maxMemory" help="0 inherits the server ceiling; nonzero values are capped by it. Applies separately to each compile, link and execution stage, not total process memory. Start with 512 MiB for C/Go, 128 MiB for interpreters.">
                <t-input-number v-model="formData.maxMemory" :min="0" :max="4096" :decimal-places="0"/>
              </t-form-item>
            </t-col>
            <t-col :span="12">
              <t-form-item label="Network Permission" name="internetAccess" help="Off blocks sockets and DNS. On permits supported networking: Go HTTP/DNS/HTTPS, C IPv4 TCP/DNS without bundled TLS, PHP IPv4 TCP/HTTP. Python/JavaScript networking is unavailable. Compilation is always offline.">
                <t-switch v-model="formData.internetAccess"></t-switch>
              </t-form-item>
            </t-col>
          </t-row>
        </div>
      </div>

      <div class="form-submit-container">
        <div class="form-submit-sub">
          <div class="form-submit-left">
            <t-space>
              <t-button theme="primary" class="form-submit-confirm" type="submit">Submit</t-button>
              <t-button type="reset" class="form-submit-cancel" theme="default" variant="base" @click="onCancel">
                Cancel
              </t-button>
            </t-space>
          </div>
        </div>
      </div>
    </t-form>
  </div>
</template>

<script setup lang="ts">
import '@/style/form.less'
import {onMounted, ref} from 'vue'
import {FormProps, MessagePlugin, SubmitContext} from 'tdesign-vue-next';
import {useRouter, useRoute} from "vue-router";
import type {CreateTemplateReq, UpdateTemplateReq} from "@/api/api";
import {client} from "@/api/client";
import {LANGUAGES} from '@/const/template'

const route = useRoute()
const router = useRouter()
const mode = ref<'create' | 'edit'>(route.name === 'createTemplate' ? 'create' : 'edit')
const id = ref<string | undefined>(route.params.id as string | undefined)
const legacyMaxCpus = ref(0)

const FORM_RULES: FormProps['rules'] = {
  name: [{required: true, message: 'Name is required'}],
  language: [{required: true, message: 'Programming Languages are required'}],
  timeout: [{required: true, message: 'Execution Timeout is required'},
    {validator: value => Number.isInteger(value) && value >= 0 && value <= 60, message: 'Use a whole number from 0 to 60'}],
  maxMemory: [{required: true, message: 'WASM Linear Memory is required'},
    {validator: value => Number.isInteger(value) && value >= 0 && value <= 4096, message: 'Use a whole number from 0 to 4096'}],
};

const formData = ref<CreateTemplateReq | UpdateTemplateReq>({
  name: '', language: [], timeout: 60, maxMemory: 512, internetAccess: false,
})
const onSubmit = (ctx: SubmitContext) => {
  if (ctx.validateResult === true) {
    if (mode.value === 'create') {
      client.templates.createTemplate(formData.value).then(() => {
        MessagePlugin.success('Create new template successfully')
        router.push({name: 'template'})
      })
    } else {
      client.templates.updateTemplate({
        id: Number(id.value),
      }, formData.value).then(res => {
        legacyMaxCpus.value = 0
        MessagePlugin.success(res.data.data)
      })
    }
  }
}

const onCancel = () => {
  router.push({name: 'template'})
}

onMounted(() => {
  if (mode.value === 'edit') {
    client.templates.getTemplate({
      id: Number(id.value),
    }).then(res => {
      const template = res.data.data
      legacyMaxCpus.value = template.maxCpu

      formData.value = {
        name: template.name,
        language: template.language,
        timeout: template.timeout,
        maxMemory: template.maxMemory,
        internetAccess: template.internetAccess,
      }
    })
  }
})
</script>

<style scoped>

</style>
