<template>
  <div class="table">
    <div class="handle-box"></div>
    <div class="container">
      <el-card class="box-card">
        <div slot="header" class="clearfix">
          <span style="color: #999;font-weight:bold;">
            SMTP(简单邮件传输协议)属性
          </span>

          <el-button style="float: right; padding: 3px 0" type="text" @click="handleTest">测试</el-button>

          <el-button
            style="float: right; padding: 3px 0 10px 0;margin-right:10px;"
            type="text"
            @click="handleEdit"
          >编辑</el-button>
        </div>
        <div style="margin-bottom: 5px;"><span style="font-size:14px;color:#999;font-weight:bold;">地址:</span>&nbsp;&nbsp;{{entity.smtpHost}}</div>
        <div style="margin-bottom: 5px;" v-if="entity.smtpPort>0"><span style="font-size:14px;color:#999;font-weight:bold;">端口:</span>&nbsp;&nbsp;{{entity.smtpPort}}</div>
        <div style="margin-bottom: 5px;" v-else><span style="font-size:14px;color:#999;font-weight:bold;">端口:</span></div>
        <div style="margin-bottom: 5px;"><span style="font-size:14px;color:#999;font-weight:bold;">用户:</span>&nbsp;&nbsp;{{entity.smtpUser}}</div>
        <div style="margin-bottom: 5px;" v-if="entity.smtpPassword"><span style="font-size:14px;color:#999;font-weight:bold;">密码:</span>&nbsp;&nbsp;***</div>
        <div style="margin-bottom: 5px;" v-else><span style="font-size:14px;color:#999;font-weight:bold;">密码:</span></div>
      </el-card>
      <br>
      <el-card class="box-card">
        <div slot="header" class="clearfix">
          <span style="color: #999;font-weight:bold;">
            系统故障告警地址
          </span>

          <el-button
            style="float: right; padding: 3px 0 10px 0;margin-right:10px;"
            type="text"
            @click="handleSysEdit"
          >编辑</el-button>
        </div>
        <div style="font-size: 16px;margin-bottom: 5px;"><span style="font-size:14px;color:#999;font-weight:bold;">邮件地址:</span>&nbsp;&nbsp;{{entity.sysAlarmEmail}}</div>
        <br>
        <div style="font-size: 14px;margin-bottom: 5px;color: #999">(数据库故障、集群节点故障会向这个地址发送告警邮件；具体任务的告警地址请在任务管理中配置)</div>
      </el-card>
    </div>
    <alarm-edit ref="alarm_edit" @refreshList="getData"></alarm-edit>
    <sys-alarm-edit ref="sys_alarm_edit" @refreshList="getData"></sys-alarm-edit>
    <alarm-test-edit ref="alarm_test_edit" @refreshList="getData"></alarm-test-edit>
  </div>
</template>

<script>
import alarmApi from "@/api/AlarmApi";
import alarmEdit from "@/views/alarm/AlarmEdit";
import alarmTestEdit from "@/views/alarm/AlarmTestEdit";
import sysAlarmEdit from "@/views/alarm/SysAlarmEdit";

export default {
  name: "AlarmList",
  components: {
    alarmEdit,
    alarmTestEdit,
    sysAlarmEdit
  },
  data() {
    return {
      entity: {}
    };
  },
  mounted() {
    this.getData();
  },
  methods: {
    getData() {
      alarmApi.getAlarmConfig().then(res => {
        if (res.data) {
          this.entity = res.data;
        }
      });
    },
    handleEdit() {
      this.$refs.alarm_edit.initPage(
        this.entity.smtpHost,
        this.entity.smtpPort,
        this.entity.smtpUser,
        this.entity.smtpPassword
      );
    },
    handleSysEdit() {
      this.$refs.sys_alarm_edit.initPage();
    },
    handleTest() {
      this.$refs.alarm_test_edit.initPage();
    }
  }
};
</script>
<style scoped>
.handle-box {
  margin-bottom: 10px;
}
.handle-input {
  width: 200px;
  display: inline-block;
  margin-right: 10px;
}
.handle-select {
  width: 100px;
  display: inline-block;
  margin-right: 10px;
}
.del-dialog-cnt {
  font-size: 16px;
  text-align: center;
}
</style>
