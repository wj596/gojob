<template>
  <div class="table">
    <div class="container">
      <div class="handle-box">
        <el-input
          size="small"
          v-model="search_user_name"
          clearable
          @clear="handleSearch"
          placeholder="请输入用户名称"
          class="handle-input"
        ></el-input>
        <el-button size="small" type="primary" icon="el-icon-search" @click="handleSearch">搜索</el-button>
        <el-button size="small" type="primary" icon="el-icon-plus" @click="handleAdd">添加</el-button>
      </div>
      <el-table
        :data="table_date"
        border
        style="width: 100%"
        ref="multipleTable"
        :row-style="{height:'36px'}"
        :header-row-style="{height:'36px'}"
        :cell-style="{padding:'1px'}"
      >
        <el-table-column prop="name" label="用户名" width="400" align="center"/>
        <el-table-column prop="email" label="电子邮箱地址" align="center"/>
        <el-table-column label="操作" width="200" align="center">
          <template slot-scope="scope">
            <el-button size="mini" type="text" @click="handleEdit(scope.row.name)">编辑</el-button>
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
    <user-edit ref="user_edit" @refreshList="getData"></user-edit>
  </div>
</template>

<script>
import userApi from "@/api/UserApi";
import { formatDate } from "@/utils/date";
import userEdit from "@/views/user/UserEdit";

export default {
  name: "UserList",
  components: {
    userEdit
  },
  data() {
    return {
      search_user_name: "",
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
    // 获取数据
    getData() {
      userApi
        .getUsers({
          page_num: this.page_num,
          page_size: this.page_size,
          name: this.search_user_name
        })
        .then(res => {
          this.table_date = res.data;
          this.table_data_total = res.total;
        });
    },
    // 新增
    handleAdd() {
      this.$refs.user_edit.initPage();
    },
    handleEdit(id) {
      this.$refs.user_edit.initPage(id);
    },
    handleDelete(id) {
      this.$confirm("此操作将永久删除该记录, 是否继续?", "提示", {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning"
      })
        .then(() => {
          userApi.deleteUser(id, status).then(res => {
            this.getData();
          });
        })
        .catch(() => {
          this.$message({
            type: "info",
            message: "已取消删除"
          });
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
