<template>
  <!-- 编辑弹出框 -->
  <el-dialog
    :title="edit_dig_title"
    :close-on-click-modal="false"
    :visible.sync="edit_dig_visible"
    @close="handleEditDigClose"
  >
    <el-form
      ref="smtp_test_edit_form"
      :model="form"
      :rules="rules"
      label-width="120px"
      size="mini"
    >
      <el-form-item label="目标邮箱地址" prop="target">
        <el-input v-model="form.target" placeholder="请输入目标邮箱地址"></el-input>
      </el-form-item>
      <el-form-item>向目标邮箱发一封邮件，以测试SMTP配置是否正确</el-form-item>
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
  name: "AlarmTestEdit",
  data() {
    return {
      edit_dig_visible: false,
      edit_dig_title: "",
      form: {},
      rules: this.validRules()
    };
  },
  methods: {
    initPage() {
      this.edit_dig_title = "SMTP配置测试";
      this.form = {
        target: ''
      };
      this.edit_dig_visible = true;
    },
    validRules() {
      return {
        target: [{ required: true, message: "请输入目标邮箱", trigger: "blur" },
        { type: 'email', message: '请输入正确的邮箱地址', trigger: ['blur', 'change'] }
        ,]
      };
    },
    // 保存编辑
    saveEdit() {
      this.$refs.smtp_test_edit_form.validate(valid => {
        if (valid) {
          alarmApi.testAlarmConfig(this.form).then(res => {
            this.edit_dig_visible = false;
            this.$emit("refreshList");
            this.$message.success(`邮件发送成功`);
          });
        } else {
          console.log("error submit!!");
          return;
        }
      });
    },
    handleEditDigClose() {
      this.form = {};
      this.$refs.smtp_test_edit_form.resetFields();
    }
  }
};
</script>