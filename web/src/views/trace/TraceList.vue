<template>
  <div class="table">
    <div class="container">
      <div class="handle-box">
        <el-select
          size="small"
          v-model="search_job_name"
          clearable
          filterable
          remote
          reserve-keyword
          placeholder="请输入任务名称"
          class="handle-input"
          :remote-method="buildJobSelectionList"
          :loading="job_selection_list_loading"
        >
          <el-option
            v-for="item in job_selection_list"
            :key="'trace.' + item.id + '.job'"
            :label="item.name"
            :value="item.name"
          ></el-option>
        </el-select>

        <el-date-picker
          size="small"
          v-model="search_time_range"
          type="datetimerange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          align="right"
          :picker-options="date_picker_options"
        >></el-date-picker>
        <el-select
          size="small"
          v-model="search_execute_status"
          class="handle-select2"
          placeholder="请选执行状态"
        >
          <el-option label value></el-option>
          <el-option label="执行成功" value="1"></el-option>
          <el-option label="执行失败" value="0"></el-option>
        </el-select>
        <el-select
          size="small"
          v-model="search_schedule_type"
          class="handle-select"
          placeholder="请选调度类型"
        >
          <el-option label value></el-option>
          <el-option label="定时" value="1"></el-option>
          <el-option label="手动" value="0"></el-option>
          <el-option label="补偿" value="2"></el-option>
        </el-select>
        <el-button size="small" type="primary" icon="el-icon-search" @click="handleSearch">搜索</el-button>
        <el-button size="small" type="info" icon="el-icon-delete-solid" @click="handleCleanEdit">清理日志</el-button>
      </div>
      <el-table
        :data="table_date"
        border
        style="width: 100%"
        :row-style="{height:'36px'}"
        :header-row-style="{height:'36px'}"
        :cell-style="{padding:'1px'}"
      >
        <el-table-column prop="jobName" label="任务名称" width="250" align="center"/>
        <el-table-column label="调度类型" width="80" align="center">
          <template slot-scope="scope">
            <span v-if="scope.row.scheduleType==0">手动</span>
            <span v-if="scope.row.scheduleType==1">定时</span>
            <span v-if="scope.row.scheduleType==2">补偿</span>
            <span v-if="scope.row.scheduleType==3">子任务</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template slot-scope="scope">
            <el-tag v-if="scope.row.executeStatus==1" size="small" type="success">执行成功</el-tag>
            <el-tag v-if="scope.row.executeStatus==0" size="small" type="danger">执行失败</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="开始时间" width="160" align="center" :formatter="startTimeFmt"/>
        <el-table-column label="结束时间" width="160" align="center" :formatter="endTimeFmt"/>
        <el-table-column prop="executeResult" label="信息" align="center"/>
        <el-table-column label="操作" align="center" width="100">
          <template slot-scope="scope">
            <el-button size="mini" type="text" @click="handleStepView(scope.row.id)">查看明细</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination">
        <el-pagination
          background
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
          :page-sizes="[10,20, 50, 100]"
          :page-size="page_size"
          layout="total, sizes, prev, pager, next, jumper"
          :total="table_data_total"
          :current-page.sync="page_num"
        ></el-pagination>
      </div>
    </div>
    <step-view ref="step_view"></step-view>
    <clean-edit ref="clean_edit" @refreshList="getData"></clean-edit>
  </div>
</template>

<script>
import traceApi from "@/api/TraceApi";
import jobApi from "@/api/JobApi";
import { formatDate } from "@/utils/date";
import stepView from "@/views/trace/StepView";
import cleanEdit from "@/views/trace/CleanEdit";

export default {
  name: "TraceList",
  components: {
    stepView
    ,cleanEdit
  },
  data() {
    return {
      search_job_name: "",
      search_time_range: "",
      search_execute_status: "",
      search_schedule_type: "",
      table_date: [],
      table_data_total: 0,
      page_num: 1,
      page_size: 10,
      date_picker_options: {},
      job_selection_list: [],
      job_selection_list_loading: false
    };
  },
  created() {
    this.buildDatePickerOptions();
    this.getData();
  },
  methods: {
    // 页码变动
    handleCurrentChange(val) {
      this.page_num = val;
      this.getData();
    },
    // 条数变动
    handleSizeChange(val) {
      this.page_size = val;
      this.getData();
    },
    // 检索
    handleSearch() {
      this.page_num = 1;
      this.getData();
    },
    // 获取数据
    getData() {
      let search_start_time = "";
      let search_end_time = "";
      if (this.search_time_range) {
        search_start_time = parseInt(
          this.search_time_range[0].valueOf() / 1000
        );
        search_end_time = parseInt(this.search_time_range[1].valueOf() / 1000);
      }
      if(this.search_job_name==null){
        this.search_job_name = ''
      }
      traceApi.getTraces({
          page_num: this.page_num,
          page_size: this.page_size,
          job_name: this.search_job_name + "",
          start_time: search_start_time + "",
          end_time: search_end_time,
          execute_status: this.search_execute_status,
          schedule_type: this.search_schedule_type
        }).then(res => {
          this.table_date = res.data;
          this.table_data_total = res.total;
        });
    },
    handleStepView(id) {
      this.$refs.step_view.initPage(id);
    },
    handleCleanEdit() {
      this.$refs.clean_edit.initPage();
    },
    startTimeFmt(row, column) {
      let date = new Date(row.startTime * 1000);
      return formatDate(date, "yyyy-MM-dd hh:mm:ss");
    },
    endTimeFmt(row, column) {
      let date = new Date(row.endTime * 1000);
      return formatDate(date, "yyyy-MM-dd hh:mm:ss");
    },
    buildJobSelectionList(name) {
      if (name !== "") {
        this.job_selection_list_loading = true;
        jobApi.getJobSelections(name).then(res => {
            this.job_selection_list = res.data;
            this.job_selection_list_loading = false;
        });
      }
    },
    buildDatePickerOptions() {
      this.date_picker_options = {
        shortcuts: [
          {
            text: "最近一小时",
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - 3600 * 1000);
              picker.$emit("pick", [start, end]);
            }
          },
          {
            text: "最近一天",
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - 3600 * 1000 * 24);
              picker.$emit("pick", [start, end]);
            }
          },
          {
            text: "最近一周",
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
              picker.$emit("pick", [start, end]);
            }
          },
          {
            text: "最近一个月",
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - 3600 * 1000 * 24 * 30);
              picker.$emit("pick", [start, end]);
            }
          },
          {
            text: "最近三个月",
            onClick(picker) {
              const end = new Date();
              const start = new Date();
              start.setTime(start.getTime() - 3600 * 1000 * 24 * 90);
              picker.$emit("pick", [start, end]);
            }
          }
        ]
      };
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
  width: 130px;
  display: inline-block;
  margin-right: 10px;
}
.handle-select2 {
  width: 130px;
  display: inline-block;
  margin-left: 10px;
  margin-right: 10px;
}
.del-dialog-cnt {
  font-size: 16px;
  text-align: center;
}
</style>
