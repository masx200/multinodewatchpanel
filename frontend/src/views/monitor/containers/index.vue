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
                <el-table-column prop="status" :label="$t('monitor.containerStatus')" min-width="140" />
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
import { Refresh } from '@element-plus/icons-vue';
import { listNodes, getNodeContainers, type NodeInfo, type ContainerInfo } from '@/api/modules/node';

const loading = ref(false);
const nodeList = ref<NodeInfo[]>([]);
const currentNodeId = ref<number | null>(null);
const containers = ref<ContainerInfo[]>([]);

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
    try {
        const res = await getNodeContainers(currentNodeId.value);
        containers.value = res.data ?? [];
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
</style>
