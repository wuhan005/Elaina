<template>
  <div>
    <section class="hero">
      <div class="hero-body">
        <p class="title">Template</p>
        <p class="subtitle">The template defines the resource limited and strategy of the sandbox.</p>
        <b-button type="is-primary is-light" @click="createModalVisible = true">New Template</b-button>
        <b-table :data="list" :loading="isLoading" striped hoverable>
          <b-table-column label="Name" v-slot="props">{{ props.row.name }}</b-table-column>
          <b-table-column label="Programming Languages" v-slot="props">
            <b-taglist>
              <b-tag v-for="(tag, index) in props.row.language.Elements" v-bind:key="index">
                {{ tag }}
              </b-tag>
            </b-taglist>
          </b-table-column>
          <b-table-column label="Time Limit (s)" v-slot="props">{{ props.row.timeout }}</b-table-column>
          <b-table-column label="Maximum CPUs" v-slot="props">{{ props.row.max_cpus }}</b-table-column>
          <b-table-column label="Maximum Memory (MB)" v-slot="props">{{ props.row.max_memory }}</b-table-column>
          <b-table-column label="Internet Access" v-slot="props">{{ props.row.internet_access }}</b-table-column>
          <b-table-column label="Maximum Number of Containers" v-slot="props">{{ props.row.max_container }}</b-table-column>
          <b-table-column label="Maximum Number of Containers per IP" v-slot="props">{{ props.row.max_container_per_ip }}</b-table-column>
          <b-table-column v-slot="props">
            <b-button type="is-light" @click="()=>{
              updateTemplateForm = JSON.parse(JSON.stringify(props.row));
              updateTemplateForm.language = updateTemplateForm.language.Elements;
              detailModalVisible = true;
            }">Edit
            </b-button>
          </b-table-column>
        </b-table>
      </div>
    </section>

    <!-- Create template modal-->
    <b-modal
        v-model="createModalVisible" has-modal-card trap-focus :destroy-on-hide="false" aria-role="dialog"
        aria-label="New Template" aria-modal>
      <template #default="props">
        <section>
          <div class="modal-card">
            <header class="modal-card-head">
              <p class="modal-card-title">New Template</p>
              <button type="button" class="delete" @click="$emit('close')"/>
            </header>
            <section class="modal-card-body">
              <b-field label="Name">
                <b-input v-model="newTemplateForm.name" required></b-input>
              </b-field>
              <b-field label="Programming Languages">
                <div class="block">
                  <b-checkbox v-model="newTemplateForm.language" native-value="php">PHP</b-checkbox>
                  <b-checkbox v-model="newTemplateForm.language" native-value="python">Python</b-checkbox>
                  <b-checkbox v-model="newTemplateForm.language" native-value="go">Go</b-checkbox>
                  <b-checkbox v-model="newTemplateForm.language" native-value="javascript">JavaScript</b-checkbox>
                </div>
              </b-field>
              <b-field label="Time Limit (s)">
                <b-numberinput v-model="newTemplateForm.timeout" :min="0" :max="60"></b-numberinput>
              </b-field>
              <b-field label="Maximum CPUs">
                <b-numberinput v-model="newTemplateForm.max_cpus" :min="0" :max="10"></b-numberinput>
              </b-field>
              <b-field label="Maximum Memory (MB)">
                <b-numberinput v-model="newTemplateForm.max_memory" :min="6" :max="2048"></b-numberinput>
              </b-field>
              <b-field label="Internet Access">
                <b-switch v-model="newTemplateForm.internet_access"></b-switch>
              </b-field>
              <b-field label="">
              </b-field>
              <b-field label="Maximum Number of Containers">
                <b-numberinput v-model="newTemplateForm.max_container" :min="0" :max="2048"></b-numberinput>
              </b-field>
              <b-field label="Maximum Number of Containers per IP">
                <b-numberinput v-model="newTemplateForm.max_container_per_ip" :min="0" :max="2048"></b-numberinput>
              </b-field>
            </section>
            <footer class="modal-card-foot">
              <b-button label="Close" @click="$emit('close')"/>
              <b-button label="Submit" type="is-primary" @click="newTemplate"/>
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
              <p class="modal-card-title">Edit Template</p>
              <button type="button" class="delete" @click="$emit('close')"/>
            </header>
            <section class="modal-card-body">
              <b-field label="Name">
                <b-input v-model="updateTemplateForm.name" required></b-input>
              </b-field>
              <b-field label="Programming Languages">
                <div class="block">
                  <b-checkbox v-model="updateTemplateForm.language" native-value="php">PHP</b-checkbox>
                  <b-checkbox v-model="updateTemplateForm.language" native-value="python">Python</b-checkbox>
                  <b-checkbox v-model="updateTemplateForm.language" native-value="go">Go</b-checkbox>
                  <b-checkbox v-model="updateTemplateForm.language" native-value="javascript">JavaScript</b-checkbox>
                </div>
              </b-field>
              <b-field label="Time Limit (s)">
                <b-numberinput v-model="updateTemplateForm.timeout" :min="0" :max="60"></b-numberinput>
              </b-field>
              <b-field label="Maximum CPUs">
                <b-numberinput v-model="updateTemplateForm.max_cpus" :min="0" :max="10"></b-numberinput>
              </b-field>
              <b-field label="Maximum Memory (MB)">
                <b-numberinput v-model="updateTemplateForm.max_memory" :min="6" :max="2048"></b-numberinput>
              </b-field>
              <b-field label="Internet Access">
                <b-switch v-model="updateTemplateForm.internet_access"></b-switch>
              </b-field>
              <b-field label="">
              </b-field>
              <b-field label="Maximum Number of Containers">
                <b-numberinput v-model="updateTemplateForm.max_container" :min="0" :max="1000"></b-numberinput>
              </b-field>
              <b-field label="Maximum Number of Containers per IP">
                <b-numberinput v-model="updateTemplateForm.max_container_per_ip" :min="0" :max="100"></b-numberinput>
              </b-field>
            </section>
            <footer class="modal-card-foot">
              <b-button label="Delete" type="is-danger" @click="deleteTemplate(updateTemplateForm)"/>
              <b-button label="Edit" type="is-primary" @click="updateTemplate"/>
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
        title: 'Delete Template',
        message: `Are you sure to <b>DELETE</b> [ ${template.name} ] ?`,
        confirmText: 'Yes, delete it.',
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