<template>
  <div>
    <section class="hero">
      <div class="hero-body">
        <p class="title">Sandbox</p>
        <p class="subtitle">The sandbox is an independent environment created based on the template.</p>
        <b-button type="is-primary is-light" @click="createModalVisible = true">New Sandbox</b-button>
        <b-table :data="list" :loading="isLoading" striped hoverable>
          <b-table-column label="UID" v-slot="props">{{ props.row.uid }}</b-table-column>
          <b-table-column label="Name" v-slot="props">{{ props.row.name }}</b-table-column>
          <b-table-column label="Template" v-slot="props">{{ props.row.template.name }}</b-table-column>
          <b-table-column label="Code" v-slot="props">{{ props.row.placeholder }}</b-table-column>
          <b-table-column label="Editable" v-slot="props">{{ props.row.editable }}</b-table-column>
          <b-table-column v-slot="props">
            <b-button type="is-light" @click="()=>{
              updateSandboxForm = JSON.parse(JSON.stringify(props.row));
              detailModalVisible = true;
            }">Edit
            </b-button>
          </b-table-column>
        </b-table>
      </div>
    </section>

    <!-- Create sandbox modal-->
    <b-modal
        v-model="createModalVisible" has-modal-card trap-focus :destroy-on-hide="false" aria-role="dialog"
        aria-label="New Sandbox" aria-modal>
      <template #default="props">
        <section>
          <div class="modal-card">
            <header class="modal-card-head">
              <p class="modal-card-title">New Sandbox</p>
              <button type="button" class="delete" @click="$emit('close')"/>
            </header>
            <section class="modal-card-body">
              <b-field label="Name">
                <b-input v-model="newSandboxForm.name" required></b-input>
              </b-field>
              <b-field label="Template">
                <b-select v-model="newSandboxForm.template_id">
                  <option
                      v-for="(template, index) in templateList"
                      :value="template.ID"
                      :key="index">
                    {{ template.name }}
                  </option>
                </b-select>
              </b-field>
              <b-field label="Code">
                <b-input type="textarea" v-model="newSandboxForm.placeholder"></b-input>
              </b-field>
              <b-field label="Editable">
                <b-switch v-model="newSandboxForm.editable"></b-switch>
              </b-field>
            </section>
            <footer class="modal-card-foot">
              <b-button label="Close" @click="$emit('close')"/>
              <b-button label="Submit" type="is-primary" @click="newSandbox"/>
            </footer>
          </div>
        </section>
      </template>
    </b-modal>

    <!-- Detail template modal -->
    <b-modal
        v-model="detailModalVisible" has-modal-card trap-focus :destroy-on-hide="false" aria-role="dialog"
        aria-label="Edit" aria-modal>
      <template #default="props">
        <section>
          <div class="modal-card">
            <header class="modal-card-head">
              <p class="modal-card-title">Edit Sandbox</p>
              <button type="button" class="delete" @click="$emit('close')"/>
            </header>
            <section class="modal-card-body">
              <b-field label="Name">
                <b-input v-model="updateSandboxForm.name" required></b-input>
              </b-field>
              <b-field label="Template">
                <b-select v-model="updateSandboxForm.template_id">
                  <option
                      v-for="(template, index) in templateList"
                      :value="template.ID"
                      :key="index">
                    {{ template.name }}
                  </option>
                </b-select>
              </b-field>
              <b-field label="Code">
                <b-input type="textarea" v-model="updateSandboxForm.placeholder"></b-input>
              </b-field>
              <b-field label="Editable">
                <b-switch v-model="updateSandboxForm.editable"></b-switch>
              </b-field>
            </section>
            <footer class="modal-card-foot">
              <b-button label="Delete" type="is-danger" @click="deleteSandbox(updateSandboxForm)"/>
              <b-button label="Edit" type="is-primary" @click="updateSandbox"/>
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
        title: 'Delete Sandbox',
        message: `Are you sure to <b>DELETE</b> [ ${template.name} ]?`,
        confirmText: 'Yes, delete it.',
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