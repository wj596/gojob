<template>
  <!-- 编辑弹出框 -->
  <el-dialog
    :title="edit_dig_title"
    :close-on-click-modal="false"
    :visible.sync="edit_dig_visible"
    @close="handleEditDigClose"
  >
    <el-table
      :height="350"
      :show-header="false"
      :data="user_mail_data"
      :row-style="{height:'30px'}"
      :header-row-style="{height:'30px'}"
      :cell-style="{padding:'1px'}"
      @selection-change="handleUserMailSelectionChange"
    >
      <el-table-column type="selection" width="55"></el-table-column>
      <el-table-column>
        <template slot-scope="scope">{{scope.row.name}} ({{scope.row.email}})</template>
      </el-table-column>
    </el-table>
    <span slot="footer" class="dialog-footer">
      <el-button @click="edit_dig_visible = false" icon="el-icon-close">取 消</el-button>
      <el-button type="primary" @click="handleSelected" icon="el-icon-check">选 中</el-button>
    </span>
  </el-dialog>
</template>

<script>
import userApi from "@/api/UserApi";

export default {
  name: "UserMailSelector",
  data() {
    return {
      edit_dig_visible: false,
      edit_dig_title: "",
      alarm_email: "",
      user_mail_data: [],
      user_mail_selected: []
    };
  },
  methods: {
    initPage(alarmEmail) {
      this.edit_dig_title = "用户邮箱选择";
      this.alarm_email = alarmEmail;
      userApi.getUsersForMailSelect().then(res => {
        this.user_mail_data = res.data;
        this.edit_dig_visible = true;
      });
    },
    handleUserMailSelectionChange(result) {
      this.user_mail_selected = result;
    },
    handleSelected() {
      for (let i in this.user_mail_selected) {
        let result = this.user_mail_selected[i];
        if (this.alarm_email == "") {
          this.alarm_email += result.email;
        } else {
          if (!this.separatorInclude(this.alarm_email, result.email)) {
            this.alarm_email += "|" + result.email;
          }
        }
      }
      this.$emit("refreshList", this.alarm_email);
      this.edit_dig_visible = false;
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
      this.alarm_email = "";
      this.user_mail_data = [];
      this.user_mail_selected = [];
    }
  }
};
</script>