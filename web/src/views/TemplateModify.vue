<template>
  <div>
    <t-form ref="form" label-align="top" :label-width="100" :rules="FORM_RULES" :data="formData" @submit="onSubmit">
      <div class="form-basic-container">
        <div class="form-basic-item">
          <div class="form-basic-container-title">{{ mode === 'create' ? 'Create Template' : 'Modify Template' }}</div>

          <t-row class="row-gap" :gutter="[32,24]">
            <t-col :span="12">
              <t-form-item label="Name" name="name">
                <t-input v-model="formData.name"/>
              </t-form-item>
            </t-col>
            <t-col :span="12">
              <t-form-item label="Programming Languages" name="language">
                <t-checkbox-group v-model="formData.language">
                  <t-checkbox v-for="language in LANGUAGES" v-bind:key="language.value" :key="language.value"
                              :label="language.label" :value="language.value">
                  </t-checkbox>
                </t-checkbox-group>
              </t-form-item>
            </t-col>
            <t-col :span="12">
              <t-form-item label="Time Limit (s)" name="timeout">
                <t-input-number v-model="formData.timeout" :min="0" :max="60"/>
              </t-form-item>
            </t-col>
            <t-col :span="12">
              <t-form-item label="Maximum CPUs" name="maxCpus">
                <t-input-number v-model="formData.maxCpus" :min="0" :max="10"/>
              </t-form-item>
            </t-col>
            <t-col :span="12">
              <t-form-item label="Maximum Memory (MB)" name="maxMemory">
                <t-input-number v-model="formData.maxMemory" :min="6" :max="1024"/>
              </t-form-item>
            </t-col>
            <t-col :span="12">
              <t-form-item label="Internet Access" name="internetAccess">
                <t-switch v-model="formData.internetAccess"></t-switch>
              </t-form-item>
            </t-col>
            <t-col :span="12">
              <t-form-item label="Maximum Number of Containers" name="maxContainer">
                <t-input-number v-model="formData.maxContainer" :min="0" :max="1000"/>
              </t-form-item>
            </t-col>
            <t-col :span="12">
              <t-form-item label="Maximum Number of Containers per IP" name="maxContainerPerIp">
                <t-input-number v-model="formData.maxContainerPerIp" :min="0" :max="100"/>
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
import {
  type CreateTemplateReq,
  type UpdateTemplateReq,
  createTemplate,
  updateTemplate,
  getTemplate
} from "@/api/template";
import {LANGUAGES} from '@/const/template'

const route = useRoute()
const router = useRouter()

const mode = ref<'create' | 'edit'>(route.name === 'createTemplate' ? 'create' : 'edit')
const id = ref<string | undefined>(route.params.id as string | undefined)

const FORM_RULES: FormProps['rules'] = {
  name: [{required: true, message: 'Name is required'}],
  language: [{required: true, message: 'Programming Languages are required'}],
  timeout: [{required: true, message: 'Time Limit is required'}],
  maxCpus: [{required: true, message: 'Maximum CPUs is required'}],
  maxMemory: [{required: true, message: 'Maximum Memory is required'}],
  maxContainer: [{required: true, message: 'Maximum Number of Containers is required'}],
  maxContainerPerIp: [{required: true, message: 'Maximum Number of Containers per IP is required'}],
};

const formData = ref<CreateTemplateReq | UpdateTemplateReq>({language: []} as CreateTemplateReq)
const onSubmit = (ctx: SubmitContext) => {
  if (ctx.validateResult === true) {
    if (mode.value === 'create') {
      createTemplate(formData.value).then(() => {
        MessagePlugin.success('Create new template successfully')
        router.push({name: 'template'})
      })
    } else {
      updateTemplate(id.value, formData.value).then(res => {
        MessagePlugin.success(res)
      })
    }
  }
}

const onCancel = () => {
  router.push({name: 'template'})
}

onMounted(() => {
  if (mode.value === 'edit') {
    getTemplate(id.value).then(res => {
      formData.value = {
        name: res.name,
        language: res.language,
        timeout: res.timeout,
        maxCpus: res.maxCpus,
        maxMemory: res.maxMemory,
        internetAccess: res.internetAccess,
        maxContainer: res.maxContainer,
        maxContainerPerIp: res.maxContainerPerIp,
      }
    })
  }
})
</script>

<style scoped>

</style>
