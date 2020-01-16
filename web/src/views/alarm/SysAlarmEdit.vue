<template>
  <!-- 编辑弹出框 -->
  <div>
    <el-dialog
      :title="edit_dig_title"
      :close-on-click-modal="false"
      :visible.sync="edit_dig_visible"
      @close="handleEditDigClose"
    >
      <el-form
        ref="cluster_alarm_edit_form"
        :model="form"
        :rules="rules"
        label-width="120px"
        size="mini"
      >
        <el-row>
          <el-col :span="20">
            <el-form-item label="告警邮箱" prop="sysAlarmEmail">
              <el-input
                v-model="form.sysAlarmEmail"
                placeholder="请输入告警邮件地址，多个用|分隔;点击后面的按钮，可选择系统用户的邮箱"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :span="4">
            &nbsp;
            <el-button
              @click="handleUserSelectionDig"
              type="primary"
              icon="el-icon-user-solid"
              circle
              size="mini"
            ></el-button>
          </el-col>
        </el-row>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button @click="edit_dig_visible = false">取 消</el-button>
        <el-button type="primary" @click="saveEdit">提 交</el-button>
      </span>
    </el-dialog>

    <el-dialog title="系统用户列表" width="40%" :visible.sync="user_selection_dig_visible">
      <el-table
        :height="300"
        :show-header="false"
        :data="user_selection_data"
        :row-style="{height:'30px'}"
        :header-row-style="{height:'30px'}"
        :cell-style="{padding:'1px'}"
        @selection-change="handleUserSelectionChange"
      >
        <el-table-column type="selection" width="55"></el-table-column>
        <el-table-column>
          <template slot-scope="scope">{{scope.row.name}} ({{scope.row.email}})</template>
        </el-table-column>
      </el-table>
      <span slot="footer" class="dialog-footer">
        <el-button @click="user_selection_dig_visible = false" icon="el-icon-close">取 消</el-button>
        <el-button type="primary" @click="handleUserSelection" icon="el-icon-check">选 中</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import userApi from "@/api/UserApi";
import alarmApi from "@/api/AlarmApi";

export default {
  name: "SysAlarmEdit",
  data() {
    return {
      edit_dig_visible: false,
      edit_dig_title: "",
      form: {},
      rules: this.validRules(),
      user_selection_dig_visible: false,
      user_selection_data: [],
      user_selection_result: []
    };
  },
  methods: {
    initPage() {
      alarmApi.getAlarmConfig().then(res => {
        if (res.data) {
          this.form = res.data;
          this.edit_dig_title = "编辑系统故障告警地址";
          this.edit_dig_visible = true;
        }
      });
    },
    validRules() {
      return {
        sysAlarmEmail: [
          {
            required: true,
            message: "请输入系统故障告警邮件地址",
            trigger: "blur"
          }
        ]
      };
    },
    // 保存编辑
    saveEdit() {
      this.$refs.cluster_alarm_edit_form.validate(valid => {
        if (valid) {
          if (this.form.sysAlarmEmail) {
            let p = this.form.sysAlarmEmail;
            let ps = p.split("|");
            for (let i in ps) {
              var reg = new RegExp(/^\S+@\S+\.\S{2,}$/);
              if (!reg.test(ps[i])) {
                this.$message.error("告警邮件中存在不合规的邮件地址，请更正");
                return;
              }
            }
          }
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
    handleUserSelectionDig() {
      userApi.getUsers({ hasEmail: true }).then(res => {
        this.user_selection_data = res.data;
        this.user_selection_dig_visible = true;
      });
    },
    handleUserSelectionChange(result) {
      this.user_selection_result = result;
    },
    handleUserSelection(result) {
      for (let i in this.user_selection_result) {
        let result = this.user_selection_result[i];
        if (this.form.sysAlarmEmail == "") {
          this.form.sysAlarmEmail += result.email;
        } else {
          if (!this.separatorInclude(this.form.sysAlarmEmail, result.email)) {
            this.form.sysAlarmEmail += "|" + result.email;
          }
        }
      }
      this.user_selection_dig_visible = false;
    },
    separatorInclude(str, substring) {
      let strs = str.split("|");
      for (let i in strs) {
        let strss = strs[i];
        if (strss == substring) {
          return true;
        }
      }
      return false;
    },
    handleEditDigClose() {
      this.form = {};
      this.$refs.cluster_alarm_edit_form.resetFields();
    }
  }
};
</script>