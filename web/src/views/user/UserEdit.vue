<template>
  <!-- 编辑弹出框 -->
  <el-dialog
    :title="edit_dig_title"
    :close-on-click-modal="false"
    :visible.sync="edit_dig_visible"
    @close="handleEditDigClose"
  >
    <el-form ref="user_edit_form" :model="form" :rules="rules" label-width="100px" size="mini">
      <el-form-item v-if="edit_dig_title=='新增用户'" label="用户名" prop="name">
        <el-input v-model="form.name" placeholder="请输入用户名"></el-input>
      </el-form-item>
      <el-form-item v-if="edit_dig_title=='新增用户'" label="密码" prop="password">
        <el-input type="password" v-model="form.password" placeholder="请输入密码"></el-input>
      </el-form-item>
      <el-form-item label="邮箱" prop="email">
        <el-input v-model="form.email" placeholder="请输入邮箱"></el-input>
      </el-form-item>
    </el-form>
    <span slot="footer" class="dialog-footer">
      <el-button @click="edit_dig_visible = false">取 消</el-button>
      <el-button type="primary" @click="saveEdit">提 交</el-button>
    </span>
  </el-dialog>
</template>

<script>
import userApi from "@/api/UserApi";

export default {
  name: "UserEdit",
  data() {
    return {
      edit_dig_visible: false,
      edit_dig_title: "",
      form: this.emptyForm(),
      rules: this.validRules()
    };
  },
  methods: {
    initPage(id) {
      if (id) {
        this.edit_dig_title = "编辑用户";
        userApi.getUser(id).then(res => {
          this.form = res.data;
          this.edit_dig_visible = true;
        });
      } else {
        this.edit_dig_title = "新增用户";
        this.edit_dig_visible = true;
      }
    },
    emptyForm() {
      return {
        name: "",
        password: "",
        email: ""
      };
    },
    validRules() {
      return {
        name: [{ required: true, message: "请输入用户名", trigger: "blur" }],
        password: [{ required: true, message: "请输入密码", trigger: "blur" }],
        email: [
          {
            type: "email",
            message: "请输入正确的邮箱地址",
            trigger: ["blur", "change"]
          }
        ]
      };
    },
    // 保存编辑
    saveEdit() {
      this.$refs.user_edit_form.validate(valid => {
        if (valid) {
          if (this.edit_dig_title === "新增用户") {
            userApi.postUser(this.form).then(res => {
              this.edit_dig_visible = false;
              this.$emit("refreshList");
              this.$message.success(`新增成功`);
            });
          } else {
            userApi.putUser({
              name:this.form.name,
              email:this.form.email
            }).then(res => {
              this.edit_dig_visible = false;
              this.$emit("refreshList");
              this.$message.success(`修改成功`);
            });
          }
        } else {
          console.log("error submit!!");
          return;
        }
      });
    },
    handleEditDigClose() {
      this.edit_dig_title = "";
      this.form = this.emptyForm();
      this.$refs.user_edit_form.resetFields();
    }
  }
};
</script>