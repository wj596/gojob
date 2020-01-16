<template>
  <div class="sidebar">
    <el-menu
      class="sidebar-el-menu"
      :default-active="onRoutes"
      :collapse="isCollapse"
      background-color="#324157"
      text-color="#bfcbd9"
      active-text-color="#20a0ff"
      unique-opened
      router
    >
      <template v-for="item in items">
        <template v-if="item.children">
          <el-submenu :index="item.name" :key="item.name">
            <template slot="title">
              <i :class="item.icon"></i>
              <span slot="title">{{ item.title }}</span>
            </template>
            <el-menu-item
              v-for="(subItem,i) in item.children"
              :key="i"
              :index="subItem.name"
            >{{ subItem.title }}</el-menu-item>
          </el-submenu>
        </template>
        <template v-else>
          <el-menu-item :index="item.name" :key="item.name">
            <i :class="item.icon"></i>
            <span slot="title">{{ item.title }}</span>
          </el-menu-item>
        </template>
      </template>
    </el-menu>
  </div>
</template>

<script>
import runtimeApi from "@/api/RuntimeApi";
import { res_standlone } from "@/data/resources.js";
import { res_cluster } from "@/data/resources.js";
export default {
  data() {
    return {
      items: []
    };
  },
  mounted() {
    runtimeApi.getRunmode().then(res => {
      if ("standalone" === res.data) {
        this.items = res_standlone;
      } else {
        this.items = res_cluster;
      }
    });
  },
  computed: {
    onRoutes() {
      return this.$route.path.replace("/", "");
    },
    isCollapse() {
      return this.$store.getters.getScollapse;
    }
  }
};
</script>

<style scoped>
.sidebar {
  display: block;
  position: absolute;
  left: 0;
  top: 70px;
  bottom: 0;
  overflow: auto;
}
.sidebar::-webkit-scrollbar {
  width: 0;
}
.sidebar-el-menu:not(.el-menu--collapse) {
  width: 250px;
}
.sidebar > ul {
  height: 100%;
}
</style>
