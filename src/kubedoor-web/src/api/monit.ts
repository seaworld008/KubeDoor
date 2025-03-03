import { http } from "@/utils/http";

type ResultTable = {
  success: boolean;
  data?: Array<any>;
  meta?: Array<any>;
  count: any;
};

/**
 * 获取K8S环境列表
 */
export const getPromEnv = () => {
  return http.request<ResultTable>("get", "/api/prom_env");
};

/**
 * 获取命名空间列表
 * @param env K8S环境
 */
export const getPromNamespace = (env: string) => {
  return http.request<ResultTable>("get", "/api/prom_ns", {
    params: { env }
  });
};

/**
 * 获取监控数据
 * @param env K8S环境
 * @param ns 命名空间（可选）
 */
export const getPromQueryData = (env: string, ns?: string) => {
  const params: Record<string, string> = { env };
  if (ns) {
    params.ns = ns;
  }
  return http.request<ResultTable>("get", "/api/prom_query", {
    params
  });
};

export const updatePodCount = (data?: any) => {
  return http.request<any>("post", "/api/sql", {
    params: {
      add_http_cors_header: 1,
      default_format: "JSONCompact"
    },
    data: `ALTER TABLE __KUBEDOORDB__.k8s_res_control UPDATE pod_count_manual=${data.pod_count_manual} WHERE env = '${data.env}' AND namespace='${data.namespace}' AND deployment='${data.deployment_name}' `,
    headers: {
      "Content-Type": "text/plain;charset=UTF-8"
    }
  });
};
