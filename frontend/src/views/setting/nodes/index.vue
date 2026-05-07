<template>
    <LayoutContent :title="$t('monitor.nodeManage')" :divider="true">
        <template #toolbar>
            <el-button type="primary" @click="openDialog('create')">
                <el-icon><Plus /></el-icon>
                {{ $t('commons.button.add') }}
            </el-button>
            <el-button @click="refreshStatus">
                <el-icon><Refresh /></el-icon>
                {{ $t('monitor.refreshStatus') }}
            </el-button>
        </template>

        <template #main>
            <el-table :data="nodeList" v-loading="loading" border stripe>
                <el-table-column prop="name" :label="$t('monitor.nodeName')" min-width="120" />
                <el-table-column prop="host" :label="$t('monitor.nodeHost')" min-width="180" />
                <el-table-column prop="port" :label="$t('monitor.nodePort')" width="80" />
                <el-table-column prop="security" :label="$t('monitor.nodeSecurity')" width="80">
                    <template #default="{ row }">
                        <el-tag :type="row.security === 'tls' ? 'success' : 'info'" size="small">
                            {{ row.security === 'tls' ? 'TLS' : 'None' }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="tags" :label="$t('monitor.nodeTags')" min-width="120">
                    <template #default="{ row }">
                        <el-tag
                            v-for="tag in row.tags ? row.tags.split(',') : []"
                            :key="tag"
                            size="small"
                            class="mr-1"
                        >
                            {{ tag }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="status" :label="$t('monitor.nodeStatus')" width="100">
                    <template #default="{ row }">
                        <el-tag :type="statusType(row.status)" size="small">
                            {{ statusLabel(row.status) }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="lastSeen" :label="$t('monitor.lastSeen')" min-width="160">
                    <template #default="{ row }">
                        {{ row.lastSeen ? formatDate(row.lastSeen) : '-' }}
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.operate')" width="160" fixed="right">
                    <template #default="{ row }">
                        <el-button link type="primary" @click="openDialog('edit', row)">
                            {{ $t('commons.button.edit') }}
                        </el-button>
                        <el-button link type="danger" @click="onDelete(row)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </template>
                </el-table-column>
            </el-table>
        </template>

        <!-- 节点编辑对话框 -->
        <el-dialog
            v-model="dialogVisible"
            :title="dialogTitle"
            width="600px"
            :close-on-click-modal="false"
        >
            <el-form
                ref="formRef"
                :model="nodeForm"
                :rules="rules"
                label-width="160px"
            >
                <el-form-item :label="$t('monitor.nodeName')" prop="name">
                    <el-input v-model="nodeForm.name" :placeholder="$t('monitor.nodeNamePlaceholder')" />
                </el-form-item>
                <el-form-item :label="$t('monitor.nodeHost')" prop="host">
                    <el-input v-model="nodeForm.host" placeholder="192.168.1.100" />
                </el-form-item>
                <el-form-item :label="$t('monitor.nodePort')" prop="port">
                    <el-input-number v-model="nodeForm.port" :min="1" :max="65535" style="width: 100%" />
                </el-form-item>
                <el-form-item :label="$t('monitor.nodeApiKey')" prop="apiKey">
                    <el-input
                        v-model="nodeForm.apiKey"
                        type="password"
                        show-password
                        :placeholder="dialogMode === 'edit' ? $t('monitor.apiKeyEditHint') : ''"
                    />
                </el-form-item>
                <el-form-item :label="$t('monitor.nodeSecurity')" prop="security">
                    <el-radio-group v-model="nodeForm.security">
                        <el-radio value="none">None</el-radio>
                        <el-radio value="tls">TLS</el-radio>
                    </el-radio-group>
                </el-form-item>
                <template v-if="nodeForm.security === 'tls'">
                    <el-form-item :label="$t('monitor.nodeServerName')">
                        <el-input
                            v-model="nodeForm.serverName"
                            :placeholder="$t('monitor.nodeServerNamePlaceholder')"
                        />
                    </el-form-item>
                    <el-form-item :label="$t('monitor.nodeAllowInsecure')">
                        <el-switch v-model="nodeForm.allowInsecure" />
                        <span class="ml-2 form-hint">{{ $t('monitor.nodeAllowInsecureHint') }}</span>
                    </el-form-item>
                    <el-form-item :label="$t('monitor.nodePinnedCert')">
                        <el-input
                            v-model="nodeForm.pinnedPeerCertSha256"
                            :placeholder="$t('monitor.nodePinnedCertPlaceholder')"
                            clearable
                        />
                    </el-form-item>
                </template>
                <el-form-item :label="$t('monitor.nodeTags')">
                    <el-input v-model="nodeForm.tags" :placeholder="$t('monitor.nodeTagsPlaceholder')" />
                </el-form-item>
            </el-form>

            <template #footer>
                <el-button @click="testConnection" :loading="testLoading">
                    {{ $t('monitor.testConnection') }}
                </el-button>
                <el-button @click="dialogVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submitForm" :loading="btnLoading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </template>
        </el-dialog>
    </LayoutContent>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted } from 'vue';
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus';
import { Plus, Refresh } from '@element-plus/icons-vue';
import {
    listNodes,
    createNode,
    updateNode,
    deleteNode,
    testNodeConnection,
    type NodeInfo,
} from '@/api/modules/node';
import { useI18n } from 'vue-i18n';
import dayjs from 'dayjs';

const { t } = useI18n();

const loading = ref(false);
const btnLoading = ref(false);
const testLoading = ref(false);
const dialogVisible = ref(false);
const dialogMode = ref<'create' | 'edit'>('create');
const dialogTitle = ref('');
const nodeList = ref<NodeInfo[]>([]);
const editId = ref<number>(0);
const formRef = ref<FormInstance>();

const nodeForm = reactive({
    name: '',
    host: '',
    port: 9999,
    apiKey: '',
    tags: '',
    security: 'none',
    serverName: '',
    allowInsecure: false,
    pinnedPeerCertSha256: '',
});

const rules = reactive<FormRules>({
    name: [{ required: true, message: t('commons.rule.required'), trigger: 'blur' }],
    host: [{ required: true, message: t('commons.rule.required'), trigger: 'blur' }],
    port: [{ required: true, message: t('commons.rule.required'), trigger: 'blur' }],
    apiKey: [
        {
            required: true,
            validator: (_, value, cb) => {
                if (dialogMode.value === 'create' && !value) {
                    cb(new Error(t('commons.rule.required')));
                } else {
                    cb();
                }
            },
            trigger: 'blur',
        },
    ],
});

const fetchNodes = async () => {
    loading.value = true;
    try {
        const res = await listNodes();
        nodeList.value = res.data ?? [];
    } finally {
        loading.value = false;
    }
};

const openDialog = (mode: 'create' | 'edit', row?: NodeInfo) => {
    dialogMode.value = mode;
    dialogTitle.value = mode === 'create' ? t('monitor.addNode') : t('monitor.editNode');
    if (mode === 'create') {
        nodeForm.name = '';
        nodeForm.host = '';
        nodeForm.port = 9999;
        nodeForm.apiKey = '';
        nodeForm.tags = '';
        nodeForm.security = 'none';
        nodeForm.serverName = '';
        nodeForm.allowInsecure = false;
        nodeForm.pinnedPeerCertSha256 = '';
    } else if (row) {
        editId.value = row.id;
        nodeForm.name = row.name;
        nodeForm.host = row.host;
        nodeForm.port = row.port;
        nodeForm.apiKey = '';
        nodeForm.tags = row.tags;
        nodeForm.security = row.security || 'none';
        nodeForm.serverName = row.serverName || '';
        nodeForm.allowInsecure = row.allowInsecure || false;
        nodeForm.pinnedPeerCertSha256 = row.pinnedPeerCertSha256 || '';
    }
    dialogVisible.value = true;
};

const testConnection = async () => {
    testLoading.value = true;
    try {
        await testNodeConnection({
            name: nodeForm.name || 'test',
            host: nodeForm.host,
            port: nodeForm.port,
            apiKey: nodeForm.apiKey,
            security: nodeForm.security,
            serverName: nodeForm.serverName,
            allowInsecure: nodeForm.allowInsecure,
            pinnedPeerCertSha256: nodeForm.pinnedPeerCertSha256,
        });
        ElMessage.success(t('monitor.connectionSuccess'));
    } catch (e: any) {
        ElMessage.error(t('monitor.connectionFailed') + ': ' + (e?.message ?? e));
    } finally {
        testLoading.value = false;
    }
};

const submitForm = async () => {
    if (!formRef.value) return;
    await formRef.value.validate(async (valid) => {
        if (!valid) return;
        btnLoading.value = true;
        try {
            const payload = {
                name: nodeForm.name,
                host: nodeForm.host,
                port: nodeForm.port,
                apiKey: nodeForm.apiKey,
                tags: nodeForm.tags,
                security: nodeForm.security,
                serverName: nodeForm.serverName,
                allowInsecure: nodeForm.allowInsecure,
                pinnedPeerCertSha256: nodeForm.pinnedPeerCertSha256,
            };
            if (dialogMode.value === 'create') {
                await createNode(payload);
                ElMessage.success(t('commons.msg.createSuccess'));
            } else {
                await updateNode(editId.value, payload);
                ElMessage.success(t('commons.msg.updateSuccess'));
            }
            dialogVisible.value = false;
            await fetchNodes();
        } finally {
            btnLoading.value = false;
        }
    });
};

const onDelete = async (row: NodeInfo) => {
    await ElMessageBox.confirm(t('commons.msg.delete'), t('commons.button.delete'), {
        confirmButtonText: t('commons.button.confirm'),
        cancelButtonText: t('commons.button.cancel'),
        type: 'warning',
    });
    await deleteNode(row.id);
    ElMessage.success(t('commons.msg.deleteSuccess'));
    await fetchNodes();
};

const refreshStatus = async () => {
    await fetchNodes();
};

const statusType = (status: number) => {
    if (status === 1) return 'success';
    if (status === 2) return 'danger';
    return 'info';
};

const statusLabel = (status: number) => {
    if (status === 1) return t('monitor.online');
    if (status === 2) return t('monitor.failed');
    return t('monitor.offline');
};

const formatDate = (d: string) => dayjs(d).format('YYYY-MM-DD HH:mm:ss');

onMounted(fetchNodes);
</script>

<style scoped>
.form-hint {
    color: #909399;
    font-size: 12px;
}
</style>
