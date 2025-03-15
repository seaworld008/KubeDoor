<template>
  <div class="realtime-monitor-container">
    <!-- 搜索表单 -->
    <div class="search-section">
      <el-form :model="searchForm" inline style="margin-bottom: -18px">
        <el-form-item label="K8S">
          <el-select
            v-model="searchForm.env"
            placeholder="请选择K8S环境"
            class="!w-[180px]"
            filterable
            @change="handleEnvChange"
          >
            <el-option
              v-for="item in envOptions"
              :key="item"
              :label="item"
              :value="item"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="命名空间">
          <el-select
            v-model="searchForm.ns"
            placeholder="请选择命名空间"
            class="!w-[180px]"
            filterable
            clearable
            @change="handleSearch"
          >
            <el-option
              v-for="item in nsOptions"
              :key="item"
              :label="item"
              :value="item"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="关键字">
          <el-input
            v-model="searchForm.keyword"
            placeholder="请输入关键字"
            clearable
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :disabled="!searchForm.env"
            @click="handleSearch"
          >
            搜索
          </el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 表格数据 -->
    <!-- 更新弹窗 -->
    <el-dialog v-model="updateDialogVisible" title="更新镜像" width="500px">
      <el-form :model="updateForm" label-width="80px">
        <el-form-item label="镜像标签">
          <el-input
            v-model="updateForm.imageTag"
            placeholder="请输入镜像标签"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="updateDialogVisible = false">取消</el-button>
          <el-button
            type="primary"
            :loading="updateLoading"
            @click="handleUpdate"
          >
            确定
          </el-button>
        </span>
      </template>
    </el-dialog>

    <div class="mt-2">
      <el-card v-loading="loading">
        <el-table
          :data="filteredTableData"
          style="width: 100%"
          stripe
          :default-sort="{ prop: 'podCount', order: 'descending' }"
        >
          <el-table-column
            prop="env"
            label="K8S"
            min-width="100"
            show-overflow-tooltip
          />
          <el-table-column
            prop="namespace"
            label="命名空间"
            min-width="80"
            show-overflow-tooltip
          />
          <el-table-column
            prop="deployment"
            label="微服务"
            min-width="140"
            show-overflow-tooltip
          />
          <el-table-column
            prop="podCount"
            label="POD"
            min-width="80"
            align="center"
            sortable
          >
            <template #header>
              <span style="color: #409eff">POD</span>
            </template>
            <template #default="scope">
              <span style="color: #409eff; font-weight: bold">{{
                scope.row.podCount
              }}</span>
            </template>
          </el-table-column>
          <el-table-column
            prop="avgCpu"
            label="平均CPU"
            min-width="100"
            align="center"
            sortable
          >
            <template #default="scope"> {{ scope.row.avgCpu }}m </template>
          </el-table-column>
          <el-table-column
            prop="maxCpu"
            label="最大CPU"
            min-width="100"
            align="center"
            sortable
          >
            <template #header>
              <span style="color: #f56c6c">最大CPU</span>
            </template>
            <template #default="scope">
              <span style="color: #f56c6c; font-weight: bold"
                >{{ scope.row.maxCpu }}m</span
              >
            </template>
          </el-table-column>
          <el-table-column
            prop="requestCpu"
            label="需求CPU"
            min-width="100"
            align="center"
            sortable
          >
            <template #default="scope"> {{ scope.row.requestCpu }}m </template>
          </el-table-column>
          <el-table-column
            prop="limitCpu"
            label="限制CPU"
            min-width="100"
            align="center"
            sortable
          >
            <template #header>
              <span style="color: #409eff">限制CPU</span>
            </template>
            <template #default="scope">
              <span style="color: #409eff; font-weight: bold"
                >{{ scope.row.limitCpu }}m</span
              >
            </template>
          </el-table-column>
          <el-table-column
            prop="avgMem"
            label="平均MEM"
            min-width="120"
            align="center"
            sortable
          >
            <template #default="scope"> {{ scope.row.avgMem }}MB </template>
          </el-table-column>

          <el-table-column
            prop="maxMem"
            label="最大MEM"
            min-width="120"
            align="center"
            sortable
          >
            <template #header>
              <span style="color: #f56c6c">最大MEM</span>
            </template>
            <template #default="scope">
              <span style="color: #f56c6c; font-weight: bold"
                >{{ scope.row.maxMem }}MB</span
              >
            </template>
          </el-table-column>

          <el-table-column
            prop="requestMem"
            label="需求MEM"
            min-width="120"
            align="center"
            sortable
          >
            <template #default="scope"> {{ scope.row.requestMem }}MB </template>
          </el-table-column>

          <el-table-column
            prop="limitMem"
            label="限制MEM"
            min-width="120"
            align="center"
            sortable
          >
            <template #header>
              <span style="color: #409eff">限制MEM</span>
            </template>
            <template #default="scope">
              <span style="color: #409eff; font-weight: bold">
                {{ scope.row.limitMem }}MB
              </span>
            </template>
          </el-table-column>
          <el-table-column label="操作" min-width="80" align="center">
            <template #default="scope">
              <el-dropdown>
                <el-button type="primary" size="small" plain>
                  操作
                  <el-icon class="el-icon--right"><arrow-down /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click="handleCapacity(scope.row)"
                      >扩缩容</el-dropdown-item
                    >
                    <el-dropdown-item @click="onReboot(scope.row)"
                      >重启</el-dropdown-item
                    >
                    <el-dropdown-item @click="openUpdateDialog(scope.row)"
                      >更新</el-dropdown-item
                    >
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from "vue";
import { ElMessage } from "element-plus";
import { ArrowDown } from "@element-plus/icons-vue";
import { getPromEnv, getPromNamespace, getPromQueryData } from "@/api/monit";
import { useResource } from "./utils/hook";

const { onChangeCapacity, onReboot, onUpdateImage } = useResource();

// 定义表单数据
const searchForm = reactive({
  env: "",
  ns: "",
  keyword: ""
});

// 定义选项数据
const envOptions = ref<string[]>([]);
const nsOptions = ref<string[]>([]);

// 表格数据
const tableData = ref<any[]>([]);
const filteredTableData = computed(() => {
  if (!searchForm.keyword) {
    return tableData.value;
  }

  const keyword = searchForm.keyword.toLowerCase();
  return tableData.value.filter(
    item =>
      item.env.toLowerCase().includes(keyword) ||
      (item.namespace && item.namespace.toLowerCase().includes(keyword)) ||
      (item.deployment && item.deployment.toLowerCase().includes(keyword))
  );
});

// 加载状态
const loading = ref(false);

// 更新弹窗相关
const updateDialogVisible = ref(false);
const updateLoading = ref(false);
const selectedDeployment = ref<any>(null);
const updateForm = reactive({
  imageTag: ""
});

const openUpdateDialog = (row: any) => {
  selectedDeployment.value = row;
  updateForm.imageTag = "";
  updateDialogVisible.value = true;
};

const handleUpdate = async () => {
  if (!updateForm.imageTag) {
    ElMessage.warning("请输入镜像标签");
    return;
  }

  updateLoading.value = true;
  try {
    await onUpdateImage(selectedDeployment.value.env, {
      deployment: selectedDeployment.value.deployment,
      namespace: selectedDeployment.value.namespace,
      image_tag: updateForm.imageTag
    });
    updateDialogVisible.value = false;
    handleSearch();
  } finally {
    updateLoading.value = false;
  }
};

// 获取K8S环境列表
const getEnvOptions = async (): Promise<void> => {
  try {
    const res = await getPromEnv();
    if (res.data && res.data.length > 0) {
      envOptions.value = res.data.map(item => item);
      searchForm.env = res.data[0];
      await getNsOptions(res.data[0]);
    }
    return Promise.resolve();
  } catch (error) {
    console.error("获取K8S环境列表失败:", error);
    ElMessage.error("获取K8S环境列表失败");
    return Promise.reject(error);
  }
};

// 处理环境变化
const handleEnvChange = async (val: string) => {
  searchForm.ns = "";
  if (val) {
    await getNsOptions(val);
    handleSearch();
  } else {
    tableData.value = [];
  }
};

// 获取命名空间列表
const getNsOptions = async (env: string): Promise<void> => {
  if (!env) {
    nsOptions.value = [];
    return Promise.resolve();
  }

  try {
    const res = await getPromNamespace(env);
    if (res.data) {
      nsOptions.value = res.data.map(item => item);
      searchForm.ns = res.data[0];
    }
    return Promise.resolve();
  } catch (error) {
    console.error("获取命名空间列表失败:", error);
    ElMessage.error("获取命名空间列表失败");
    return Promise.reject(error);
  }
};

// 处理搜索
const handleSearch = async () => {
  if (!searchForm.env) {
    ElMessage.warning("请选择K8S环境");
    return;
  }

  loading.value = true;
  try {
    const res = await getPromQueryData(searchForm.env, searchForm.ns);
    if (res.data) {
      tableData.value = res.data.map(item => ({
        env: item[0] || "-",
        namespace: item[1] || "-",
        deployment: item[2] || "-",
        podCount: item[3] || 0,
        avgCpu: item[4] ? Math.round(item[4]) : 0,
        maxCpu: item[5] ? Math.round(item[5]) : 0,
        requestCpu: item[6] ? Math.round(item[6]) : 0,
        limitCpu: item[7] ? Math.round(item[7]) : 0,
        avgMem: item[8] ? Math.round(item[8]) : 0,
        maxMem: item[9] ? Math.round(item[9]) : 0,
        requestMem: item[10] ? Math.round(item[10]) : 0,
        limitMem: item[11] ? Math.round(item[11]) : 0
      }));
    } else {
      tableData.value = [];
    }
  } catch (error) {
    console.error("获取监控数据失败:", error);
    ElMessage.error("获取监控数据失败");
    tableData.value = [];
  } finally {
    loading.value = false;
  }
};

// 处理重置
const handleReset = async () => {
  searchForm.env = envOptions.value[0] || "";
  searchForm.keyword = "";
  await handleEnvChange(envOptions.value[0]);
};

const handleCapacity = async (row: any) => {
  await onChangeCapacity(row);
  // handleSearch();
};

// 页面加载时获取环境列表
onMounted(async () => {
  await getEnvOptions();
  handleSearch();
});
</script>

<style scoped>
.search-section {
  background-color: #fff;
  padding: 16px;
  border-radius: 8px;
}

.el-form-item {
  margin-bottom: 18px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
