<script setup lang="ts">
import { ref } from "vue";
import ReCol from "@/components/ReCol";
import { formRules } from "../utils/rule";
import { FormProps } from "../utils/types";
import addLine from "@iconify-icons/ri/add-line";
import { transformI18n } from "@/plugins/i18n";

const props = withDefaults(defineProps<FormProps>(), {
  formInline: () => ({
    namespace: "",
    deployment: "",
    pod_count_manual: "",
    limit_cpu_m: "",
    limit_mem_mb: "",
    request_cpu_m: "",
    request_mem_mb: "",
    pod_count: ""
  }),
  namespace: () => [],
  isEdit: false
});

const formRef = ref();
const newFormInline = ref(props.formInline);

function getRef() {
  return formRef.value;
}

defineExpose({ getRef });
</script>

<template>
  <el-form
    ref="formRef"
    :model="newFormInline"
    :rules="formRules"
    label-width="145px"
  >
    <el-row :gutter="30">
      <re-col :value="20" :xs="24" :sm="24">
        <el-form-item
          :label="transformI18n('resource.column.namespace')"
          prop="namespace"
        >
          <!-- <el-input v-model="newFormInline.namespace" :disabled="isEdit" /> -->
          <el-select
            v-model="newFormInline.namespace"
            :placeholder="transformI18n('resource.placeholder')"
            allow-create
            filterable
            :disabled="isEdit"
          >
            <el-option
              v-for="item in props.namespace"
              :key="item[0]"
              :label="item[0]"
              :value="item[0]"
            />
          </el-select>
        </el-form-item>
      </re-col>
      <re-col :value="20" :xs="24" :sm="24">
        <el-form-item
          :label="transformI18n('resource.column.deployment')"
          prop="deployment"
        >
          <el-input v-model="newFormInline.deployment" :disabled="isEdit" />
        </el-form-item>
      </re-col>
      <re-col v-if="isEdit" :value="20" :xs="24" :sm="24">
        <el-form-item
          :label="transformI18n('resource.column.podCount')"
          prop="pod_count"
        >
          <el-input
            v-model="newFormInline.pod_count"
            type="number"
            :disabled="isEdit"
          />
        </el-form-item>
      </re-col>
      <re-col :value="20" :xs="24" :sm="24">
        <el-form-item
          :label="transformI18n('resource.column.podCountManual')"
          prop="pod_count_manual"
        >
          <el-input v-model="newFormInline.pod_count_manual" type="number" />
        </el-form-item>
      </re-col>
      <re-col :value="20" :xs="24" :sm="24">
        <el-form-item
          :label="transformI18n('resource.column.requestCpuM')"
          prop="request_cpu_m"
        >
          <el-input
            v-model="newFormInline.request_cpu_m"
            type="number"
            :disabled="isEdit"
          />
        </el-form-item>
      </re-col>
      <re-col :value="20" :xs="24" :sm="24">
        <el-form-item
          :label="transformI18n('resource.column.limitCpuM')"
          prop="limit_cpu_m"
        >
          <el-input v-model="newFormInline.limit_cpu_m" type="number" />
        </el-form-item>
      </re-col>
      <re-col :value="20" :xs="24" :sm="24">
        <el-form-item
          :label="transformI18n('resource.column.requestMemMb')"
          prop="request_mem_mb"
        >
          <el-input
            v-model="newFormInline.request_mem_mb"
            type="number"
            :disabled="isEdit"
          />
        </el-form-item>
      </re-col>

      <re-col :value="20" :xs="24" :sm="24">
        <el-form-item
          :label="transformI18n('resource.column.limitMemMb')"
          prop="limit_mem_mb"
        >
          <el-input v-model="newFormInline.limit_mem_mb" type="number" />
        </el-form-item>
      </re-col>
    </el-row>
  </el-form>
</template>

<style scoped>
.avatar-uploader .el-upload {
  position: relative;
  overflow: hidden;
  cursor: pointer;
}

.avatar-uploader-icon {
  width: 145px;
  height: 145px;
  font-size: 40px;
  line-height: 145px;
  color: #8c939d;
  text-align: center;
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
}

.avatar-uploader-icon:hover {
  color: #409eff;
  border-color: #409eff;
}

.avatar {
  display: block;
  width: 145px;
  height: 145px;
}
</style>
