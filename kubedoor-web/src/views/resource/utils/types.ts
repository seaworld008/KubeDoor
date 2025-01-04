interface FormItemProps {
  namespace: string;
  deployment: string;
  pod_count_manual: string;
  limit_cpu_m: string;
  limit_mem_mb: string;
  [string: string]: any;
}
interface FormProps {
  formInline: FormItemProps;
  isEdit: boolean;
  namespace: [];
}

export type { FormItemProps, FormProps };
