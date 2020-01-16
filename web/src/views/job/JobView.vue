<template>
  <!-- 编辑弹出框 -->
  <el-dialog
    width="70%"
    :title="edit_dig_title"
    :close-on-click-modal="false"
    :visible.sync="edit_dig_visible"
  >
    <div style="overflow: auto;">
      <table class="view_table">
        <tr>
          <td class="view_table_label">作业名称</td>
          <td class="view_table_td">{{job.name}}</td>
          <td class="view_table_label">Cron表达式</td>
          <td class="view_table_td">{{job.cron}}</td>
        </tr>
        <tr>
          <td class="view_table_label">通讯协议/请求方法</td>
          <td class="view_table_td">{{job.protocol}}/get</td>
          <td class="view_table_label">任务执行方法</td>
          <td class="view_table_td">{{job.uri}}</td>
        </tr>
        <tr>
          <td class="view_table_label">附加参数</td>
          <td class="view_table_td">{{job.httpParam}}</td>
          <td class="view_table_label">Header参数</td>
          <td class="view_table_td">{{job.httpHeaderParam}}</td>
        </tr>
        <tr>
          <td class="view_table_label">数字签名</td>
          <td class="view_table_td">
            <span v-if="job.HttpSign==1">启用</span>
            <span v-else>不启用</span>
          </td>
          <td class="view_table_label">执行超时时间</td>
          <td class="view_table_td">
            <span v-if="job.timeout>0">{{job.timeout}} 秒</span>
            <span v-else>&nbsp;-&nbsp;</span>
          </td>
        </tr>
        <tr>
          <td class="view_table_label">执行失败重试次数</td>
          <td class="view_table_td">
            <span v-if="job.retryCount>0">{{job.retryCount}} 次</span>
            <span v-else>&nbsp;-&nbsp;</span>
          </td>
          <td class="view_table_label">执行失败重试间隔</td>
          <td class="view_table_td">
            <span v-if="job.retryWaitTime>0">{{job.retryWaitTime}} 秒</span>
            <span v-else>&nbsp;-&nbsp;</span>
          </td>
        </tr>
        <tr>
          <td class="view_table_label">执行失败处理策略</td>
          <td class="view_table_td">
            <span v-if="job.failTakeover==0">错误安全</span>
            <span v-if="job.failTakeover==1">错误重试</span>
          </td>
          <td class="view_table_label">misfire超时时间</td>
          <td class="view_table_td">
            <span v-if="job.misfireThreshold>0">{{job.misfireThreshold}} 秒</span>
            <span v-else>&nbsp;-&nbsp;</span>
          </td>
        </tr>
        <tr v-if="job.subJobDisplay">
          <td class="view_table_label">子任务</td>
          <td class="view_table_td" colspan="3">
            {{job.subJobDisplay}}
            &nbsp;&nbsp;&nbsp;&nbsp;
            (
            <span v-if="job.subJobScheduleStrategy==0">执行完毕触发</span>
            <span v-else-if="job.subJobScheduleStrategy==1">执行成功触发</span>
            <span v-else-if="job.subJobScheduleStrategy==2">执行失败触发</span>
            <span v-else></span>)
          </td>
        </tr>
        <tr>
          <td class="view_table_label">执行节点选择策略</td>
          <td class="view_table_td" colspan="3">
            <span v-if="job.executorSelectStrategy=='random'">随机 选择一个执行节点</span>
            <span v-if="job.executorSelectStrategy=='round'">轮询 选一个执行节点</span>
            <span v-if="job.executorSelectStrategy=='weight_random'">加权随机 选择一个执行节点</span>
            <span v-if="job.executorSelectStrategy=='weight_round'">加权轮询 选一个执行节点</span>
            <span v-if="job.executorSelectStrategy=='sharding'">分片 选择多个执行节点</span>
          </td>
        </tr>
        <tr v-if="job.executorSelectStrategy=='sharding'">
          <td class="view_table_label">分片总数</td>
          <td class="view_table_td" colspan="3">
            {{job.shardingCount}}
            &nbsp;&nbsp; &nbsp;&nbsp;分片参数：
            {{job.shardingParam}}
          </td>
        </tr>
        <tr>
          <td class="view_table_label">执行节点</td>
          <td class="view_table_td" colspan="3">
            <div v-for="(node, index) in job.executors" :key="'node.' + index + '.key3'">
              节点{{index+1}}、
              地址: {{node.address}} &emsp;
              状态:
              <span
                v-if="node.status==1"
                style="color:#67C23A;"
              >上线</span>
              <span v-if="node.status==0" style="color:#F56C6C;">下线</span>
              &emsp;
              <span v-if="displayWeight">权重:{{node.weight}}</span>
            </div>
          </td>
        </tr>
        <tr>
          <td class="view_table_label">告警邮箱</td>
          <td class="view_table_td">{{job.alarmEmail}}</td>
          <td class="view_table_label">备注</td>
          <td class="view_table_td">{{job.remark}}</td>
        </tr>
      </table>
    </div>
    <span slot="footer" class="dialog-footer">
      <el-button @click="edit_dig_visible = false">确 定</el-button>
    </span>
  </el-dialog>
</template>

<script>
import jobApi from "@/api/JobApi";

export default {
  name: "JobView",
  data() {
    return {
      edit_dig_visible: false,
      edit_dig_title: "作业明细",
      displayWeight: false,
      job: {}
    };
  },
  methods: {
    initPage(id) {
      jobApi.getJob(id).then(res => {
        this.job = res.data;
        this.job.failTakeover = res.data.failTakeover + "";
        this.job.sendTraceId = res.data.sendTraceId + "";
        if (
          this.job.executorSelectStrategy == "weight_random" ||
          this.job.executorSelectStrategy == "weight_round"
        ) {
          this.displayWeight = true;
        }
      });
      this.edit_dig_visible = true;
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
.view_table_label {
  width: 12%;
  border-width: 1px;
  padding: 8px;
  border-style: solid;
  border-color: #dcdfe6;
  background-color: #ffffff;
  text-align: right;
  color: #999;
  font-weight: bold;
  font-size: 12px;
}
.view_table_td {
  width: 40%;
  border-width: 1px;
  padding: 8px;
  border-style: solid;
  border-color: #dcdfe6;
  background-color: #ffffff;
  font-size: 13px;
}
</style>