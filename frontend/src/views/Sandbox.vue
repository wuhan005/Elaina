<template>
  <div>
    <section class="hero">
      <div class="hero-body">
        <p class="title">沙箱</p>
        <p class="subtitle">沙箱为基于运行模板创建的独立运行环境。</p>
        <b-button type="is-primary is-light" @click="createModalVisible = true">添加沙箱</b-button>
        <b-table :data="list" :loading="isLoading" striped hoverable>
          <b-table-column label="UID" v-slot="props">{{ props.row.uid }}</b-table-column>
          <b-table-column label="沙箱名称" v-slot="props">{{ props.row.name }}</b-table-column>
          <b-table-column label="所属模板" v-slot="props">{{ props.row.template.name }}</b-table-column>
          <b-table-column label="初始内容" v-slot="props">{{ props.row.placeholder }}</b-table-column>
          <b-table-column label="是否可编辑" v-slot="props">{{ props.row.editable }}</b-table-column>
          <b-table-column v-slot="props">
            <b-button type="is-light" @click="()=>{
              updateSandboxForm = JSON.parse(JSON.stringify(props.row));
              detailModalVisible = true;
            }">操作
            </b-button>
          </b-table-column>
        </b-table>
      </div>
    </section>

    <!-- Create sandbox modal-->
    <b-modal
        v-model="createModalVisible" has-modal-card trap-focus :destroy-on-hide="false" aria-role="dialog"
        aria-label="添加沙箱" aria-modal>
      <template #default="props">
        <section>
          <div class="modal-card">
            <header class="modal-card-head">
              <p class="modal-card-title">添加沙箱</p>
              <button type="button" class="delete" @click="$emit('close')"/>
            </header>
            <section class="modal-card-body">
              <b-field label="模板名称">
                <b-input v-model="newSandboxForm.name" required></b-input>
              </b-field>
              <b-field label="运行模板">
                <b-select v-model="newSandboxForm.template_id">
                  <option
                      v-for="(template, index) in templateList"
                      :value="template.ID"
                      :key="index">
                    {{ template.name }}
                  </option>
                </b-select>
              </b-field>
              <b-field label="初始内容">
                <b-input type="textarea" v-model="newSandboxForm.placeholder"></b-input>
              </b-field>
              <b-field label="是否可编辑">
                <b-switch v-model="newSandboxForm.editable"></b-switch>
              </b-field>
            </section>
            <footer class="modal-card-foot">
              <b-button label="关闭" @click="$emit('close')"/>
              <b-button label="添加沙箱" type="is-primary" @click="newSandbox"/>
            </footer>
          </div>
        </section>
      </template>
    </b-modal>

    <!-- Detail template modal -->
    <b-modal
        v-model="detailModalVisible" has-modal-card trap-focus :destroy-on-hide="false" aria-role="dialog"
        aria-label="操作沙箱" aria-modal>
      <template #default="props">
        <section>
          <div class="modal-card">
            <header class="modal-card-head">
              <p class="modal-card-title">操作沙箱</p>
              <button type="button" class="delete" @click="$emit('close')"/>
            </header>
            <section class="modal-card-body">
              <b-field label="模板名称">
                <b-input v-model="updateSandboxForm.name" required></b-input>
              </b-field>
              <b-field label="运行模板">
                <b-select v-model="updateSandboxForm.template_id">
                  <option
                      v-for="(template, index) in templateList"
                      :value="template.ID"
                      :key="index">
                    {{ template.name }}
                  </option>
                </b-select>
              </b-field>
              <b-field label="初始内容">
                <b-input type="textarea" v-model="updateSandboxForm.placeholder"></b-input>
              </b-field>
              <b-field label="是否可编辑">
                <b-switch v-model="updateSandboxForm.editable"></b-switch>
              </b-field>
            </section>
            <footer class="modal-card-foot">
              <b-button label="删除沙箱" type="is-danger" @click="deleteSandbox(updateSandboxForm)"/>
              <b-button label="修改沙箱" type="is-primary" @click="updateSandbox"/>
            </footer>
          </div>
        </section>
      </template>
    </b-modal>

  </div>
</template>

<script>
export default {
  name: "Sandbox",
  data() {
    return {
      list: [],
      templateList: [],

      newSandboxForm: {
        name: '',
        template_id: 0,
        placeholder: '',
        editable: true,
      },
      updateSandboxForm: {},

      isLoading: true,
      createModalVisible: false,
      detailModalVisible: false,
    }
  },
  mounted() {
    this.fetchTemplateList()
    this.fetchList()
  },

  methods: {
    fetchList() {
      this.utils.GET('/m/sandboxes').then(res => {
        this.list = res;
        this.isLoading = false;
      }).catch(err => this.$buefy.toast.open({message: err.response.data.msg, type: 'is-danger'}))
    },

    fetchTemplateList() {
      this.utils.GET('/m/templates').then(res => {
        this.templateList = res
      }).catch(err => this.$buefy.toast.open({message: err.response.data.msg, type: 'is-danger'}))
    },

    newSandbox() {
      this.utils.POST('/m/sandbox', this.newSandboxForm)
          .then(res => {
            this.createModalVisible = false
            this.newSandboxForm = {
              name: '',
              template_id: 0,
              placeholder: '',
              editable: true,
            }
            this.fetchList()
            this.$buefy.toast.open({message: res, type: 'is-success'})
          })
          .catch(err => this.$buefy.toast.open({message: err.response.data.msg, type: 'is-danger'}))
    },

    updateSandbox() {
      this.utils.PUT('/m/sandbox', this.updateSandboxForm)
          .then(res => {
            this.detailModalVisible = false
            this.fetchList()
            this.$buefy.toast.open({message: res, type: 'is-success'})
          })
          .catch(err => this.$buefy.toast.open({message: err.response.data.msg, type: 'is-danger'}))
    },

    deleteSandbox(template) {
      this.$buefy.dialog.confirm({
        title: '删除模板',
        message: `您确认要<b>删除</b> [ ${template.name} ] 吗？`,
        confirmText: '确认删除',
        type: 'is-danger',
        hasIcon: true,
        onConfirm: () => {
          this.utils.DELETE('/m/sandbox?id=' + template.ID)
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