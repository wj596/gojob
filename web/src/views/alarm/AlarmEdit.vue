<template>
  <!-- 编辑弹出框 -->
  <el-dialog
    :title="edit_dig_title"
    :close-on-click-modal="false"
    :visible.sync="edit_dig_visible"
    @close="handleEditDigClose"
  >
    <el-form ref="setting_edit_form" :model="form" :rules="rules" label-width="120px" size="mini">
      <el-form-item label="地址" prop="smtpHost">
        <el-input v-model="form.smtpHost" placeholder="请输入SMTP地址，注意不是邮件地址"></el-input>
      </el-form-item>
      <el-form-item label="端口" prop="smtpPort">
        <el-input v-model="form.smtpPort" placeholder="请输入SMTP端口"></el-input>
      </el-form-item>
      <el-form-item label="用户名" prop="smtpUser">
        <el-input v-model="form.smtpUser" placeholder="请输入SMTP用户名"></el-input>
      </el-form-item>
      <el-form-item label="密码" prop="smtpPassword">
        <el-input type="password" v-model="form.smtpPassword" placeholder="请输入SMTP密码"></el-input>
      </el-form-item>
    </el-form>
    <span slot="footer" class="dialog-footer">
      <el-button @click="edit_dig_visible = false">取 消</el-button>
      <el-button type="primary" @click="saveEdit">提 交</el-button>
    </span>
  </el-dialog>
</template>

<script>
import alarmApi from "@/api/AlarmApi";

export default {
  name: "AlarmEdit",
  data() {
    return {
      edit_dig_visible: false,
      edit_dig_title: "",
      form: {},
      rules: this.validRules()
    };
  },
  methods: {
    initPage(host, port, user, password) {
      this.edit_dig_title = "编辑SMTP(简单邮件传输协议)属性";
      let newPort = "";
      if (port > 0) {
        newPort = port;
      }
      this.form = {
        smtpHost: host,
        smtpPort: newPort,
        smtpUser: user,
        smtpPassword: password
      };

      this.edit_dig_visible = true;
    },
    validRules() {
      return {
        smtpHost: [
          { required: true, message: "请输入地址", trigger: "blur" }
        ],
        smtpPort: [
          { required: true, message: "请输入端口", trigger: "change" },
          {
            validator(rule, value, callback) {
              if (Number.isInteger(Number(value)) && Number(value) > 0) {
                callback();
              } else {
                callback(new Error("端口必须为数字"));
              }
            },
            trigger: "blur"
          }
        ],
        smtpUser: [
          { required: true, message: "请输入用户名", trigger: "blur" }
        ],
        smtpPassword: [
          { required: true, message: "请输入密码", trigger: "blur" }
        ]
      };
    },
    // 保存编辑
    saveEdit() {
      this.$refs.setting_edit_form.validate(valid => {
        if (valid) {
          this.form.smtpPort = parseInt(this.form.smtpPort);
          alarmApi.putAlarmConfig(this.form).then(res => {
            this.edit_dig_visible = false;
            this.$emit("refreshList");
            this.$message.success(`修改成功`);
          });
        } else {
          console.log("error submit!!");
          return;
        }
      });
    },
    handleEditDigClose() {
      this.form = {};
      this.$refs.setting_edit_form.resetFields();
    }
  }
};
</script>