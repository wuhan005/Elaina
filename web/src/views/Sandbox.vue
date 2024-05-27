<template>
  <div>
    <t-card class="list-card-container" :bordered="false">
      <t-row justify="space-between">
        <div class="left-operation-container">
          <t-button @click="onCreate"> New Sandbox</t-button>
        </div>
      </t-row>
      <t-table
          :data="data"
          :columns="COLUMNS"
          row-key="id"
          vertical-align="top"
          :hover="true"
          :pagination="pagination"
          :loading="isLoading"
          @page-change="pagination = $event; getSandboxes()"
      >
        <template #editable="{row}">
          <t-tag :theme="row.editable ? 'primary' : 'danger'" variant="outline">
            {{ row.editable ? 'Yes' : 'No' }}
          </t-tag>
        </template>
        <template #template="{row}">
          {{ row.template.name }}
        </template>
        <template #placeholder="{row}">
          <code>{{ row.placeholder }}</code>
        </template>
        <template #createdAt="{row}">
          {{ dayjs(row.createdAt).format('YYYY-MM-DD HH:mm:ss') }}
        </template>
        <template #op="{row}">
          <t-space>
            <t-link theme="primary" :href="`/r/${row.uid}`" target="_blank">VIEW</t-link>
            <t-link theme="primary" @click="onUpdate(row.id)">EDIT</t-link>
            <t-popconfirm content="Are you sure you want to delete this sandbox?" @confirm="onDelete(row.id)">
              <t-link theme="danger">DELETE</t-link>
            </t-popconfirm>
          </t-space>
        </template>
      </t-table>
    </t-card>
  </div>
</template>

<script setup lang="ts">
import {onMounted, ref} from 'vue';
import {MessagePlugin, PrimaryTableCol, TableRowData, type PaginationProps} from 'tdesign-vue-next';
import dayjs from 'dayjs'
import {type Sandbox, listSandboxes, deleteSandbox} from "@/api/sandbox";
import {type Template, listTemplates, deleteTemplate} from "@/api/template";
import {useRouter} from "vue-router";

const COLUMNS: PrimaryTableCol<TableRowData>[] = [
  {title: 'UID', colKey: 'uid'},
  {title: 'Name', colKey: 'name'},
  {title: 'Template', colKey: 'template'},
  {title: 'Editable', colKey: 'editable'},
  {title: 'Placeholder', colKey: 'placeholder', ellipsis: true, width: 250},
  {title: 'Created At', ellipsis: true, width: 180, colKey: 'createdAt'},
  {title: 'Operators', align: 'left', fixed: 'right', width: 160, colKey: 'op'},
];
const pagination = ref<PaginationProps>({
  pageSize: 20,
  current: 1,
})
const isLoading = ref<boolean>(false);
const data = ref<Sandbox[]>([]);
const router = useRouter()

const getSandboxes = () => {
  isLoading.value = true

  listSandboxes({page: pagination.value.current, pageSize: pagination.value.pageSize}).then(res => {
    data.value = res.sandboxes
    pagination.value.total = res.total
  }).finally(() => {
    isLoading.value = false
  })
}

const onCreate = () => {
  router.push({name: 'createSandbox'})
}

const onUpdate = (id: number) => {
  router.push({name: 'editSandbox', params: {id: id.toString()}})
}

const onDelete = (id: number) => {
  deleteSandbox(id).then(res => {
    MessagePlugin.success(res)
  }).finally(() => {
    getSandboxes()
  })
}

onMounted(() => {
  getSandboxes()
})
</script>

<style scoped lang="less">
.list-card-container {
  padding: var(--td-comp-paddingTB-xxl) var(--td-comp-paddingLR-xxl);

  :deep(.t-card__body) {
    padding: 0;
  }
}

.left-operation-container {
  display: flex;
  align-items: center;
  margin-bottom: var(--td-comp-margin-xxl);

  .selected-count {
    display: inline-block;
    margin-left: var(--td-comp-margin-l);
    color: var(--td-text-color-secondary);
  }
}

.search-input {
  width: 360px;
}
</style>
