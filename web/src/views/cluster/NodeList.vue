<template>
  <div class="table">
    <div class="handle-box">
      <el-button
        size="small"
        type="primary"
        icon="el-icon-refresh"
        @click="handleRefresh"
      >刷新</el-button>
    </div>
    <div class="container">
      <el-table
        :data="table_date"
        border
        style="width: 100%"
        ref="multipleTable"
        :row-style="{height:'36px'}"
        :header-row-style="{height:'36px'}"
        :cell-style="{padding:'1px'}"
      >
        <el-table-column label="节点" align="center">
            <template slot-scope="scope">
                {{scope.row.nodeName}}（{{scope.row.tcpAddr}}）
            </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="150" align="center">
            <template slot-scope="scope">
              <span v-if="scope.row.status=='-1'"></span>
              <span v-if="scope.row.status=='0'">Follower</span>
              <span v-if="scope.row.status=='1'">Candidate</span>
              <span style="color:#409EFF;" v-if="scope.row.status=='2'">Leader</span>
            </template>
        </el-table-column>
        <el-table-column prop="suffrage" label="类型" width="150" align="center"/>
        <el-table-column prop="lastContact" label="最近活跃时间" width="300" align="center"/>
        <el-table-column label="操作" width="100" align="center">
          <template slot-scope="scope">
            <el-button
              size="mini"
              type="text"
              v-if="scope.row.allowOffline=='1'"
              @click="handleOffline(scope.row.nodeName)"
            >下线</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script>
import clusterApi from "@/api/ClusterApi";

export default {
  name: "NodeList",
  data() {
    return {      
      table_date: []
    };
  },
  created() {
    this.getData();
  },
  methods: {
    // 获取数据
    getData() {
      clusterApi.getNodes().then(res => {
        this.table_date = res.data;
        this.table_data_total = res.total;
      });
    },
    handleRefresh(){
        this.getData();
    },
    handleOffline(nodeName){
      clusterApi.removeNode(nodeName).then(res => {
          this.getData();
      });
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
