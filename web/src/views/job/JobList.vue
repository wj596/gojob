<template>
  <div class="table">
    <div class="container">
      <div class="handle-box">
        <el-input
          size="small"
          v-model="search_job_name"
          clearable
          @clear="handleSearch"
          placeholder="请输入任务名称"
          class="handle-input"
        ></el-input>
        <el-input
          size="small"
          clearable
          @clear="handleSearch"
          v-model="search_job_creator"
          placeholder="请输入创建人"
          class="handle-input"
        ></el-input>
        <el-select
          size="small"
          v-model="search_job_status"
          class="handle-select"
          placeholder="请选择状态"
        >
          <el-option label value></el-option>
          <el-option label="正常" value="1"></el-option>
          <el-option label="挂起" value="0"></el-option>
        </el-select>
        <el-button size="small" type="primary" icon="el-icon-search" @click="handleSearch">搜索</el-button>
        <el-button size="small" type="primary" icon="el-icon-plus" @click="handleAdd">添加</el-button>
      </div>
      <el-table
        :data="table_date"
        border
        style="width: 100%"
        :row-style="{height:'36px'}"
        :header-row-style="{height:'36px'}"
        :cell-style="{padding:'1px'}"
      >
        <el-table-column prop="name" label="名称" width="250" align="center"/>
        <el-table-column prop="cron" label="Cron表达式" width="170" align="center"/>
        <el-table-column label="状态" width="65" align="center">
          <template slot-scope="scope">
            <el-tag v-if="scope.row.status==1" size="small" type="success">正常</el-tag>
            <el-tag v-if="scope.row.status==0" size="small" type="danger">挂起</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="执行节点数量/选择策略" width="200" align="center">
          <template slot-scope="scope">
            <span v-if="scope.row.executorCount==0">无</span>
            <span v-else>{{scope.row.executorCount}} 个/
            <span v-if="scope.row.executorSelectStrategy=='random'">随机</span>
            <span v-if="scope.row.executorSelectStrategy=='round'">轮询</span>
            <span v-if="scope.row.executorSelectStrategy=='weight_random'">加权随机</span>
            <span v-if="scope.row.executorSelectStrategy=='weight_round'">加权轮询</span>
            <span v-if="scope.row.executorSelectStrategy=='sharding'">分片</span>
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="creator" width="80" label="创建人" align="center"/>
        <el-table-column width="140" label="创建时间" align="center" :formatter="createTimeTimeFmt"/>
        <el-table-column label="操作" align="center">
          <template slot-scope="scope">
            <el-button v-if="scope.row.status==1" size="mini" type="text" @click="handleView(scope.row.id)">查看</el-button>
            <el-button
              size="mini"
              v-if="scope.row.status==0"
              @click="handleStatus(scope.row.id,scope.row.status)"
              type="text"
            >启动</el-button>
            <el-button
              size="mini"
              v-if="scope.row.status==1"
              @click="handleStatus(scope.row.id,scope.row.status)"
              type="text"
            >挂起</el-button>
            <el-button size="mini" type="text" @click="handleEdit(scope.row.id)">编辑</el-button>
            <el-button v-if="scope.row.status==1" size="mini" type="text" @click="handleLaunch(scope.row.id)">触发一次</el-button>
            <el-button size="mini" type="text" @click="handleDelete(scope.row.id)">删除</el-button>
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

    <job-view ref="job_view"></job-view>
    <job-edit ref="job_edit" @refreshList="getData"></job-edit>
  </div>
</template>

<script>
import jobApi from "@/api/JobApi";
import { formatDate } from "@/utils/date";
import jobView from "@/views/job/JobView";
import jobEdit from "@/views/job/JobEdit";

export default {
  name: "JobList",
  components: {
    jobEdit,
    jobView
  },
  data() {
    return {
      search_job_name: "",
      search_job_creator: "",
      search_job_status: "",
      table_date: [],
      table_data_total: 0,
      page_num: 1,
      page_size: 10
    };
  },
  created() {
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
    // 列表选择监听
    handleSelectionChange(val) {
      this.table_selections = val;
    },
    // 获取数据
    getData() {
      jobApi
        .getJobs({
          page_num: this.page_num,
          page_size: this.page_size,
          name: this.search_job_name,
          creator: this.search_job_creator,
          status: this.search_job_status
        })
        .then(res => {
          this.table_date = res.data;
          this.table_data_total = res.total;
        });
    },
    // 新增
    handleAdd() {
      this.$refs.job_edit.initPage();
    },
    // 编辑
    handleEdit(id) {
      this.$refs.job_edit.initPage(id);
    },
    // 编辑
    handleView(id) {
      this.$refs.job_view.initPage(id);
    },
    // 编辑
    handleStatus(id, status) {
      if (status == 0) {
        status = 1;
        jobApi.updateStatus(id, status).then(res => {
          this.getData();
        });
      } else {
        status = 0;
        this.$confirm("挂起的作业,将不会被调度执行, 是否继续?", "提示", {
          confirmButtonText: "确定",
          cancelButtonText: "取消",
          type: "warning"
        })
          .then(() => {
            jobApi.updateStatus(id, status).then(res => {
              this.getData();
            });
          })
          .catch(() => {
            this.$message({
              type: "info",
              message: "已取消挂起"
            });
          });
      }
    },
    handleDelete(id) {
      this.$confirm("此操作将永久删除该记录, 是否继续?", "提示", {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning"
      })
        .then(() => {
          jobApi.deleteJob(id, status).then(res => {
            this.getData();
          });
        })
        .catch(() => {
          this.$message({
            type: "info",
            message: "已取消删除"
          });
        });
    },
    handleLaunch(id) {
      this.$confirm("此操作将会触发执行一次任务, 是否继续?", "提示", {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning"
      })
        .then(() => {
          jobApi.launchJob(id, status).then(res => {
            this.$message({
              type: "success",
              message: "任务触发成功，调度情况请到 '调度日志' 模块查看"
            });
          });
        })
        .catch(() => {
          this.$message({
            type: "info",
            message: "已取消触发"
          });
        });
    },
    createTimeTimeFmt(row, column) {
      let date = new Date(row.createTime);
      return formatDate(date, "yyyy-MM-dd hh:mm");
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
  width: 110px;
  display: inline-block;
  margin-right: 10px;
}
.del-dialog-cnt {
  font-size: 16px;
  text-align: center;
}
</style>
