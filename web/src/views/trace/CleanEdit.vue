<template>
  <!-- 编辑弹出框 -->
  <el-dialog
    :title="edit_dig_title"
    :close-on-click-modal="false"
    :visible.sync="edit_dig_visible"
    @close="handleEditDigClose"
  >
    <el-form ref="clean_edit_form" :model="form" label-width="120px" size="mini">
      <el-form-item label="任务" prop="jobId">
        <el-select v-model="form.jobId" placeholder="请选择任务" class="handle-select mr10">
          <el-option label="全部任务" value="0"></el-option>
          <el-option
            v-for="item in job_selection_list"
            :key="'clean_trace.' + item.id + '.job'"
            :label="item.name"
            :value="item.id"
          ></el-option>
        </el-select>
      </el-form-item>
      <el-form-item
        label="清理范围"
        prop="scope"
        :rules="[{ required: true, message: '请选择清理范围', trigger: 'blur'}]"
      >
        <el-select v-model="form.scope" placeholder="请选择清理范围" class="handle-select mr10">
          <el-option label="清理全部日志数据" value="1"></el-option>
          <el-option label="清理一周前的日志数据" value="2"></el-option>
          <el-option label="清理一个月前的日志数据" value="3"></el-option>
          <el-option label="清理二个月前的日志数据" value="4"></el-option>
          <el-option label="清理三个月前的日志数据" value="5"></el-option>
          <el-option label="清理六个月前的日志数据" value="6"></el-option>
          <el-option label="清理一年前的日志数据" value="7"></el-option>
        </el-select>
      </el-form-item>
    </el-form>
    <span slot="footer" class="dialog-footer">
      <el-button @click="edit_dig_visible = false">取 消</el-button>
      <el-button type="primary" @click="saveEdit">清 理</el-button>
    </span>
  </el-dialog>
</template>

<script>
import jobApi from "@/api/JobApi";
import traceApi from "@/api/TraceApi";

export default {
  name: "CleanEdit",
  data() {
    return {
      edit_dig_visible: false,
      edit_dig_title: "",
      form: {},
      job_selection_list: []
    };
  },
  methods: {
    initPage() {
      this.edit_dig_title = "清理日志";
      this.form = {
        jobId: "0",
        scope: ""
      };
      jobApi.getJobSelections("").then(res => {
        this.job_selection_list = res.data;
        this.edit_dig_visible = true;
      });
    },
    saveEdit() {
      this.$refs.clean_edit_form.validate(valid => {
        if (valid) {
          traceApi.cleanTrace(this.form).then(res => {
            this.edit_dig_visible = false;
            this.$emit("refreshList");
            this.$message.success(`日志数据清理成功`);
          });
        } else {
          console.log("error submit!!");
          return;
        }
      });
    },
    handleEditDigClose() {
      this.form = {};
      this.$refs.clean_edit_form.resetFields();
    }
  }
};
</script>