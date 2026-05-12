<template>
    <LayoutContent :title="$t('monitor.containerMonitor')" :divider="true">
        <template #toolbar>
            <el-select
                v-model="currentNodeId"
                :placeholder="$t('monitor.selectNode')"
                style="width: 200px"
                @change="fetchContainers"
            >
                <el-option
                    v-for="node in nodeList"
                    :key="node.id"
                    :label="node.name"
                    :value="node.id"
                />
            </el-select>
            <el-button class="ml-2" @click="fetchContainers" :loading="loading">
                <el-icon><Refresh /></el-icon>
            </el-button>
        </template>

        <template #main>
            <div v-if="!currentNodeId" class="empty-tip">
                <el-empty :description="$t('monitor.noNodeSelected')" />
            </div>
            <div v-else-if="dockerError" class="docker-error-state">
                <el-icon :size="48" color="var(--el-color-warning)"><WarningFilled /></el-icon>
                <p class="docker-error-title">{{ $t('monitor.dockerNotDetected') }}</p>
                <p class="docker-error-detail">{{ dockerError }}</p>
                <el-button type="primary" @click="fetchContainers" :loading="loading">
                    <el-icon class="mr-1"><Refresh /></el-icon>
                    {{ $t('monitor.retry') }}
                </el-button>
            </div>
            <el-table v-else :data="containers" v-loading="loading" border stripe>
                <el-table-column prop="name" :label="$t('monitor.containerName')" min-width="160" />
                <el-table-column prop="imageName" :label="$t('monitor.imageName')" min-width="200" />
                <el-table-column prop="state" :label="$t('monitor.containerState')" width="100">
                    <template #default="{ row }">
                        <el-tag :type="row.state === 'running' ? 'success' : 'danger'" size="small">
                            {{ row.state }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="runTime" label="运行时长" width="140" />
                <el-table-column prop="network" label="IP地址" width="150">
                    <template #default="{ row }">
                        <span v-if="row.network && row.network.length > 0">
                            {{ row.network.filter(ip => ip).join(', ') }}
                        </span>
                        <span v-else class="text-gray">-</span>
                    </template>
                </el-table-column>
                <el-table-column prop="ports" label="端口映射" min-width="200">
                    <template #default="{ row }">
                        <template v-if="row.ports && row.ports.length > 0">
                            <el-tooltip
                                v-if="row.ports.length > 2"
                                :content="row.ports.join('\n')"
                                placement="top"
                            >
                                <span>{{ row.ports.slice(0, 2).join(', ') }}... (+{{ row.ports.length - 2 }})</span>
                            </el-tooltip>
                            <span v-else>{{ row.ports.join(', ') }}</span>
                        </template>
                        <span v-else class="text-gray">-</span>
                    </template>
                </el-table-column>
                <el-table-column prop="cpuPercent" label="CPU%" width="90">
                    <template #default="{ row }">
                        {{ row.cpuPercent?.toFixed(2) ?? '0.00' }}%
                    </template>
                </el-table-column>
                <el-table-column :label="$t('monitor.memUsage')" width="130">
                    <template #default="{ row }">
                        {{ formatBytes(row.memUsage) }} / {{ formatBytes(row.memLimit) }}
                    </template>
                </el-table-column>
            </el-table>
        </template>
    </LayoutContent>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { Refresh, WarningFilled } from '@element-plus/icons-vue';
import { listNodes, getNodeContainers, type NodeInfo, type ContainerInfo } from '@/api/modules/node';

const loading = ref(false);
const nodeList = ref<NodeInfo[]>([]);
const currentNodeId = ref<number | null>(null);
const containers = ref<ContainerInfo[]>([]);
const dockerError = ref('');

const fetchNodes = async () => {
    const res = await listNodes();
    nodeList.value = res.data ?? [];
    if (!currentNodeId.value && nodeList.value.length > 0) {
        const online = nodeList.value.find((n) => n.status === 1);
        currentNodeId.value = online?.id ?? nodeList.value[0].id;
        await fetchContainers();
    }
};

const fetchContainers = async () => {
    if (!currentNodeId.value) return;
    loading.value = true;
    dockerError.value = '';
    try {
        const res = await getNodeContainers(currentNodeId.value);
        containers.value = res.data ?? [];
    } catch (e: any) {
        const msg = e?.response?.data?.message || e?.message || String(e);
        if (/docker/i.test(msg)) {
            dockerError.value = msg;
            containers.value = [];
        } else {
            throw e;
        }
    } finally {
        loading.value = false;
    }
};

const formatBytes = (bytes: number) => {
    if (!bytes) return '0 B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + ' MB';
    return (bytes / 1024 / 1024 / 1024).toFixed(2) + ' GB';
};

onMounted(fetchNodes);
</script>

<style scoped>
.empty-tip {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 300px;
}
.docker-error-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 300px;
    gap: 12px;
}
.docker-error-title {
    font-size: 16px;
    font-weight: 500;
    color: var(--el-text-color-primary);
    margin: 0;
}
.docker-error-detail {
    font-size: 13px;
    color: var(--el-text-color-secondary);
    margin: 0;
    max-width: 500px;
    text-align: center;
}
.text-gray {
    color: var(--el-text-color-placeholder);
}
</style>
