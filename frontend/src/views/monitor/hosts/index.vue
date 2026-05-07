<template>
    <LayoutContent :title="$t('monitor.hostList')" :divider="true">
        <template #toolbar>
            <el-button @click="fetchNodes" :loading="loading">
                <el-icon><Refresh /></el-icon>
                {{ $t('monitor.refreshStatus') }}
            </el-button>
        </template>

        <template #main>
            <el-table :data="nodeList" v-loading="loading" border stripe>
                <el-table-column prop="name" :label="$t('monitor.nodeName')" min-width="120" />
                <el-table-column prop="host" :label="$t('monitor.nodeHost')" min-width="180" />
                <el-table-column prop="port" :label="$t('monitor.nodePort')" width="80" />
                <el-table-column prop="status" :label="$t('monitor.nodeStatus')" width="100">
                    <template #default="{ row }">
                        <el-tag :type="statusType(row.status)" size="small">
                            {{ statusLabel(row.status) }}
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
                <el-table-column prop="lastSeen" :label="$t('monitor.lastSeen')" min-width="160">
                    <template #default="{ row }">
                        {{ row.lastSeen ? formatDate(row.lastSeen) : '-' }}
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.operate')" width="120" fixed="right">
                    <template #default="{ row }">
                        <el-button link type="primary" @click="goDashboard(row)">
                            {{ $t('monitor.viewDashboard') }}
                        </el-button>
                    </template>
                </el-table-column>
            </el-table>
        </template>
    </LayoutContent>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { Refresh } from '@element-plus/icons-vue';
import { useRouter } from 'vue-router';
import { listNodes, type NodeInfo } from '@/api/modules/node';
import dayjs from 'dayjs';

const router = useRouter();
const loading = ref(false);
const nodeList = ref<NodeInfo[]>([]);

const fetchNodes = async () => {
    loading.value = true;
    try {
        const res = await listNodes();
        nodeList.value = res.data ?? [];
    } finally {
        loading.value = false;
    }
};

const statusType = (status: number) => {
    if (status === 1) return 'success';
    if (status === 2) return 'danger';
    return 'info';
};

const statusLabel = (status: number) => {
    if (status === 1) return '在线';
    if (status === 2) return '异常';
    return '离线';
};

const formatDate = (d: string) => dayjs(d).format('YYYY-MM-DD HH:mm:ss');

const goDashboard = (row: NodeInfo) => {
    router.push({ name: 'MonitorDashboard', query: { nodeId: row.id } });
};

onMounted(fetchNodes);
</script>
