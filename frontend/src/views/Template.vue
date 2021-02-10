<template>
  <div>
    <at-button type="primary" @click="createModelVisible = true" style="float: right">添加模板</at-button>
    <h1>运行模板</h1>
    <at-table style="margin-top: 10px" :columns="columns" :data="list" stripe></at-table>

    <!-- Create template modal-->
    <at-modal v-model="createModelVisible" @on-confirm="newTemplate">
      <h1>添加模板</h1>
      <div style="margin-top: 10px">
        <span>模板名称</span>
        <at-input v-model="newTemplateForm.name"></at-input>
        <br>
        <span>编程语言</span><br>
        <at-checkbox-group v-model="newTemplateForm.language">
          <at-checkbox label="php">PHP</at-checkbox>
          <at-checkbox label="python">Python</at-checkbox>
          <at-checkbox label="go">Go</at-checkbox>
        </at-checkbox-group>
        <br><br>
        <span>超时时间</span>
        <at-input-number v-model="newTemplateForm.timeout" :min="0" :max="60"></at-input-number>
        <br><br>
        <span>最大 CPU 数</span>
        <at-input-number v-model="newTemplateForm.max_cpus" :min="0" :max="10"></at-input-number>
        <br><br>
        <span>最大内存 (MB)</span>
        <at-input-number v-model="newTemplateForm.max_memory" :min="6" :max="2048"></at-input-number>
        <br><br>
        <span>开放外网</span><br>
        <at-switch v-model="newTemplateForm.internet_access"></at-switch>
        <br><br>
        <span>自定义 DNS 解析</span>
        <br><br>
        <span>最大容器数</span>
        <at-input-number v-model="newTemplateForm.max_container" :min="0" :max="1000"></at-input-number>
        <br><br>
        <span>单 IP 最大容器数</span>
        <at-input-number v-model="newTemplateForm.max_container_per_ip" :min="0" :max="100"></at-input-number>
      </div>
      <div slot="footer">
        <at-button type="primary" @click="newTemplate">添加</at-button>
      </div>
    </at-modal>

  </div>
</template>

<script>
export default {
  name: "Template",
  data() {
    return {
      columns: [
        {title: '模板名称', key: 'name'},
        {title: '编程语言', key: 'language'},
        {title: '超时时间', key: 'timeout'},
        {title: '最大 CPU 数', key: 'max_cpus'},
        {title: '最大内存', key: 'max_memory'},
        {title: '开放外网', key: 'internet_access'},
        {title: '最大容器数', key: 'max_container'},
        {title: '单 IP 最大容器数', key: 'max_container_per_ip'},
      ],
      list: [],

      newTemplateForm: {
        name: '',
        language: [],
        timeout: 0,
        max_cpus: 1,
        max_memory: 1024,
        internet_access: false,
        dns: {},
        max_container: 100,
        max_container_per_ip: 5,
      },

      createModelVisible: false,
    }
  },
  mounted() {
    this.fetchList()
  },

  methods: {
    fetchList() {
      this.utils.GET('/m/templates').then(res => {
        this.list = res
      }).catch(err => this.$Message.error(err.response.data.msg))
    },
    newTemplate() {
      this.utils.POST('/m/template', this.newTemplateForm)
          .then(res => this.$Message.success(res))
          .catch(err => this.$Message.error(err.response.data.msg))
    }
  }
}
</script>

<style scoped>

</style>