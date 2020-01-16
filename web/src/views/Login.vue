<template>
  <div class="login-container">
    <el-form class="login-form" :model="loginForm" label-position="left">
      <h3 class="title">gojob分布式任务调度系统</h3>
      <el-form-item>
        <el-input v-model="loginForm.username" @focus="error=''">
          			<template slot="prepend">帐号</template>
        </el-input>
      </el-form-item>
      <el-form-item>
        <el-input type="password" v-model="loginForm.password" @focus="error=''">
          <template slot="prepend">密码</template>
        </el-input>
      </el-form-item>
      <el-form-item>
        <el-button type="primary"  :loading="loading" @click="handleLogin" style="width:100%;">
          登陆
        </el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script>
import userApi from "@/api/UserApi";
import { Message } from "element-ui";
import {storeAccredit} from '@/utils/accredit'

export default {
  name: "Login",
  data() {
    return {
      loginForm: {
        username: "admin",
        password: "123456"
      },
      loading: false
    };
  },
  methods: {
    handleLogin() {
      if (this.loginForm.username === "") {
        Message({
          message: "用户名不能为空",
          type: "error",
          duration: 5 * 1000
        });
        return false;
      }
      if (this.loginForm.password === "") {
        Message({
          message: "密码不能为空",
          type: "error",
          duration: 5 * 1000
        });
        return false;
      }
      this.loading = true;

      userApi.login({
          name: this.loginForm.username,
          password: this.loginForm.password
      }).then(accredit => {
          storeAccredit(accredit)
          this.loading = false
          this.$router.push({ path: "/" })
      }).catch(error => {
          this.loading = false;
      });
    }
  }
};
</script>
<style scoped>
.login-container {
  position: fixed;
  height: 100%;
  width: 100%;
  background-color: #2d3a4b;
}
.login-form {
  position: absolute;
  left: 0;
  right: 0;
  width: 520px;
  padding: 35px 35px 15px 35px;
  margin: 120px auto;
}
.title {
  font-size: 26px;
  font-weight: 400;
  color: #eee;
  margin: 0px auto 40px auto;
  text-align: center;
  font-weight: bold;
}
.show-pwd {
  position: absolute;
  right: 10px;
  top: 7px;
  font-size: 16px;
  color: #889aa4;
  cursor: pointer;
  user-select: none;
}
.tips {
  font-size: 14px;
  color: #ff0000;
  margin-bottom: 10px;
}
</style>

