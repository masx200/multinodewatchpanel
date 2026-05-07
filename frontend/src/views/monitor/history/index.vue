<template>
    <LayoutContent :title="$t('monitor.historyMonitor')" :divider="true">
        <template #toolbar>
            <el-select
                v-model="query.nodeId"
                :placeholder="$t('monitor.selectNode')"
                style="width: 160px"
                @change="onSearch"
            >
                <el-option
                    v-for="node in nodeList"
                    :key="node.id"
                    :label="node.name"
                    :value="node.id"
                />
            </el-select>
            <el-select v-model="query.metricType" style="width: 140px; margin-left: 8px" @change="onSearch">
                <el-option label="CPU (%)" value="cpu" />
                <el-option :label="$t('monitor.memory') + ' (%)'" value="memory" />
                <el-option :label="$t('monitor.disk') + ' (%)'" value="disk" />
                <el-option :label="$t('monitor.netUpload') + ' (KB/s)'" value="net_upload" />
                <el-option :label="$t('monitor.netDownload') + ' (KB/s)'" value="net_download" />
                <el-option :label="$t('monitor.ioRead') + ' (KB/s)'" value="io_read" />
                <el-option :label="$t('monitor.ioWrite') + ' (KB/s)'" value="io_write" />
            </el-select>
            <el-date-picker
                v-model="timeRange"
                type="datetimerange"
                style="width: 360px; margin-left: 8px"
                :range-separator="$t('commons.table.to')"
                :start-placeholder="$t('commons.search.startTime')"
                :end-placeholder="$t('commons.search.endTime')"
                format="YYYY-MM-DD HH:mm:ss"
                value-format="YYYY-MM-DD HH:mm:ss"
                @change="onSearch"
            />
        </template>

        <template #main>
            <div v-if="!query.nodeId" class="empty-tip">
                <el-empty :description="$t('monitor.noNodeSelected')" />
            </div>
            <div v-else v-loading="loading">
                <!-- 折线图 -->
                <div
                    v-if="historyData.length > 0"
                    ref="chartRef"
                    style="width: 100%; height: 360px"
                ></div>
                <el-empty v-else :description="$t('monitor.noDataInRange')" />

                <div v-if="historyData.length > 0" class="mt-2 text-right text-sm" style="color: #909399">
                    {{ $t('monitor.totalRecords', { count: historyData.length }) }}
                    &nbsp;&nbsp;
                    {{ $t('monitor.metricLabel') }}：{{ metricLabel(query.metricType) }}
                </div>
            </div>
        </template>
    </LayoutContent>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted, onUnmounted, watch, nextTick } from 'vue';
import { listNodes, searchMonitorHistory, type NodeInfo, type MonitorHistoryItem } from '@/api/modules/node';
import { useI18n } from 'vue-i18n';
import dayjs from 'dayjs';
import * as echarts from 'echarts/core';
import { LineChart } from 'echarts/charts';
import {
    TitleComponent,
    TooltipComponent,
    GridComponent,
    DataZoomComponent,
    LegendComponent,
    ToolboxComponent,
} from 'echarts/components';
import { CanvasRenderer } from 'echarts/renderers';

echarts.use([
    LineChart,
    TitleComponent,
    TooltipComponent,
    GridComponent,
    DataZoomComponent,
    LegendComponent,
    ToolboxComponent,
    CanvasRenderer,
]);

const { t } = useI18n();
const loading = ref(false);
const nodeList = ref<NodeInfo[]>([]);
const historyData = ref<MonitorHistoryItem[]>([]);
const chartRef = ref<HTMLDivElement | null>(null);
let chartInstance: echarts.ECharts | null = null;

const timeRange = ref<[string, string]>([
    dayjs().subtract(1, 'hour').format('YYYY-MM-DD HH:mm:ss'),
    dayjs().format('YYYY-MM-DD HH:mm:ss'),
]);

const query = reactive({
    nodeId: 0,
    metricType: 'cpu',
});

const metricLabel = (type: string) => {
    const map: Record<string, string> = {
        cpu: 'CPU (%)',
        memory: t('monitor.memory') + ' (%)',
        disk: t('monitor.disk') + ' (%)',
        net_upload: t('monitor.netUpload') + ' (KB/s)',
        net_download: t('monitor.netDownload') + ' (KB/s)',
        io_read: t('monitor.ioRead') + ' (KB/s)',
        io_write: t('monitor.ioWrite') + ' (KB/s)',
    };
    return map[type] ?? type;
};

const buildChartOption = (data: MonitorHistoryItem[]) => {
    const times = data.map((d) => d.time);
    const values = data.map((d) => +d.value.toFixed(2));
    const isPercent = ['cpu', 'memory', 'disk'].includes(query.metricType);

    return {
        tooltip: {
            trigger: 'axis',
            formatter: (params: any[]) => {
                const p = params[0];
                return `${p.axisValue}<br/>${metricLabel(query.metricType)}：<b>${p.value}</b>`;
            },
        },
        toolbox: {
            feature: {
                dataZoom: { yAxisIndex: 'none' },
                restore: {},
                saveAsImage: {},
            },
        },
        grid: { left: 60, right: 40, top: 40, bottom: 60 },
        dataZoom: [
            { type: 'inside', start: 0, end: 100 },
            { type: 'slider', start: 0, end: 100 },
        ],
        xAxis: {
            type: 'category',
            data: times,
            axisLabel: {
                rotate: 30,
                formatter: (val: string) => val.slice(5),
            },
        },
        yAxis: {
            type: 'value',
            min: 0,
            max: isPercent ? 100 : undefined,
            axisLabel: {
                formatter: (v: number) => (isPercent ? v + '%' : v),
            },
        },
        series: [
            {
                name: metricLabel(query.metricType),
                type: 'line',
                smooth: true,
                symbol: 'none',
                lineStyle: { width: 2 },
                areaStyle: { opacity: 0.12 },
                data: values,
                color: '#409EFF',
            },
        ],
    };
};

const renderChart = async () => {
    await nextTick();
    if (!chartRef.value) return;
    if (!chartInstance) {
        chartInstance = echarts.init(chartRef.value);
    }
    chartInstance.setOption(buildChartOption(historyData.value), true);
};

const fetchNodes = async () => {
    const res = await listNodes();
    nodeList.value = res.data ?? [];
    if (nodeList.value.length > 0) {
        const online = nodeList.value.find((n) => n.status === 1);
        query.nodeId = online?.id ?? nodeList.value[0].id;
        await onSearch();
    }
};

const onSearch = async () => {
    if (!query.nodeId) return;
    loading.value = true;
    try {
        const [start, end] = timeRange.value ?? [
            dayjs().subtract(1, 'hour').format('YYYY-MM-DD HH:mm:ss'),
            dayjs().format('YYYY-MM-DD HH:mm:ss'),
        ];
        const res = await searchMonitorHistory({
            nodeId: query.nodeId,
            metricType: query.metricType,
            startTime: start,
            endTime: end,
        });
        historyData.value = res.data ?? [];
        if (historyData.value.length > 0) {
            await renderChart();
        }
    } finally {
        loading.value = false;
    }
};

// 图表随容器宽度变化自适应
const handleResize = () => chartInstance?.resize();

watch(
    () => historyData.value,
    async (val) => {
        if (val.length > 0) await renderChart();
    },
);

onMounted(() => {
    fetchNodes();
    window.addEventListener('resize', handleResize);
});

onUnmounted(() => {
    window.removeEventListener('resize', handleResize);
    chartInstance?.dispose();
    chartInstance = null;
});
</script>

<style scoped>
.empty-tip {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 300px;
}
</style>
