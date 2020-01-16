<template>
  <!-- 编辑弹出框 -->
  <div>
    <el-dialog
      width="80%"
      title="帮助"
      :close-on-click-modal="false"
      :visible.sync="edit_dig_visible"
      :fullscreen="false"
    >
      <div style="overflow: auto;">
        <table class="help_table">
          <tr>
            <td class="help_table_label">Cron表达式</td>
            <td class="help_table_td">
              填写标准Cron表达即可，可百度。如：
              <br>每10分钟执行一次：0 */10 * * * ?
              <br>每天中午12点执行：0 0 12 * * ?
              <br>每月15日上午10:15执行：0 15 10 15 * ?
            </td>
          </tr>
          <tr>
            <td class="help_table_label">任务执行方法</td>
            <td class="help_table_td">
              假如对账服务的URL地址是：http://192.168.16.10:8080/store/bills/check
              <br>任务执行方法应该填写：/store/bills/check ;即
              <br>192.168.16.10:8080即为执行节点
            </td>
          </tr>
          <tr>
            <td class="help_table_label">附加参数</td>
            <td class="help_table_td">
              Header参数的格式为："key=value"键值对，多个用"|"分开
              <br>假如我们填写：user_name=timedTask
              <br>调用对账服务时会带上这个参数：http://192.168.16.10:8080/store/bills/check?user_name=timedTask
              <br>假如我们填写：user_name=timedTask|operation_type=auto
              <br>调用对账服务时会带上这两个参数：http://192.168.16.10:8080/store/bills/check?user_name=timedTask&operation_type=auto
              <br>
            </td>
          </tr>
          <tr>
            <td class="help_table_label">Header参数</td>
            <td class="help_table_td">
              Header参数的格式为："key=value"键值对，多个用"|"分开
              <br>假如我们填写：Authorization=wangjie
              <br>调用对账服务时会在请求头(header)中添加属性Authorization，值为wangjie
            </td>
          </tr>
          <tr>
            <td class="help_table_label">执行超时时间</td>
            <td class="help_table_td">
              调用"任务执行方法"的超时间，单位为"秒"
              <br>假如我们填写的超时时间为30，如果调用对账服务30秒内没有响应,即认为调用失败。
            </td>
          </tr>
          <tr>
            <td class="help_table_label">执行失败重试次数</td>
            <td class="help_table_td">
              调用"任务执行方法"失败，进行指定次数的重试
              <br>假如我们填写的执行失败重试次数为3，如果调用对账服务失败会进行重试，直到调用成功或者重试次数等于3
              <br>主要是防止网络抖动引起的调用失败，调用失败的情况包括"任务执行方法"未返回200状态和调用超时
            </td>
          </tr>
          <tr>
            <td class="help_table_label">执行失败重试间隔</td>
            <td class="help_table_td">调用"任务执行方法"失败，进行指定次数的重试，每次重试间的间隔时间，单位为"秒"</td>
          </tr>
          <tr>
            <td class="help_table_label">执行失败处理策略</td>
            <td class="help_table_td">
              假如对账服务在3个节点进行了部署，分别是：
              <br>http://192.168.16.10:8080/store/bills/check
              <br>http://192.168.16.11:8080/store/bills/check
              <br>http://192.168.16.12:8080/store/bills/check
              <br>当前任务选中的执行节点是192.168.16.10:8080，调用对账服务失败，会根据"执行失败重试次数"和"执行失败重试间隔"进行重试，重试数次后认定为最终失败。
              <br>当"执行失败处理策略"选择"失败安全"，则调度系统会默不作声的记录下日志，结束任务，等待下次调度。
              <br>当"执行失败处理策略"选择"失败转移"，则调度系统会尝试调用执行节点"192.168.16.11:8080"上的对账服务，如果还是失败，会尝试调用执行节点"192.168.16.12:8080"上的对账服务，直到调用成功或无执行节点

              


            </td>
          </tr>
        </table>
      </div>

      <span slot="footer" class="dialog-footer">
        <el-button @click="edit_dig_visible = false" icon="el-icon-close">关 闭</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
export default {
  name: "JobEditHelp",
  data() {
    return {
      edit_dig_visible: false
    };
  },
  methods: {
    initPage() {
      this.edit_dig_visible = true;
    }
  }
};
</script>
<style type="text/css">
.help_table {
  width: 100%;
  font-family: verdana, arial, sans-serif;
  font-size: 11px;
  color: #333333;
  border-width: 1px;
  border-color: #dcdfe6;
  border-collapse: collapse;
}
.help_table_label {
  width: 5%;
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
.help_table_td {
  border-width: 1px;
  padding: 8px;
  border-style: solid;
  border-color: #dcdfe6;
  background-color: #ffffff;
  font-size: 13px;
}
</style>