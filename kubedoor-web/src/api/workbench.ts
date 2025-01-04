import { http } from "@/utils/http";

type Result = {
  success: boolean;
  data: Array<any>;
};

export const initByDays = (days: number) => {
  return http.request<Result>("post", "/api/table", { params: { days } });
};

export const whSwitch = (action: string) => {
  return http.request<Result>("get", "/api/webhook_switch", {
    params: { action }
  });
};
