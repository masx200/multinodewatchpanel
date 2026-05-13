<template>
    <LayoutContent :title="$t('monitor.processMonitor')" :divider="true">
        <template #toolbar>
            <el-select
                v-model="currentNodeId"
                :placeholder="$t('monitor.selectNode')"
                style="width: 200px"
                @change="onNodeChange"
            >
                <el-option
                    v-for="node in nodeList"
                    :key="node.id"
                    :label="node.name"
                    :value="node.id"
                />
            </el-select>
            <el-button class="ml-2" @click="fetchProcesses" :loading="loading">
                <el-icon><Refresh /></el-icon>
            </el-button>
        </template>

        <template #main>
            <div v-if="!currentNodeId" class="empty-tip">
                <el-empty :description="$t('monitor.noNodeSelected')" />
            </div>
            <el-table v-else :data="filteredList" v-loading="loading" border stripe max-height="calc(100vh - 220px)">
                <el-table-column prop="PID" label="PID" width="90" sortable />
                <el-table-column prop="name" :label="$t('monitor.processName')" min-width="180" show-overflow-tooltip />
                <el-table-column prop="username" :label="$t('monitor.processUser')" width="120" show-overflow-tooltip />
                <el-table-column prop="cpuPercent" label="CPU%" width="90" sortable :sort-method="sortByCpu">
                    <template #default="{ row }">
                        {{ row.cpuValue?.toFixed(1) ?? '0.0' }}%
                    </template>
                </el-table-column>
                <el-table-column prop="rss" :label="$t('monitor.memUsage')" width="110" sortable :sort-method="sortByMem" />
                <el-table-column prop="numThreads" :label="$t('monitor.threads')" width="90" sortable />
                <el-table-column prop="numConnections" :label="$t('monitor.connections')" width="110" sortable />
                <el-table-column prop="status" :label="$t('monitor.processStatus')" width="100">
                    <template #default="{ row }">
                        <el-tag :type="statusTagType(row.status)" size="small">
                            {{ row.status }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="startTime" :label="$t('monitor.startTime')" width="170" />
            </el-table>
        </template>
    </LayoutContent>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
import { Refresh } from '@element-plus/icons-vue';
import { listNodes, getNodeProcesses, type NodeInfo, type PsProcessData } from '@/api/modules/node';

const loading = ref(false);
const nodeList = ref<NodeInfo[]>([]);
const currentNodeId = ref<number | null>(null);
const processes = ref<PsProcessData[]>([]);
let timer: ReturnType<typeof setInterval> | null = null;

const filteredList = computed(() => processes.value);

const statusTagType = (status: string) => {
    if (status === 'sleep') return 'info';
    if (status === 'running') return 'success';
    if (status === 'stop') return 'danger';
    return 'warning';
};

const sortByCpu = (a: PsProcessData, b: PsProcessData) => (a.cpuValue ?? 0) - (b.cpuValue ?? 0);
const sortByMem = (a: PsProcessData, b: PsProcessData) => (a.rssValue ?? 0) - (b.rssValue ?? 0);

const fetchNodes = async () => {
    const res = await listNodes();
    nodeList.value = res.data ?? [];
    if (!currentNodeId.value && nodeList.value.length > 0) {
        const online = nodeList.value.find((n) => n.status === 1);
        currentNodeId.value = online?.id ?? nodeList.value[0].id;
        await fetchProcesses();
    }
};

const onNodeChange = () => {
    processes.value = [];
    fetchProcesses();
};

const fetchProcesses = async () => {
    if (!currentNodeId.value) return;
    loading.value = true;
    try {
        const res = await getNodeProcesses(currentNodeId.value);
        processes.value = res.data ?? [];
    } catch {
        processes.value = [];
    } finally {
        loading.value = false;
    }
};

const startPolling = () => {
    stopPolling();
    timer = setInterval(fetchProcesses, 5000);
};

const stopPolling = () => {
    if (timer) {
        clearInterval(timer);
        timer = null;
    }
};

onMounted(() => {
    fetchNodes();
    startPolling();
});

onBeforeUnmount(stopPolling);
</script>

<style scoped>
.empty-tip {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 300px;
}
</style>
