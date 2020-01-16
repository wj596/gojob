<template>
  <!-- 编辑弹出框 -->
  <el-dialog :title="edit_dig_title" :close-on-click-modal="false" :visible.sync="edit_dig_visible">
    <div style="height: 350px;overflow: auto;">
      <table class="view_table">
        <tr v-for="(step, index) in details" :key="'step.' + index + '.key'">
          <td class="view_table_td">{{step}}</td>
        </tr>
      </table>
    </div>
    <span slot="footer" class="dialog-footer">
      <el-button @click="edit_dig_visible = false">确 定</el-button>
    </span>
  </el-dialog>
</template>

<script>
import traceApi from "@/api/TraceApi";

export default {
  name: "StepView",
  data() {
    return {
      edit_dig_visible: false,
      edit_dig_title: "调度执行明细",
      details: []
    };
  },
  methods: {
    initPage(id) {
      this.details = [];
      traceApi.getTrace(id).then(res => {
        if (res.data && res.data.executeDetail) {
          this.details = res.data.executeDetail.split("<line>");
        }
        this.edit_dig_visible = true;
      });
    }
  }
};
</script>
<style type="text/css">
.view_table {
  width: 100%;
  font-family: verdana, arial, sans-serif;
  font-size: 11px;
  color: #333333;
  border-width: 1px;
  border-color: #dcdfe6;
  border-collapse: collapse;
}
.view_table_td {
  border-width: 1px;
  padding: 8px;
  border-style: solid;
  border-color: #dcdfe6;
  background-color: #ffffff;
}
</style>