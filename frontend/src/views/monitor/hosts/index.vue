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
                        <el-tag v-for="tag in row.tags ? row.tags.split(',') : []" :key="tag" size="small" class="mr-1">
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

            <!-- 热力图 -->
            <div class="heatmap-container">
                <div class="heatmap-header">
                    <h3>节点在线状态热力图</h3>
                    <div class="heatmap-controls">
                        <el-radio-group v-model="selectedHours" size="small" @change="onHoursChange">
                            <el-radio-button :value="6">6小时</el-radio-button>
                            <el-radio-button :value="12">12小时</el-radio-button>
                            <el-radio-button :value="24">24小时</el-radio-button>
                            <el-radio-button :value="72">3天</el-radio-button>
                            <el-radio-button :value="168">7天</el-radio-button>
                        </el-radio-group>
                        <el-button @click="fetchHeatmap" :loading="heatmapLoading" size="small" class="ml-4">
                            <el-icon><Refresh /></el-icon>
                            刷新
                        </el-button>
                    </div>
                </div>
                <div ref="heatmapChart" style="width: 100%; height: 400px"></div>
            </div>
        </template>
    </LayoutContent>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { Refresh } from '@element-plus/icons-vue';
import { useRouter } from 'vue-router';
import { listNodes, getNodeHeatmap, type NodeInfo, type NodeHeatmapData } from '@/api/modules/node';
import dayjs from 'dayjs';
import * as echarts from 'echarts';

const MAX_TIME_SLOTS = 300;

const router = useRouter();
const loading = ref(false);
const heatmapLoading = ref(false);
const nodeList = ref<NodeInfo[]>([]);
const heatmapChart = ref<HTMLElement | null>(null);
let chartInstance: echarts.ECharts | null = null;

const selectedHours = ref(24);

// 根据选定的小时数计算合理的 stepMinutes，确保数据点不超过 MAX_TIME_SLOTS
const calcStepMinutes = (hours: number): number => {
    let step = 5; // 默认5分钟
    const slots = (hours * 60) / step;
    if (slots > MAX_TIME_SLOTS) {
        step = Math.ceil((hours * 60) / MAX_TIME_SLOTS);
        // 向上取整到5的倍数，更友好
        step = Math.ceil(step / 5) * 5;
    }
    return step;
};

const fetchNodes = async () => {
    loading.value = true;
    try {
        const res = await listNodes();
        nodeList.value = res.data ?? [];
    } finally {
        loading.value = false;
    }
};

const fetchHeatmap = async () => {
    heatmapLoading.value = true;
    const stepMinutes = calcStepMinutes(selectedHours.value);
    try {
        const res = await getNodeHeatmap({ hours: selectedHours.value, stepMinutes });
        const data: NodeHeatmapData = res.data!;
        renderHeatmap(data);
    } catch (err) {
        console.error('Failed to fetch heatmap data:', err);
    } finally {
        heatmapLoading.value = false;
    }
};

const onHoursChange = () => {
    fetchHeatmap();
};

const renderHeatmap = (data: NodeHeatmapData) => {
    if (!heatmapChart.value) return;

    if (!chartInstance) {
        chartInstance = echarts.init(heatmapChart.value);
    }

    // 将 values[nodeIndex][timeIndex] 转换为热力图数据格式 [timeIndex, nodeIndex, value]
    const heatmapData: number[][] = [];
    for (let i = 0; i < data.values.length; i++) {
        for (let j = 0; j < data.values[i].length; j++) {
            if (data.values[i][j] >= 0) {
                heatmapData.push([j, i, data.values[i][j]]);
            }
        }
    }

    // 动态计算x轴label间隔，避免标签过密
    const labelInterval = Math.max(Math.floor(data.times.length / 20), 0);

    const option: echarts.EChartsOption = {
        tooltip: {
            position: 'top',
            formatter: (params: any) => {
                const time = data.times[params.data[0]];
                const node = data.nodes[params.data[1]];
                const status = params.data[2] === 1 ? '在线' : '离线';
                return `${node}<br/>${time}<br/>状态: ${status}`;
            },
        },
        grid: {
            left: '10%',
            right: '5%',
            top: '10%',
            bottom: '15%',
        },
        xAxis: {
            type: 'category',
            data: data.times,
            splitArea: { show: true },
            axisLabel: {
                rotate: 45,
                fontSize: 10,
                interval: labelInterval,
            },
        },
        yAxis: {
            type: 'category',
            data: data.nodes,
            splitArea: { show: true },
        },
        visualMap: {
            min: 0,
            max: 1,
            calculable: true,
            orient: 'vertical',
            left: 'right',
            top: 'center',
            inRange: {
                color: ['#FF6B6B', '#51CF66'], // 红色=离线，绿色=在线
            },
            textStyle: {
                color: '#000',
            },
            formatter: (value: number) => (value === 1 ? '在线' : '离线'),
        },
        series: [
            {
                name: '在线状态',
                type: 'heatmap',
                data: heatmapData,
                label: { show: false },
                emphasis: {
                    itemStyle: {
                        shadowBlur: 10,
                        shadowColor: 'rgba(0, 0, 0, 0.5)',
                    },
                },
            },
        ],
    };

    chartInstance.setOption(option, true);
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
    router.push({ name: 'MonitorDashboard', query: { nodeId: String(row.id) } });
};

const handleResize = () => {
    chartInstance?.resize();
};

onMounted(() => {
    fetchNodes();
    fetchHeatmap();
    window.addEventListener('resize', handleResize);
});

onUnmounted(() => {
    window.removeEventListener('resize', handleResize);
    chartInstance?.dispose();
});
</script>

<style scoped>
.heatmap-container {
    margin-top: 20px;
    padding: 20px;
    background: #fff;
    border-radius: 4px;
    border: 1px solid #e4e7ed;
}

.heatmap-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 15px;
}

.heatmap-header h3 {
    margin-bottom: 0;
    color: #303133;
    font-size: 16px;
}

.heatmap-controls {
    display: flex;
    align-items: center;
}

.ml-4 {
    margin-left: 16px;
}
</style>
