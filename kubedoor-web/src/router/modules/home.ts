const { VITE_HIDE_HOME } = import.meta.env;
const Layout = () => import("@/layout/index.vue");
import { $t } from "@/plugins/i18n";

export default {
  path: "/",
  name: "Resource",
  component: Layout,
  redirect: "/resource",
  meta: {
    icon: "ep:menu",
    title: $t("menus.resourceManagement"),
    rank: 0
  },
  children: [
    // {
    //   path: "/welcome",
    //   name: "Welcome",
    //   component: () => import("@/views/welcome/index.vue"),
    //   meta: {
    //     title: "首页",
    //     showLink: VITE_HIDE_HOME === "true" ? false : true
    //   }
    // },
    {
      path: "/resource",
      name: "resource",
      component: () => import("@/views/resource/index.vue"),
      meta: {
        title: $t("menus.resourceManagement"),
        showLink: VITE_HIDE_HOME === "true" ? false : true
      }
    }
  ]
} satisfies RouteConfigsTable;
