<template>
  <div>
    <t-form ref="form" label-align="top" :label-width="100" :rules="FORM_RULES" :data="formData" @submit="onSubmit">
      <div class="form-basic-container">
        <div class="form-basic-item">
          <div class="form-basic-container-title">{{ mode === 'create' ? 'Create Sandbox' : 'Modify Sandbox' }}</div>

          <t-row class="row-gap" :gutter="[32,24]">
            <t-col :span="12">
              <t-form-item label="Name" name="name">
                <t-input v-model="formData.name"/>
              </t-form-item>
            </t-col>
            <t-col :span="12">
              <t-form-item label="Template" name="templateID">
                <t-select v-model="formData.templateID">
                  <t-option v-for="template in templates" :key="template.id" :value="template.id"
                            :label="template.name"/>
                </t-select>
              </t-form-item>
            </t-col>
            <t-col :span="12">
              <t-form-item label="Code" name="placeholder">
                <t-textarea v-model="formData.placeholder" rows="4"/>
              </t-form-item>
            </t-col>
            <t-col :span="12">
              <t-form-item label="Editable" name="editable">
                <t-switch v-model="formData.editable"></t-switch>
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
  type CreateSandboxReq,
  type UpdateSandboxReq,
  createSandbox,
  updateSandbox,
  getSandbox,
} from "@/api/sandbox";
import {type Template, allTemplates} from "@/api/template";

const route = useRoute()
const router = useRouter()
const mode = ref<'create' | 'edit'>(route.name === 'createSandbox' ? 'create' : 'edit')
const id = ref<string | undefined>(route.params.id as string | undefined)

const FORM_RULES: FormProps['rules'] = {
  name: [{required: true, message: 'Name is required'}],
  templateID: [{required: true, message: 'Template is required'}],
};

const formData = ref<CreateSandboxReq | UpdateSandboxReq>({} as CreateSandboxReq)
const onSubmit = (ctx: SubmitContext) => {
  if (ctx.validateResult === true) {
    if (mode.value === 'create') {
      createSandbox(formData.value).then(() => {
        MessagePlugin.success('Create new sandbox successfully')
        router.push({name: 'sandbox'})
      })
    } else {
      updateSandbox(id.value, formData.value).then(res => {
        MessagePlugin.success(res)
      })
    }
  }
}

const onCancel = () => {
  router.push({name: 'sandbox'})
}

const templates = ref<Template[]>([])

onMounted(() => {
  allTemplates().then(res => {
    templates.value = res
  })

  if (mode.value === 'edit') {
    getSandbox(id.value).then(res => {
      formData.value = {
        name: res.name,
        templateID: res.templateID,
        placeholder: res.placeholder,
        editable: res.editable,
      }
    })
  }
})
</script>

<style scoped>

</style>
