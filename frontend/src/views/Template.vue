<template>
  <div>
    <section class="hero">
      <div class="hero-body">
        <p class="title">运行模板</p>
        <p class="subtitle">运行模板定义了沙箱的基本资源限制、访问策略。</p>
        <b-button type="is-primary is-light" @click="createModalVisible = true">添加模板</b-button>
        <b-table :data="list" :loading="isLoading" striped hoverable>
          <b-table-column label="模板名称" v-slot="props">{{ props.row.name }}</b-table-column>
          <b-table-column label="编程语言" v-slot="props">
            <b-taglist>
              <b-tag v-for="(tag, index) in props.row.language.Elements" v-bind:key="index">
                {{ tag }}
              </b-tag>
            </b-taglist>
          </b-table-column>
          <b-table-column label="超时时间 (s)" v-slot="props">{{ props.row.timeout }}</b-table-column>
          <b-table-column label="最大 CPU 数" v-slot="props">{{ props.row.max_cpus }}</b-table-column>
          <b-table-column label="最大内存 (MB)" v-slot="props">{{ props.row.max_memory }}</b-table-column>
          <b-table-column label="开放外网" v-slot="props">{{ props.row.internet_access }}</b-table-column>
          <b-table-column label="最大容器数" v-slot="props">{{ props.row.max_container }}</b-table-column>
          <b-table-column label="单 IP 最大容器数" v-slot="props">{{ props.row.max_container_per_ip }}</b-table-column>
          <b-table-column v-slot="props">
            <b-button type="is-light" @click="()=>{
              updateTemplateForm = JSON.parse(JSON.stringify(props.row));
              updateTemplateForm.language = updateTemplateForm.language.Elements;
              detailModalVisible = true;
            }">操作
            </b-button>
          </b-table-column>
        </b-table>
      </div>
    </section>

    <!-- Create template modal-->
    <b-modal
        v-model="createModalVisible" has-modal-card trap-focus :destroy-on-hide="false" aria-role="dialog"
        aria-label="添加模板" aria-modal>
      <template #default="props">
        <section>
          <div class="modal-card">
            <header class="modal-card-head">
              <p class="modal-card-title">添加模板</p>
              <button type="button" class="delete" @click="$emit('close')"/>
            </header>
            <section class="modal-card-body">
              <b-field label="模板名称">
                <b-input v-model="newTemplateForm.name" required></b-input>
              </b-field>
              <b-field label="编程语言">
                <div class="block">
                  <b-checkbox v-model="newTemplateForm.language" native-value="php">PHP</b-checkbox>
                  <b-checkbox v-model="newTemplateForm.language" native-value="python">Python</b-checkbox>
                  <b-checkbox v-model="newTemplateForm.language" native-value="go">Go</b-checkbox>
                  <b-checkbox v-model="newTemplateForm.language" native-value="javascript">JavaScript</b-checkbox>
                </div>
              </b-field>
              <b-field label="超时时间 (s)">
                <b-numberinput v-model="newTemplateForm.timeout" :min="0" :max="60"></b-numberinput>
              </b-field>
              <b-field label="最大 CPU 数">
                <b-numberinput v-model="newTemplateForm.max_cpus" :min="0" :max="10"></b-numberinput>
              </b-field>
              <b-field label="最大内存 (MB)">
                <b-numberinput v-model="newTemplateForm.max_memory" :min="6" :max="2048"></b-numberinput>
              </b-field>
              <b-field label="开放外网">
                <b-switch v-model="newTemplateForm.internet_access"></b-switch>
              </b-field>
              <b-field label="自定义 DNS 解析">
              </b-field>
              <b-field label="最大容器数">
                <b-numberinput v-model="newTemplateForm.max_container" :min="6" :max="2048"></b-numberinput>
              </b-field>
              <b-field label="单 IP 最大容器数">
                <b-numberinput v-model="newTemplateForm.max_container_per_ip" :min="6" :max="2048"></b-numberinput>
              </b-field>
            </section>
            <footer class="modal-card-foot">
              <b-button label="关闭" @click="$emit('close')"/>
              <b-button label="添加模板" type="is-primary" @click="newTemplate"/>
            </footer>
          </div>
        </section>
      </template>
    </b-modal>

    <!-- Detail template modal -->
    <b-modal
        v-model="detailModalVisible" has-modal-card trap-focus :destroy-on-hide="false" aria-role="dialog"
        aria-label="操作模板" aria-modal>
      <template #default="props">
        <section>
          <div class="modal-card">
            <header class="modal-card-head">
              <p class="modal-card-title">操作模板</p>
              <button type="button" class="delete" @click="$emit('close')"/>
            </header>
            <section class="modal-card-body">
              <b-field label="模板名称">
                <b-input v-model="updateTemplateForm.name" required></b-input>
              </b-field>
              <b-field label="编程语言">
                <div class="block">
                  <b-checkbox v-model="updateTemplateForm.language" native-value="php">PHP</b-checkbox>
                  <b-checkbox v-model="updateTemplateForm.language" native-value="python">Python</b-checkbox>
                  <b-checkbox v-model="updateTemplateForm.language" native-value="go">Go</b-checkbox>
                  <b-checkbox v-model="updateTemplateForm.language" native-value="javascript">JavaScript</b-checkbox>
                </div>
              </b-field>
              <b-field label="超时时间 (s)">
                <b-numberinput v-model="updateTemplateForm.timeout" :min="0" :max="60"></b-numberinput>
              </b-field>
              <b-field label="最大 CPU 数">
                <b-numberinput v-model="updateTemplateForm.max_cpus" :min="0" :max="10"></b-numberinput>
              </b-field>
              <b-field label="最大内存 (MB)">
                <b-numberinput v-model="updateTemplateForm.max_memory" :min="6" :max="2048"></b-numberinput>
              </b-field>
              <b-field label="开放外网">
                <b-switch v-model="updateTemplateForm.internet_access"></b-switch>
              </b-field>
              <b-field label="自定义 DNS 解析">
              </b-field>
              <b-field label="最大容器数">
                <b-numberinput v-model="updateTemplateForm.max_container" :min="6" :max="2048"></b-numberinput>
              </b-field>
              <b-field label="单 IP 最大容器数">
                <b-numberinput v-model="updateTemplateForm.max_container_per_ip" :min="6" :max="2048"></b-numberinput>
              </b-field>
            </section>
            <footer class="modal-card-foot">
              <b-button label="删除模板" type="is-danger" @click="deleteTemplate(updateTemplateForm)"/>
              <b-button label="修改模板" type="is-primary" @click="updateTemplate"/>
            </footer>
          </div>
        </section>
      </template>
    </b-modal>

  </div>
</template>

<script>
export default {
  name: "Template",
  data() {
    return {
      list: [],

      newTemplateForm: {
        name: '',
        language: [],
        timeout: 5,
        max_cpus: 1,
        max_memory: 30,
        internet_access: false,
        dns: {},
        max_container: 100,
        max_container_per_ip: 5,
      },
      updateTemplateForm: {},

      isLoading: true,
      createModalVisible: false,
      detailModalVisible: false,
    }
  },
  mounted() {
    this.fetchList()
  },

  methods: {
    fetchList() {
      this.isLoading = true
      this.utils.GET('/m/templates').then(res => {
        this.list = res
        this.isLoading = false
      }).catch(err => this.$buefy.toast.open({message: err.response.data.msg, type: 'is-danger'}))
    },

    newTemplate() {
      this.utils.POST('/m/template', this.newTemplateForm)
          .then(res => {
            this.createModalVisible = false
            this.newTemplateForm = {
              name: '',
              language: [],
              timeout: 5,
              max_cpus: 1,
              max_memory: 30,
              internet_access: false,
              dns: {},
              max_container: 100,
              max_container_per_ip: 5,
            }
            this.fetchList()
            this.$buefy.toast.open({message: res, type: 'is-success'})
          })
          .catch(err => this.$buefy.toast.open({message: err.response.data.msg, type: 'is-danger'}))
    },

    updateTemplate() {
      this.utils.PUT('/m/template', this.updateTemplateForm)
          .then(res => {
            this.detailModalVisible = false
            this.fetchList()
            this.$buefy.toast.open({message: res, type: 'is-success'})
          })
          .catch(err => this.$buefy.toast.open({message: err.response.data.msg, type: 'is-danger'}))
    },

    deleteTemplate(template) {
      this.$buefy.dialog.confirm({
        title: '删除模板',
        message: `您确认要<b>删除</b> [ ${template.name} ] 吗？`,
        confirmText: '确认删除',
        type: 'is-danger',
        hasIcon: true,
        onConfirm: () => {
          this.utils.DELETE('/m/template?id=' + template.ID)
              .then(res => {
                this.$buefy.toast.open({message: res, type: 'is-success'});
                this.detailModalVisible = false;
                this.fetchList()
              })
              .catch(err => this.$buefy.toast.open({message: err.response.data.msg, type: 'is-danger'}))
        }
      })
    }
  }
}
</script>

<style scoped>

</style>