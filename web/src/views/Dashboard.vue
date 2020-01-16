<template>
  <div>
    <el-row :gutter="20" class="mgb20">
      <el-col :span="24">
        <el-button
          icon="el-icon-refresh"
          size="mini"
          @click="handleRuntimeRefresh"
          style="float: right; margin-right: 10px;"
        >刷新</el-button>
      </el-col>
    </el-row>
    <el-row :gutter="20" class="mgb20">
      <el-col :span="6">
        <el-card shadow="hover" :body-style="{padding: '0px'}">
          <div class="grid-content grid-con-4">
            <i class="el-icon-monitor grid-con-icon"></i>
            <div class="grid-cont-left">
              <div class="user-info-list">&nbsp;&nbsp;&nbsp;&nbsp;运行模式：{{runModeName}}</div>
              <div class="user-info-list">&nbsp;&nbsp;&nbsp;&nbsp;启动时间：{{runtime.startTime}}</div>
              <div class="user-info-list" v-if="runtime.runMode=='cluster'">
                &nbsp;&nbsp;&nbsp;&nbsp;调度节点：正常{{runtime.usableNodeAmount}} 、
                <font
                  style="color:red"
                >故障{{runtime.disabledNodeAmount}}</font>
              </div>
              <div class="user-info-list">
                数据库节点：正常{{runtime.usableDBAmount}} 、
                <font
                  style="color:red"
                >故障{{runtime.disabledDBAmount}}</font>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" :body-style="{padding: '0px'}">
          <div class="grid-content grid-con-1">
            <i class="el-icon-news grid-con-icon"></i>
            <div class="grid-cont-right">
              <div class="grid-num">{{runtime.jobCount}}</div>
              <div>任务数量</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" :body-style="{padding: '0px'}">
          <div class="grid-content grid-con-2">
            <i class="el-icon-copy-document grid-con-icon"></i>
            <div class="grid-cont-right">
              <div class="grid-num">{{runtime.executeNodeCount}}</div>
              <div>执行节点数量</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" :body-style="{padding: '0px'}">
          <div class="grid-content grid-con-3">
            <i class="el-icon-s-operation grid-con-icon"></i>
            <div class="grid-cont-right">
              <div class="grid-num">{{runtime.triggerTimes}}</div>
              <div>总调度次数</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="hover" :body-style="{ height: '304px'}">
      <div slot="header" class="clearfix">
        <span>失败率TOP20</span>
        <el-button
          icon="el-icon-refresh"
          size="small"
          @click="handleStatisticRefresh"
          style="float: right;"
        >刷新</el-button>&nbsp;
        <el-select
          @change="handleStatisticChange"
          v-model="range"
          placeholder="请选择"
          size="small"
          style="float: right; margin-right: 10px;"
        >
          <el-option label="今天" value="today"></el-option>
          <el-option label="本周" value="week"></el-option>
          <el-option label="本月" value="month"></el-option>
          <el-option label="全部" value="all"></el-option>
        </el-select>
      </div>
      <div id="myChart" :style="{ height: '300px'}"></div>
    </el-card>
  </div>
</template>

<script>
import runtimeApi from "@/api/RuntimeApi";
import traceApi from "@/api/TraceApi";
import { Loading } from "element-ui";
export default {
  name: "dashboard",
  data() {
    return {
      range: "today",
      runModeName: "",
      runtime: {},
      loading: {}
    };
  },
  methods: {
    handleRuntimeRefresh() {
      runtimeApi.getRuntime().then(res => {
        if (res.data) {
          this.runtime = res.data;
          if (res.data.runMode === "standalone") {
            this.runModeName = "单机";
          } else {
            this.runModeName = "集群";
          }
        }
      });
    },
    startLoading() {
      this.loading = Loading.service({
        lock: true,
        text: "统计中... ...",
        target: document.querySelector("#myChart"),
        background: "rgba(0,0,0,0.1)"
      });
    },
    handleStatisticRefresh(){
      this.handleStatisticChange();
    },
    handleStatisticChange() {
      if ("today" === this.range) {
        this.statisticToday();
      }
      if ("week" === this.range) {
        this.statisticWeek();
      }
      if ("month" === this.range) {
        this.statisticMonth();
      }
      if ("all" === this.range) {
        this.statisticAll();
      }
    },
    statisticToday() {
      this.startLoading();
      traceApi
        .statisticToday()
        .then(res => {
          if (res.data) {
            this.handleChart(res.data);
          }
          this.loading.close();
        })
        .catch(error => {
          this.loading.close();
        });
    },
    statisticWeek() {
      this.startLoading();
      traceApi
        .statisticWeek()
        .then(res => {
          if (res.data) {
            this.handleChart(res.data);
          }
          this.loading.close();
        })
        .catch(error => {
          this.loading.close();
        });
    },
    statisticMonth() {
      this.startLoading();
      traceApi
        .statisticMonth()
        .then(res => {
          if (res.data) {
            this.handleChart(res.data);
          }
          this.loading.close();
        })
        .catch(error => {
          this.loading.close();
        });
    },
    statisticAll() {
      this.startLoading();
      traceApi
        .statisticAll()
        .then(res => {
          if (res.data) {
            this.handleChart(res.data);
          }
          this.loading.close();
        })
        .catch(error => {
          this.loading.close();
        });
    },
    handleChart(data) {
      var myChart = this.$echarts.init(document.getElementById("myChart"));
      let xData = [];
      let yData = [];
      for (let i = 0; i < data.length; i++) {
        let item = data[i];
        xData.push(item.name);
        yData.push(item.rate);
      }
      myChart.setOption({
        color: ["#3398DB"],
        tooltip: {
          trigger: "axis",
          formatter(params) {
            let item = params[0];
            let dataIndex = item.dataIndex;
            let h = "<div><p>" + data[dataIndex].name + "</p></div>";
            h += "<div><p>调度次数：" + data[dataIndex].total + "</p></div>";
            h +=
              "<div><p>调度成功次数：" + data[dataIndex].succeed + "</p></div>";
            h +=
              "<div><p>调度失败次数：" + data[dataIndex].failed + "</p></div>";
            h += "<div><p>调度失败率：" + data[dataIndex].rate + "% </p></div>";
            return h;
          }
        },
        grid: {
          left: "3%",
          right: "4%",
          bottom: "3%",
          containLabel: true
        },
        xAxis: [
          {
            type: "category",
            data: xData,
            axisLine: { lineStyle: { color: "#008acd" } },
            axisLabel: {
              interval: 0,
              formatter: function(value, index) {
                return value;
              }
            }
          }
        ],
        yAxis: [
          {
            type: "value",
            axisLabel: {
              show: true,
              interval: "auto",
              formatter: "{value} %"
            },
            show: true
          }
        ],
        series: [
          {
            name: "执行失败率",
            type: "bar",
            barWidth: 25,
            data: yData
          }
        ]
      });
    }
  },
  mounted() {
    this.handleRuntimeRefresh();
    this.statisticToday();
  },
  computed: {
    role() {
      return this.name === "admin" ? "超级管理员" : "普通用户";
    }
  }
};
</script>


<style scoped>
.el-row {
  margin-bottom: 10px;
}

.grid-content {
  display: flex;
  align-items: center;
  height: 100px;
}

.grid-cont-left {
  margin-left: 10px;
  flex: 1;
  text-align: left;
  font-size: 12px;
  color: #999;
}

.grid-cont-right {
  flex: 1;
  text-align: center;
  font-size: 12px;
  color: #999;
}

.grid-num {
  font-size: 30px;
  font-weight: bold;
}

.grid-con-icon {
  font-size: 50px;
  width: 100px;
  height: 100px;
  text-align: center;
  line-height: 100px;
  color: #fff;
}

.grid-con-1 .grid-con-icon {
  background: rgb(45, 140, 240);
}

.grid-con-1 .grid-num {
  color: rgb(45, 140, 240);
}

.grid-con-2 .grid-con-icon {
  background: rgb(100, 213, 114);
}

.grid-con-2 .grid-num {
  color: rgb(45, 140, 240);
}

.grid-con-3 .grid-con-icon {
  background: rgb(242, 94, 67);
}

.grid-con-3 .grid-num {
  color: rgb(242, 94, 67);
}

.grid-con-4 .grid-con-icon {
  background: #909399;
}

.grid-con-4 .grid-num {
  color: #909399;
}

.user-info {
  display: flex;
  align-items: center;
  padding-bottom: 20px;
  border-bottom: 2px solid #ccc;
  margin-bottom: 20px;
}

.user-avator {
  width: 120px;
  height: 120px;
  border-radius: 50%;
}

.user-info-cont {
  padding-left: 50px;
  flex: 1;
  font-size: 14px;
  color: #999;
}

.user-info-cont div:first-child {
  font-size: 30px;
  color: #222;
}

.user-info-list {
  font-size: 12px;
  color: #000;
  line-height: 17px;
}

.user-info-list span {
  margin-left: 70px;
}

.mgb20 {
  margin-bottom: 10px;
}
.mgt20 {
  margin-top: 20px;
}

.todo-item {
  font-size: 14px;
}

.todo-item-del {
  text-decoration: line-through;
  color: #999;
}
</style>