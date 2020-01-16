<template>
  <div id="app">
    <router-view/>
  </div>
</template>

<script>
export default {
  name: "App",
  created() {
    this.$router.push("/dashboard"); // 默认页面,系统首页
  }
};
</script>
<style>
[v-clock] {
  display: none;
}

@import "assets/css/main.css";
@import "assets/css/color-dark.css"; /*深色主题*/
</style>