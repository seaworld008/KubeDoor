<script setup lang="ts">
import { ref } from "vue";
import { transformI18n } from "@/plugins/i18n";
import { useResource } from "./utils/hook";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";

import Upload from "@iconify-icons/ri/upload-line";
import Role from "@iconify-icons/ri/admin-line";
import Password from "@iconify-icons/ri/lock-password-line";
import More from "@iconify-icons/ep/more-filled";
import Delete from "@iconify-icons/ep/delete";
import EditPen from "@iconify-icons/ep/edit-pen";
import Refresh from "@iconify-icons/ep/refresh";
import AddFill from "@iconify-icons/ri/add-circle-line";
import DownArrow from "@iconify-icons/ri/arrow-down-s-line";

defineOptions({
  name: "Resource Management"
});

const tableRef = ref();

const {
  maxDay,
  namespaceList,
  deploymentList,
  onNamespaceChange,
  onDeploymentChange,
  onKeywordChange,
  selectedNum,
  onSelectionCancel,
  handleSelectionChange,
  queryForm,
  loading,
  columns,
  dataList,
  deviceDetection,
  onSearch,
  resetForm,
  openDialog,
  onChangeCapacity,
  onReboot
} = useResource(tableRef);
</script>

<template>
  <div :class="['flex', 'justify-between', deviceDetection() && 'flex-wrap']">
    <div :class="[deviceDetection() ? ['w-full', 'mt-2'] : 'w-full']">
      <el-form
        :inline="true"
        :model="queryForm"
        class="search-form bg-bg_color w-[99/100] pl-8 pt-[12px] overflow-auto"
      >
        <el-form-item
          :label="transformI18n('resource.column.namespace')"
          prop="namespace"
        >
          <el-select
            v-model="queryForm.namespace"
            :placeholder="transformI18n('resource.placeholder')"
            class="!w-[180px]"
            filterable
            clearable
            @change="onNamespaceChange"
            @clear="onNamespaceChange('')"
          >
            <el-option
              v-for="item in namespaceList"
              :key="item[0]"
              :label="item[0]"
              :value="item[0]"
            />
          </el-select>
        </el-form-item>
        <el-form-item
          :label="transformI18n('resource.column.deployment')"
          prop="deployment"
        >
          <el-select
            v-model="queryForm.deployment"
            :placeholder="transformI18n('resource.placeholder')"
            filterable
            clearable
            class="!w-[180px]"
            @change="onDeploymentChange"
          >
            <el-option
              v-for="item in deploymentList"
              :key="item[0]"
              :label="item[0]"
              :value="item[0]"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="transformI18n('resource.keyword')" prop="keyword">
          <el-input
            v-model="queryForm.keyword"
            class="!w-[210px]"
            :placeholder="transformI18n('resource.keywordPlaceholder')"
            @input="onKeywordChange"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :icon="useRenderIcon('ri:search-line')"
            @click="onSearch"
          >
            {{ transformI18n("resource.search") }}
          </el-button>
          <el-button :icon="useRenderIcon(Refresh)" @click="resetForm()">
            {{ transformI18n("resource.reset") }}
          </el-button>
        </el-form-item>
      </el-form>

      <PureTableBar title="" :columns="columns" @refresh="onSearch">
        <template #title>
          <div class="ml-[-17px]">
            <el-alert
              :title="`标红字段数据为：最近10天，最大资源使用日【${maxDay}】高峰时段各资源P95数值（值为-1表示未配置）`"
              type="warning"
              show-icon
              :closable="false"
            />
          </div>
        </template>
        <template #buttons>
          <el-button
            type="primary"
            :icon="useRenderIcon(AddFill)"
            @click="openDialog()"
          >
            {{ transformI18n("resource.add") }}
          </el-button>
        </template>
        <template v-slot="{ size, dynamicColumns }">
          <div
            v-if="selectedNum > 0"
            v-motion-fade
            class="bg-[var(--el-fill-color-light)] w-full h-[46px] mb-2 pl-4 flex items-center"
          >
            <div class="flex-auto">
              <span
                style="font-size: var(--el-font-size-base)"
                class="text-[rgba(42,46,54,0.5)] dark:text-[rgba(220,220,242,0.5)]"
              >
                {{ transformI18n("status.pureSelected") }} {{ selectedNum }}
              </span>
              <el-button type="primary" text @click="onSelectionCancel">
                {{ transformI18n("buttons.pureCancelSelected") }}
              </el-button>
            </div>
            <el-button
              type="primary"
              link
              class="mr-1"
              @click="onChangeCapacity(undefined)"
            >
              {{ transformI18n("resource.operation.scale") }}
            </el-button>
            <!-- <el-button
              type="primary"
              link
              class="mr-1"
              @click="onChangeCapacity(undefined, false)"
            >
              扩缩(蓬莱)
            </el-button> -->
            <el-button
              type="danger"
              link
              class="mr-1"
              @click="onReboot(undefined)"
            >
              {{ transformI18n("resource.operation.reboot") }}
            </el-button>
            <!-- <el-button
              type="danger"
              link
              class="mr-1"
              @click="onReboot(undefined, false)"
            >
              重启(蓬莱)
            </el-button> -->
          </div>
          <pure-table
            ref="tableRef"
            row-key="index"
            :border="true"
            adaptive
            :adaptiveConfig="{ offsetBottom: 28 }"
            align-whole="center"
            table-layout="auto"
            :loading="loading"
            :size="size"
            :data="dataList"
            :columns="dynamicColumns"
            :header-cell-style="{
              background: 'var(--el-fill-color-light)',
              color: 'var(--el-text-color-primary)'
            }"
            :default-sort="{ prop: 'p95_pod_cpu_pct', order: 'descending' }"
            @selection-change="handleSelectionChange"
          >
            <template #operation="{ row }">
              <el-button
                class="reset-margin"
                plain
                type="primary"
                :size="size"
                @click="openDialog('修改', row)"
              >
                {{ transformI18n("resource.edit") }}
              </el-button>
              <!-- <el-divider direction="vertical" />
              <el-button
                class="reset-margin"
                link
                type="danger"
                :size="size"
                @click="onChangeCapacity(row)"
              >
                {{ transformI18n("resource.operation.scale") }}
              </el-button>
              <el-divider direction="vertical" />
              <el-button
                class="reset-margin"
                link
                type="danger"
                :size="size"
                @click="onReboot(row)"
              >
                {{ transformI18n("resource.operation.reboot") }}
              </el-button> -->
            </template>
          </pure-table>
        </template>
      </PureTableBar>
    </div>
  </div>
</template>

<style scoped lang="scss">
:deep(.el-dropdown-menu__item i) {
  margin: 0;
}

:deep(.el-button:focus-visible) {
  outline: none;
}

.main-content {
  margin: 20px !important;
}

.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}
</style>
