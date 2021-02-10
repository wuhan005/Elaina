<template>
  <div>
    <at-button type="primary" @click="createModelVisible = true" style="float: right">添加沙箱</at-button>
    <h1>沙箱</h1>

    <at-table style="margin-top: 10px" :columns="columns" :data="list" stripe></at-table>

    <!-- Create sandbox modal-->
    <at-modal v-model="createModelVisible" @on-confirm="newSandbox">
      <h1>添加沙箱</h1>
      <div style="margin-top: 10px">
        <span>所属模板</span><br>
        <at-select v-model="newSandboxForm.template_id" style="width:130px">
          <at-option v-for="(tmpl, index) in templateList" v-bind:key="index" :value="tmpl.ID">
            {{ tmpl.name }}
          </at-option>
        </at-select>
        <br><br>
        <span>初始内容</span>
        <at-textarea v-model="newSandboxForm.placeholder" autosize></at-textarea>
        <br><br>
        <span>是否可编辑</span><br>
        <at-switch v-model="newSandboxForm.editable"></at-switch>
      </div>
      <div slot="footer">
        <at-button type="primary" @click="newSandbox">添加</at-button>
      </div>
    </at-modal>
  </div>
</template>

<script>
export default {
  name: "Sandbox",
  data() {
    return {
      columns: [
        {title: 'UID', key: 'uid'},
        {title: '所属模板', key: 'template_id'},
        {title: '模板', key: 'template'},
        {title: '初始内容', key: 'placeholder'},
        {title: '是否可编辑', key: 'editable'}
      ],
      list: [],
      templateList: [],

      newSandboxForm: {
        template_id: 0,
        placeholder: '',
        editable: true,
      },

      createModelVisible: false,
    }
  },
  mounted() {
    this.fetchTemplateList()
    this.fetchList()
  },

  methods: {
    fetchList() {
      this.utils.GET('/m/sandboxes').then(res => {
        this.list = res
      }).catch(err => this.$Message.error(err.response.data.msg))
    },
    fetchTemplateList() {
      this.utils.GET('/m/templates').then(res => {
        this.templateList = res
      }).catch(err => this.$Message.error(err.response.data.msg))
    },
    newSandbox() {
      this.utils.POST('/m/sandbox', this.newSandboxForm)
          .then(res => this.$Message.success(res))
          .catch(err => this.$Message.error(err.response.data.msg))
    }
  }
}
</script>

<style scoped>

</style>