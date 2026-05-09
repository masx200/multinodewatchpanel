<template>
    <LayoutContent :title="$t('monitor.dashboard')" :divider="true">
        <template #toolbar>
            <el-select
                v-model="currentNodeId"
                :placeholder="$t('monitor.selectNode')"
                style="width: 400px"
                @change="onNodeChange"
            >
                <el-option
                    v-for="node in nodeList"
                    :key="node.id"
                    :label="node.name"
                    :value="node.id"
                />
            </el-select>
            <el-button class="ml-2" @click="refresh" :loading="loading">
                <el-icon><Refresh /></el-icon>
            </el-button>
        </template>

        <template #main>
            <div v-if="!currentNodeId" class="empty-tip">
                <el-empty :description="$t('monitor.noNodeSelected')" />
            </div>

            <div v-else v-loading="loading">
                <!-- 概览卡片 -->
                <el-row :gutter="16" class="mb-4">
                    <el-col :xs="12" :sm="6">
                        <el-card shadow="hover" class="metric-card">
                            <div class="metric-label">CPU</div>
                            <el-progress
                                type="dashboard"
                                :percentage="dashboard.cpu"
                                :color="progressColor(dashboard.cpu)"
                                :width="160"
                            />
                            <div class="metric-value">{{ dashboard.cpu.toFixed(1) }}%</div>
                        </el-card>
                    </el-col>
                    <el-col :xs="12" :sm="6">
                        <el-card shadow="hover" class="metric-card">
                            <div class="metric-label">{{ $t('monitor.memory') }}</div>
                            <el-progress
                                type="dashboard"
                                :percentage="dashboard.memory"
                                :color="progressColor(dashboard.memory)"
                                :width="160"
                            />
                            <div class="metric-value">{{ dashboard.memory.toFixed(1) }}%</div>
                        </el-card>
                    </el-col>
                    <el-col :xs="12" :sm="6">
                        <el-card shadow="hover" class="metric-card">
                            <div class="metric-label">{{ $t('monitor.disk') }}</div>
                            <el-progress
                                type="dashboard"
                                :percentage="dashboard.disk"
                                :color="progressColor(dashboard.disk)"
                                :width="160"
                            />
                            <div class="metric-value">{{ dashboard.disk.toFixed(1) }}%</div>
                        </el-card>
                    </el-col>
                    <el-col :xs="12" :sm="6">
                        <el-card shadow="hover" class="metric-card">
                            <div class="metric-label">{{ $t('monitor.load') }}</div>
                            <div class="load-values">
                                <span>1m: {{ dashboard.load1.toFixed(2) }}</span>
                                <span>5m: {{ dashboard.load5.toFixed(2) }}</span>
                                <span>15m: {{ dashboard.load15.toFixed(2) }}</span>
                            </div>
                        </el-card>
                    </el-col>
                </el-row>

                <!-- 系统信息 -->
                <el-row :gutter="16" class="mb-4">
                    <el-col :span="24">
                        <el-card shadow="hover">
                            <template #header>
                                <span>{{ $t('monitor.systemInfo') }}</span>
                            </template>
                            <el-descriptions :column="4" border size="large">
                                <el-descriptions-item :label="$t('monitor.hostname')">
                                    {{ dashboard.hostname || '-' }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('monitor.osVersion')">
                                    {{ dashboard.osVersion || '-' }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('monitor.ipAddress')">
                                    {{ dashboard.ipAddress || '-' }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('monitor.uptime')">
                                    {{ formatUptime(dashboard.uptime) }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('monitor.memoryDetail')">
                                    {{ formatBytes(dashboard.memoryUsed) }} / {{ formatBytes(dashboard.memoryTotal) }}
                                    ({{ dashboard.memory.toFixed(1) }}%)
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('monitor.swapDetail')">
                                    {{ dashboard.swapTotal > 0 ? `${formatBytes(dashboard.swapUsed)} / ${formatBytes(dashboard.swapTotal)}` : '-' }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('monitor.diskDetail')">
                                    {{ dashboard.diskTotal > 0 ? `${formatBytes(dashboard.diskUsed)} / ${formatBytes(dashboard.diskTotal)}` : '-' }}
                                    ({{ dashboard.disk.toFixed(1) }}%)
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('monitor.netUpload')">
                                    {{ formatBytes(dashboard.netUp) }}/s ↑
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('monitor.netDownload')">
                                    {{ formatBytes(dashboard.netDown) }}/s ↓
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('monitor.ioRead')">
                                    {{ formatBytes(dashboard.ioRead) }}/s
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('monitor.ioWrite')">
                                    {{ formatBytes(dashboard.ioWrite) }}/s
                                </el-descriptions-item>
                            </el-descriptions>
                        </el-card>
                    </el-col>
                </el-row>
            </div>
        </template>
    </LayoutContent>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { Refresh } from '@element-plus/icons-vue';
import { listNodes, getNodeDashboard, type NodeInfo, type NodeDashboard } from '@/api/modules/node';

const loading = ref(false);
const nodeList = ref<NodeInfo[]>([]);
const currentNodeId = ref<number | null>(null);
const dashboard = ref<NodeDashboard>({
    nodeId: 0,
    nodeName: '',
    status: 0,
    cpu: 0,
    memory: 0,
    disk: 0,
    netUp: 0,
    netDown: 0,
    ioRead: 0,
    ioWrite: 0,
    load1: 0,
    load5: 0,
    load15: 0,
    hostname: '',
    osVersion: '',
    uptime: 0,
    memoryTotal: 0,
    memoryUsed: 0,
    memoryAvail: 0,
    swapTotal: 0,
    swapUsed: 0,
    diskTotal: 0,
    diskUsed: 0,
    ipAddress: '',
});

let timer: ReturnType<typeof setInterval> | null = null;

const fetchNodes = async () => {
    const res = await listNodes();
    nodeList.value = res.data ?? [];
    if (!currentNodeId.value && nodeList.value.length > 0) {
        const online = nodeList.value.find((n) => n.status === 1);
        currentNodeId.value = online?.id ?? nodeList.value[0].id;
        if (currentNodeId.value) await loadDashboard();
    }
};

const loadDashboard = async () => {
    if (!currentNodeId.value) return;
    loading.value = true;
    try {
        const res = await getNodeDashboard(currentNodeId.value);
        dashboard.value = res.data;
    } finally {
        loading.value = false;
    }
};

const onNodeChange = () => loadDashboard();
const refresh = () => loadDashboard();

const progressColor = (val: number) => {
    if (val >= 90) return '#f56c6c';
    if (val >= 70) return '#e6a23c';
    return '#67c23a';
};

const formatUptime = (seconds: number) => {
    if (!seconds) return '-';
    const d = Math.floor(seconds / 86400);
    const h = Math.floor((seconds % 86400) / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    return `${d}d ${h}h ${m}m`;
};

const formatBytes = (bytes: number) => {
    if (!bytes) return '0 B';
    if (bytes < 1024) return bytes.toFixed(0) + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + ' MB';
    return (bytes / 1024 / 1024 / 1024).toFixed(2) + ' GB';
};

onMounted(() => {
    fetchNodes();
    timer = setInterval(loadDashboard, 10000);
});

onUnmounted(() => {
    if (timer) clearInterval(timer);
});
</script>

<style scoped>
.metric-card {
    text-align: center;
    padding: 16px;
}
.metric-label {
    font-size: 28px;
    color: #909399;
    margin-bottom: 16px;
}
.metric-value {
    font-size: 36px;
    font-weight: 600;
    margin-top: 16px;
}
.load-values {
    display: flex;
    flex-direction: column;
    gap: 8px;
    font-size: 28px;
    padding: 32px 0;
}
.empty-tip {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 600px;
}
</style>
