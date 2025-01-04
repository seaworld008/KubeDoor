// import "./reset.css";
// import dayjs from "dayjs";
import { deviceDetection } from "@pureadmin/utils";
import { getCollection } from "@/api/resource";
// import { ElMessageBox } from "element-plus";
import { ref, reactive } from "vue";
import { transformI18n } from "@/plugins/i18n";
import { message } from "@/utils/message";

export function useCollection() {
  const queryForm = reactive({
    date: ""
  });
  const dataList = ref([]);
  const loading = ref(false);

  const columns: TableColumnList = [
    {
      label: transformI18n("resource.column.env"),
      prop: "env"
    },
    // {
    //   label: transformI18n("resource.column.envPct"),
    //   prop: "env_pct",
    //   formatter: ({ env_pct }) => env_pct.toFixed(2) + "%"
    // },
    {
      label: transformI18n("resource.column.namespace"),
      prop: "namespace"
    },
    {
      label: transformI18n("resource.column.deployment"),
      prop: "deployment"
    },
    {
      label: transformI18n("resource.column.podCount"),
      prop: "pod_count",
      sortable: true
    },
    {
      label: transformI18n("resource.column.p95PodQps"),
      prop: "p95_pod_qps",
      sortable: true,
      hide: true,
      formatter: ({ p95_pod_qps }) => p95_pod_qps.toFixed(2)
    },
    {
      label: transformI18n("resource.column.p95PodCpuPct"),
      prop: "p95_pod_cpu_pct",
      sortable: true,
      formatter: ({ p95_pod_cpu_pct }) => p95_pod_cpu_pct.toFixed(2) + "%"
    },
    {
      label: transformI18n("resource.column.p95PodLoad"),
      prop: "p95_pod_load",
      sortable: true,
      formatter: ({ p95_pod_load }) => p95_pod_load.toFixed(2)
    },
    {
      label: transformI18n("resource.column.requestPodCpuM"),
      prop: "request_pod_cpu_m",
      sortable: true,
      formatter: ({ request_pod_cpu_m }) => request_pod_cpu_m.toFixed(2)
    },
    {
      label: transformI18n("resource.column.limitPodCpuM"),
      prop: "limit_pod_cpu_m",
      sortable: true,
      formatter: ({ limit_pod_cpu_m }) => limit_pod_cpu_m.toFixed(2)
    },
    {
      label: transformI18n("resource.column.p95PodWssPct"),
      prop: "p95_pod_wss_pct",
      sortable: true,
      formatter: ({ p95_pod_wss_pct }) => p95_pod_wss_pct.toFixed(2) + "%"
    },
    {
      label: transformI18n("resource.column.p95PodWssMb"),
      prop: "p95_pod_wss_mb",
      sortable: true
    },
    {
      label: transformI18n("resource.column.requestPodMemMb"),
      prop: "request_pod_mem_mb",
      sortable: true,
      formatter: ({ request_pod_mem_mb }) => request_pod_mem_mb.toFixed(2)
    },
    {
      label: transformI18n("resource.column.limitPodMemMb"),
      prop: "limit_pod_mem_mb",
      sortable: true,
      formatter: ({ limit_pod_mem_mb }) => limit_pod_mem_mb.toFixed(2)
    },
    {
      label: transformI18n("resource.column.p95PodG1gcQps"),
      prop: "p95_pod_g1gc_qps",
      sortable: true,
      hide: true,
      formatter: ({ p95_pod_g1gc_qps }) => p95_pod_g1gc_qps.toFixed(2)
    },
    {
      label: transformI18n("resource.column.podJvmMaxMb"),
      prop: "pod_jvm_max_mb",
      sortable: true,
      formatter: ({ pod_jvm_max_mb }) => pod_jvm_max_mb.toFixed(2),
      hide: true
    },
    {
      label: transformI18n("resource.column.date"),
      prop: "date",
      hide: true
    }
  ];

  async function onSearch() {
    if (!queryForm.date) {
      message(transformI18n("resource.message.dateIsNecessary"), {
        type: "warning"
      });
      return;
    }
    loading.value = true;
    const { data, meta } = await getCollection(queryForm.date);
    dataList.value = data.map(item => {
      let obj = {};
      for (let index = 0; index < meta.length; index++) {
        const key = meta[index].name;
        obj[key] = item[index];
      }
      return obj;
    });
    loading.value = false;
  }

  function resetForm() {
    queryForm.date = "";
    onSearch();
  }

  return {
    queryForm,
    loading,
    columns,
    dataList,
    deviceDetection,
    onSearch,
    resetForm
  };
}
