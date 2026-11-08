<template>
  <div>
    <t-card class="list-card-container" :bordered="false">
      <t-row justify="space-between">
        <div class="left-operation-container">
          <t-button @click="onCreate"> New Template</t-button>
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
          @page-change="pagination = $event; getTemplates()"
      >
        <template #name="{row}">
          <t-space direction="vertical" :size="4">
            <span>{{ row.name }}</span>
            <t-tag v-if="row.maxCpus !== 0" theme="warning" variant="outline">Legacy CPU quota: edit required</t-tag>
          </t-space>
        </template>
        <template #language="{row}">
          <t-space :size="2">
            <t-tag variant="outline" v-for="item in row.language" v-bind:key="item">
              {{ LANGUAGES_MAP[item] }}
            </t-tag>
          </t-space>
        </template>
        <template #timeout="{row}">
          {{ row.timeout === 0 ? '10 (default)' : row.timeout }}
        </template>
        <template #maxMemory="{row}">
          {{ row.maxMemory === 0 ? '4096 (WASM maximum)' : row.maxMemory }}
        </template>
        <template #internetAccess="{row}">
          <t-tag :theme="row.internetAccess ? 'primary' : 'danger'" variant="outline">
            {{ row.internetAccess ? 'Allowed' : 'Denied' }}
          </t-tag>
        </template>
        <template #createdAt="{row}">
          {{ dayjs(row.createdAt).format('YYYY-MM-DD HH:mm:ss') }}
        </template>
        <template #op="{row}">
          <t-space>
            <t-link theme="primary" @click="onUpdate(row.id)">EDIT</t-link>
            <t-popconfirm content="Are you sure you want to delete this template?" @confirm="onDelete(row.id)">
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
import {type Template, listTemplates, deleteTemplate} from "@/api/template";
import {useRouter} from "vue-router";
import {LANGUAGES_MAP} from "@/const/template";

const COLUMNS: PrimaryTableCol<TableRowData>[] = [
  {title: 'ID', colKey: 'id',},
  {title: 'Name', colKey: 'name',},
  {title: 'Programming Languages', colKey: 'language'},
  {title: 'Execution Timeout (s)', colKey: 'timeout'},
  {title: 'WASM Linear Memory (MiB)', colKey: 'maxMemory'},
  {title: 'Network Permission', colKey: 'internetAccess'},
  {title: 'Created At', ellipsis: true, width: 180, colKey: 'createdAt'},
  {title: 'Operators', align: 'left', fixed: 'right', width: 160, colKey: 'op'},
];
const pagination = ref<PaginationProps>({
  pageSize: 20,
  current: 1,
})
const isLoading = ref<boolean>(false);
const data = ref<Template[]>([]);
const router = useRouter()

const getTemplates = () => {
  isLoading.value = true

  listTemplates({page: pagination.value.current, pageSize: pagination.value.pageSize}).then(res => {
    data.value = res.templates
    pagination.value.total = res.total
  }).finally(() => {
    isLoading.value = false
  })
}

const onCreate = () => {
  router.push({name: 'createTemplate'})
}

const onUpdate = (id: number) => {
  router.push({name: 'editTemplate', params: {id: id.toString()}})
}

const onDelete = (id: number) => {
  deleteTemplate(id).then(res => {
    MessagePlugin.success(res)
  }).finally(() => {
    getTemplates()
  })
}

onMounted(() => {
  getTemplates()
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
