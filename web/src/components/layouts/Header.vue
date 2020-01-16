<template>
  <div class="header">
    <div class="logo">gojob分布式任务调度</div>
    <!-- 折叠按钮 -->
    <div class="collapse-btn" @click="collapseChage">
      <span v-if="collapsed">
        <i class="el-icon-s-unfold"></i>
      </span>
      <span v-else>
        <i class="el-icon-s-fold"></i>
      </span>
    </div>
    <div class="header-right">
      <div class="header-user-con">
        <!-- 用户名下拉菜单 -->
        <el-dropdown class="user-name" trigger="click" @command="handleCommand">
          <span class="el-dropdown-link">
            {{username}}
            <i class="el-icon-caret-bottom"></i>
          </span>
          <el-dropdown-menu slot="dropdown">
            <el-dropdown-item divided command="updatePassword">安全设置</el-dropdown-item>
            <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
          </el-dropdown-menu>
        </el-dropdown>
      </div>
    </div>
    <user-password-edit ref="user_password_edit"></user-password-edit>
  </div>
</template>
<script>
import userPasswordEdit from "@/views/user/UserPasswordEdit";
import userApi from "@/api/UserApi";
import { clearAccredit } from "@/utils/accredit";

export default {
  data() {
    return {
      fullscreen: false,
      collapsed: false
    };
  },
  components: {
    userPasswordEdit
  },
  computed: {
    username() {
      return this.$store.getters.getUserName;
    }
  },
  methods: {
    // 用户名下拉菜单选择事件
    handleCommand(command) {
      if (command == "logout") {
        userApi.logout().then(() => {
          clearAccredit();
          this.$router.push("/login");
        });
      }
      if (command == "updatePassword") {
        this.$refs.user_password_edit.initPage();
      }
    },
    // 侧边栏折叠
    collapseChage() {
      this.$store.dispatch("toggleScollapse");
      if (this.$store.getters.getScollapse) {
        this.collapsed = true;
      } else {
        this.collapsed = false;
      }
    }
  },
  mounted() {
    if (document.body.clientWidth < 1500) {
      this.$store.dispatch("closeScollapse");
    }
  }
};
</script>
<style scoped>
.header {
  position: relative;
  box-sizing: border-box;
  width: 100%;
  height: 70px;
  font-size: 22px;
  color: #fff;
}
.collapse-btn {
  float: left;
  padding: 0 1px;
  cursor: pointer;
  line-height: 70px;
}
.header .logo {
  float: left;
  width: 250px;
  line-height: 70px;
}
.header-right {
  float: right;
  padding-right: 50px;
}
.header-user-con {
  display: flex;
  height: 70px;
  align-items: center;
}
.btn-fullscreen {
  transform: rotate(45deg);
  margin-right: 5px;
  font-size: 24px;
}
.btn-bell,
.btn-fullscreen {
  position: relative;
  width: 30px;
  height: 30px;
  text-align: center;
  border-radius: 15px;
  cursor: pointer;
}
.btn-bell-badge {
  position: absolute;
  right: 0;
  top: -2px;
  width: 8px;
  height: 8px;
  border-radius: 4px;
  background: #f56c6c;
  color: #fff;
}
.btn-bell .el-icon-bell {
  color: #fff;
}
.user-name {
  margin-left: 10px;
}
.user-avator {
  margin-left: 20px;
}
.user-avator img {
  display: block;
  width: 40px;
  height: 40px;
  border-radius: 50%;
}
.el-dropdown-link {
  color: #fff;
  cursor: pointer;
}
.el-dropdown-menu__item {
  text-align: center;
}
</style>