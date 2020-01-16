<template>
  <!-- 编辑弹出框 -->
  <el-dialog
    :title="edit_dig_title"
    :close-on-click-modal="false"
    :visible.sync="edit_dig_visible"
    @close="handleEditDigClose"
  >
    <el-form
      ref="user_password_edit_form"
      :model="form"
      :rules="rules"
      label-width="100px"
      size="mini"
    >
      <el-form-item label="新密码" prop="password">
        <el-input type="password" v-model="form.password" placeholder="请输入新密码"></el-input>
      </el-form-item>
      <el-form-item label="确认新密码" prop="affirmPassword">
        <el-input type="password" v-model="form.affirmPassword" placeholder="请确认新密码"></el-input>
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
  name: "UserPasswordEdit",
  data() {
    return {
      edit_dig_visible: false,
      edit_dig_title: "",
      form: {
        name: "",
        password: "",
        affirmPassword: ""
      },
      rules: this.validRules()
    };
  },
  methods: {
    initPage() {
      this.edit_dig_title = "修改密码";
      this.edit_dig_visible = true;
    },
    validRules() {
      return {
        password: [
          { required: true, message: "请输入新密码", trigger: "blur" }
        ],
        affirmPassword: [
          { required: true, message: "请确认新密码", trigger: "blur" }
        ]
      };
    },
    // 保存编辑
    saveEdit() {
      this.$refs.user_password_edit_form.validate(valid => {
        if (valid) {
          if (this.form.password != this.form.affirmPassword) {
            this.$message.error("两次密码输入不一致，请更正");
            return;
          }
          this.form.name = this.$store.getters.getUserName;
          userApi.putUser(this.form).then(res => {
            this.edit_dig_visible = false;
            this.$message.success(`修改成功`);
          });
        } else {
          console.log("error submit!!");
          return;
        }
      });
    },
    handleEditDigClose() {
      this.form = {
        name: "",
        password: "",
        affirmPassword: ""
      };
      this.$refs.user_password_edit_form.resetFields();
    }
  }
};
</script>