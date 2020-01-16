<template>
  <!-- 编辑弹出框 -->
  <div>
    <el-dialog
      :title="edit_dig_title"
      :close-on-click-modal="false"
      :visible.sync="edit_dig_visible"
      :fullscreen="true"
      @close="handleEditDigClose"
    >
      <el-form ref="job_edit_form" :model="form" :rules="rules" label-width="150px" size="mini">
        <el-row>
          <el-col :span="12">
            <el-form-item label="任务名称" prop="name">
              <el-input v-model="form.name" placeholder="请输入一个简短并且有意义的名称"></el-input>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Cron表达式" prop="cron">
              <el-input v-model="form.cron" placeholder="请输入Cron表达式"></el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="12">
            <el-form-item label="通讯协议/请求方法" prop="protocol">
              <el-select v-model="form.protocol" class="handle-select mr10">
                <el-option label="http/get" value="http"></el-option>
                <el-option label="https/get" value="https"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="任务执行方法" prop="uri">
              <el-input
                v-model="form.uri"
                placeholder="请输入任务方法的资源标识符(URI), 如：http://localhost:8080/bills/check 则输入：/bills/check "
              ></el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="12">
            <el-form-item label="附加参数" prop="httpParam">
              <el-input
                v-model="form.httpParam"
                placeholder="请输入附加参数，格式为name=value，多个用|分隔，如：name1=value1|name2=value2"
              ></el-input>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Header参数" prop="httpHeaderParam">
              <el-input
                v-model="form.httpHeaderParam"
                placeholder="请输入Header参数，格式为name=value，多个用|分隔，如：name1=value1|name2=value2"
              ></el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="12">
            <el-form-item label="数字签名" prop="httpSign">
              <el-select v-model="form.httpSign" class="handle-select mr10">
                <el-option label="不启用" value="0"></el-option>
                <el-option label="启用" value="1"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="执行超时时间(秒)" prop="timeout">
              <el-input-number v-model="form.timeout" :min="10" :max="300"></el-input-number>
              <span style="font-size: 13px;color: #999;">&nbsp;&nbsp;可填入范围：10 - 300（秒）</span>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="12">
            <el-form-item label="执行失败重试次数" prop="retryCount">
              <el-input-number v-model="form.retryCount" :min="0" :max="10"></el-input-number>
              <span style="font-size: 13px;color: #999;">&nbsp;&nbsp;为0 表示不重试</span>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="执行失败重试间隔(秒)" prop="retryWaitTime">
              <el-input-number v-model="form.retryWaitTime" :min="0" :max="1000"></el-input-number>
              <span style="font-size: 13px;color: #999;">&nbsp;&nbsp;为0 表示无间隔</span>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row>
          <el-col :span="12">
            <el-form-item label="执行失败处理策略" prop="failTakeover">
              <el-select v-model="form.failTakeover" class="handle-select mr10">
                <el-option label="失败转移" value="1"></el-option>
                <el-option label="失败安全" value="0"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="misfire超时时间(秒)" prop="misfireThreshold">
              <el-input-number v-model="form.misfireThreshold" :min="0" :max="1800"></el-input-number>
              <span style="font-size: 13px;color: #999;">&nbsp;&nbsp;为0 表示不启用misfire机制</span>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="12">
            <el-row>
              <el-col :span="12">
                <el-form-item label="子任务" prop="subJobIds">
                  <el-select v-model="form.subJobIds" multiple placeholder="请选择子任务,可多选">
                    <el-option
                      v-for="item in subjob_options"
                      :key="item.id"
                      :label="item.name"
                      :value="item.id"
                    ></el-option>
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item
                  v-if="form.subJobIds&&form.subJobIds.length>0"
                  label="子任务触发策略"
                  prop="subJobScheduleStrategy"
                >
                  <el-select
                    v-model="form.subJobScheduleStrategy"
                    placeholder="请选择子任务触发策略"
                    class="handle-select mr10"
                  >
                    <el-option label="调度完毕触发" value="0"></el-option>
                    <el-option label="调度成功触发" value="1"></el-option>
                    <el-option label="调度失败触发" value="2"></el-option>
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>
          </el-col>
          <el-col :span="11">
            <el-form-item label="告警邮箱" prop="alarmEmail">
              <el-input v-model="form.alarmEmail" placeholder="请输入告警邮件，多个用|分隔;点击后面的按钮，可选择系统用户的邮箱"></el-input>
            </el-form-item>
          </el-col>
          <el-col :span="1">
            &nbsp;
            <el-button
              @click="handleUserMailSelectorDig"
              type="primary"
              icon="el-icon-user-solid"
              circle
              size="mini"
            ></el-button>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="12">
            <el-form-item label="执行节点选择策略" prop="executorSelectStrategy">
              <el-select
                v-model="form.executorSelectStrategy"
                placeholder="请选择执行节点选择策略"
                class="handle-select mr10"
              >
                <el-option label="随机 选择一个执行节点" value="random"></el-option>
                <el-option label="轮询 选一个执行节点" value="round"></el-option>
                <el-option label="加权随机 选择一个执行节点" value="weight_random"></el-option>
                <el-option label="加权轮询 选一个执行节点" value="weight_round"></el-option>
                <el-option label="分片 选择多个执行节点" value="sharding"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="备注" prop="remark">
              <el-input v-model="form.remark" placeholder="请输入任务备注信息"></el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="12">
            <el-form-item
              v-if="form.executorSelectStrategy=='sharding'"
              label="分片总数"
              prop="shardingCount"
            >
              <el-input-number v-model="form.shardingCount" :min="1" :max="10000"></el-input-number>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item
              v-if="form.executorSelectStrategy=='sharding'"
              label="分片参数"
              prop="shardingParam"
            >
              <el-input
                v-model="form.shardingParam"
                placeholder="请输入分片参数多个用|分开，如：340101|340102|340103|340104"
              ></el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <fieldset style="border-color: #999;border-width: 1px;">
          <legend style="font-weight:bold;font-size: 12px;">
            <span>执行节点列表</span>
          </legend>
          <el-row>
            <el-col :span="24">
              <el-row v-for="(node, index) in form.executors" :key="'node.' + index + '.key2'">
                <el-col :span="10">
                  <el-form-item
                    label="节点地址"
                    :prop="`executors[${index}].address`"
                    :rules="[{ required: true, message: '请输入执行节点地址', trigger: 'blur'}]"
                  >
                    <el-input v-model="node.address" placeholder="请输入执行节点地址，格式为 'IP:端口'"></el-input>
                  </el-form-item>
                </el-col>
                <el-col :span="5">
                  <el-form-item
                    label="权重"
                    prop="weight"
                    v-if="form.executorSelectStrategy=='weight_random'||form.executorSelectStrategy=='weight_round'"
                  >
                    <el-input-number v-model="node.weight" :min="1" :max="100"></el-input-number>
                  </el-form-item>
                </el-col>
                <el-col :span="5">
                  <el-form-item label="状态" prop="weight">
                    <el-select v-model="node.status" class="handle-select mr10">
                      <el-option label="上线" value="1"></el-option>
                      <el-option label="下线" value="0"></el-option>
                    </el-select>
                  </el-form-item>
                </el-col>
                <el-col :span="4">
                  <el-form-item>
                    <el-button
                      v-if="index==0"
                      @click="handleAddExecutor()"
                      type="success"
                      icon="el-icon-plus"
                    >增加节点</el-button>
                    <el-button
                      v-if="index>0"
                      @click="handleDelExecutor(index)"
                      type="danger"
                      icon="el-icon-delete"
                    >删除节点</el-button>
                  </el-form-item>
                </el-col>
              </el-row>
            </el-col>
          </el-row>
        </fieldset>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button @click="handleHelp" type="warning" icon="el-icon-question">帮助</el-button>&nbsp;&nbsp;&nbsp;&nbsp;
        <el-button @click="edit_dig_visible = false" icon="el-icon-close">取 消</el-button>
        <el-button type="primary" @click="saveEdit" icon="el-icon-check">提 交</el-button>
      </span>
    </el-dialog>
    <user-mail-selector ref="user_mail_selector" @refreshList="handleUserMailSelected"></user-mail-selector>
    <job-edit-help ref="job_edit_help"></job-edit-help>
  </div>
</template>

<script>
import jobApi from "@/api/JobApi";
import userApi from "@/api/UserApi";
import jobEditHelp from "@/views/job/JobEditHelp";
import userMailSelector from "@/views/job/UserMailSelector";

export default {
  name: "JobEdit",
  components: {
    jobEditHelp,
    userMailSelector
  },
  data() {
    return {
      edit_dig_visible: false,
      subjob_options: [],
      edit_dig_title: "",
      form: this.emptyForm(),
      rules: this.validRules()
    };
  },
  methods: {
    initPage(id) {
      if (id) {
        this.edit_dig_title = "编辑任务";
        jobApi.getJob(id).then(res => {
          this.form = res.data;
          this.form.httpSign = res.data.httpSign + "";
          this.form.failTakeover = res.data.failTakeover + "";
          this.form.subJobScheduleStrategy =
            res.data.subJobScheduleStrategy + "";
          for (let index in this.form.executors) {
            let node = this.form.executors[index];
            node.status = node.status + "";
          }
          jobApi.getSubJobSelections(id).then(res => {
            this.subjob_options = res.data;
          });
          this.edit_dig_visible = true;
        });
      } else {
        jobApi.getSubJobSelections('').then(res => {
          this.subjob_options = res.data;
          this.edit_dig_title = "新增任务";
          this.edit_dig_visible = true;
        });
      }
    },
    emptyForm() {
      return {
        name: "", // 任务名称
        cron: "", // cron 表达式
        protocol: "http", // 网络协议 http / https
        uri: "", // 任务的资源标识符
        remark: "", // 备注
        status: 1, // 状态 0暂停 1正常
        creator: "", // 创建人
        preJobId: "", // 前置任务ID
        timeout: 60, // 任务超时时间
        retryCount: 0, // 重试次数
        retryWaitTime: "", // 重试间隔（秒）
        failTakeover: "1", // 故障转移 0不转移 1转移
        misfireThreshold: 0, // 触发器超时时间（秒）
        executorSelectStrategy: "", // 执行器选择策略 随机 全部 分片
        httpParam: "", // http参数
        httpHeaderParam: "", // http头参数
        httpSign: "0", // 数字签名
        shardingCount: 0, // 分片总数
        shardingParam: "", // 分片参数
        alarmEmail: "", // 告警邮箱
        subJobScheduleStrategy: "0", // 子JOB触发策略 0执行完毕触发 1执行成功触发 2执行失败触发
        subJobIds: [], // 子JOB
        executors: [
          {
            address: "",
            weight: 0,
            status: "1"
          }
        ] // 执行节点
      };
    },
    validRules() {
      var validateCron = (rule, value, callback) => {
        if (value === "") {
          callback(new Error("请输入Cron表达式"));
        } else {
          jobApi.validateCron(value).then(res => {
            let ok = res.data;
            if (ok) {
              callback();
            } else {
              callback(new Error("请输正确的Cron表达式"));
            }
          });
        }
      };
      return {
        name: [{ required: true, message: "请输入任务名称", trigger: "blur" }],
        cron: [{ validator: validateCron, required: true, trigger: "blur" }],
        protocol: [
          { required: true, message: "请选择通信协议", trigger: "blur" }
        ],
        uri: [
          { required: true, message: "请输入任务执行方法", trigger: "blur" }
        ],
        executorSelectStrategy: [
          { required: true, message: "请输入执行节点选择策略", trigger: "blur" }
        ],
        failTakeover: [
          { required: true, message: "请选择失败处理策略", trigger: "blur" }
        ],
        timeout: [
          { required: true, message: "请输入执行超时时间", trigger: "blur" }
        ]
      };
    },
    handleAddExecutor() {
      this.form.executors.push({
        address: "",
        weight: 0,
        status: "1"
      });
    },
    handleDelExecutor(index) {
      this.form.executors.splice(index, 1);
    },
    handleHelp() {
      this.$refs.job_edit_help.initPage();
    },
    handleUserMailSelectorDig() {
      this.$refs.user_mail_selector.initPage(this.form.alarmEmail);
    },
    handleUserMailSelected(val) {
      this.form.alarmEmail = val
    },
    // 保存编辑
    saveEdit() {
      this.$refs.job_edit_form.validate(valid => {
        if (valid) {
          if (this.form.executorSelectStrategy != "sharding") {
            this.form.shardingCount = 0;
            this.form.shardingParam = "";
          }
          if (this.form.httpParam) {
            let p = this.form.httpParam;
            let ps = p.split("|");
            for (let i in ps) {
              let pss = ps[i].split("=");
              if (pss.length != 2) {
                this.$message.error(
                  "附加参数格式错误，请使用 'key=value' 键值对的格式，多个键值对用|分隔"
                );
                return;
              }
            }
          }
          if (this.form.httpHeaderParam) {
            let p = this.form.httpHeaderParam;
            let ps = p.split("|");
            for (let i in ps) {
              let pss = ps[i].split("=");
              if (pss.length != 2) {
                this.$message.error(
                  "Header参数格式错误，请使用 'key=value' 键值对的格式，多个键值对用|分隔"
                );
                return;
              }
            }
          }
          if (this.form.shardingCount > 0) {
            if (this.form.shardingParam) {
              let p = this.form.shardingParam;
              let ps = p.split("|");
              if (ps.length != this.form.shardingCount) {
                this.$message.error(
                  "分片参数错误，使用|分隔并且保证数量和分片总数一致"
                );
                return;
              }
            }
          }
          if (this.form.alarmEmail) {
            let p = this.form.alarmEmail;
            let ps = p.split("|");
            for (let i in ps) {
              var reg = new RegExp(/^\S+@\S+\.\S{2,}$/);
              if (!reg.test(ps[i])) {
                this.$message.error("告警邮件中存在不合规的邮件地址，请更正");
                return;
              }
            }
          }

          for (let index in this.form.executors) {
            let node = this.form.executors[index];
            var reg = new RegExp(
              /^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\:([0-9]|[1-9]\d{1,3}|[1-5]\d{4}|6[0-5]{2}[0-3][0-5])$/
            );
            if (!reg.test(node.address)) {
              this.$message.error(
                "执行节点列表，第" +
                  (parseInt(index) + 1) +
                  "行的地址格式不正确，请使用  'IP:端口'  的格式"
              );
              return;
            }
          }
          for (let i in this.form.executors) {
            let node = this.form.executors[i];
            for (let j in this.form.executors) {
              if (i != j) {
                let node2 = this.form.executors[j];
                if (node.address == node2.address) {
                  this.$message.error(
                    "执行节点列表，存在与第" +
                      (parseInt(i) + 1) +
                      "行 相同的节点地址，请更换"
                  );
                  return;
                }
              }
            }
          }
          for (let index in this.form.executors) {
            let node = this.form.executors[index];
            node.status = parseInt(node.status);
          }

          this.form.failTakeover = parseInt(this.form.failTakeover);
          this.form.httpSign = parseInt(this.form.httpSign);
          this.form.subJobScheduleStrategy = parseInt(
            this.form.subJobScheduleStrategy
          );
          if (this.form.id) {
            jobApi.putJob(this.form).then(res => {
              this.edit_dig_visible = false;
              this.$emit("refreshList");
              this.$message.success(`编辑成功`);
            });
          } else {
            this.form.creator = this.$store.getters.getUserName;
            jobApi.postJob(this.form).then(res => {
              this.edit_dig_visible = false;
              this.$emit("refreshList");
              this.$message.success(`新增成功`);
            });
          }
        } else {
          console.log("error submit!!");
          return;
        }
      });
    },
    handleEditDigClose() {
      this.form = this.emptyForm();
      this.$refs.job_edit_form.resetFields();
    }
  }
};
</script>