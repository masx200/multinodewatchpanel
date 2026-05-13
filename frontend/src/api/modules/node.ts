import http from '@/api';

export interface NodeInfo {
    id: number;
    name: string;
    host: string;
    port: number;
    status: number; // 0=offline 1=online 2=failed
    lastSeen?: string;
    tags: string;
    isDefault: boolean;
    serverName: string;
    security: string; // "none" | "tls"
    allowInsecure: boolean;
    pinnedPeerCertSha256: string;
    createdAt: string;
    updatedAt: string;
}

export interface NodeCreate {
    name: string;
    host: string;
    port: number;
    apiKey: string;
    tags?: string;
    serverName?: string;
    security?: string;
    allowInsecure?: boolean;
    pinnedPeerCertSha256?: string;
    certPem?: string;
}

export interface NodeUpdate {
    name: string;
    host: string;
    port: number;
    apiKey?: string;
    tags?: string;
    serverName?: string;
    security?: string;
    allowInsecure?: boolean;
    pinnedPeerCertSha256?: string;
    certPem?: string;
}

export interface NodeDashboard {
    nodeId: number;
    nodeName: string;
    status: number;
    cpu: number;
    memory: number;
    disk: number;
    netUp: number;
    netDown: number;
    ioRead: number;
    ioWrite: number;
    load1: number;
    load5: number;
    load15: number;
    hostname: string;
    osVersion: string;
    uptime: number;
    memoryTotal: number;
    memoryUsed: number;
    memoryAvail: number;
    swapTotal: number;
    swapUsed: number;
    diskTotal: number;
    diskUsed: number;
    ipAddress: string;
}

export interface MonitorSearchReq {
    nodeId: number;
    metricType: string;
    startTime: string;
    endTime: string;
}

export interface MonitorHistoryItem {
    time: string;
    value: number;
}

export interface ContainerInfo {
    containerID: string;
    name: string;
    imageID: string;
    imageName: string;
    createTime: string;
    state: string;
    runTime: string;          // 运行时长，如 "Up 2 hours"
    network: string[];         // IP 地址列表
    ports: string[];          // 端口映射

    isFromApp: boolean;
    isFromCompose: boolean;
    appName: string;
    appInstallName: string;
    websites: string[];
    isPinned: boolean;
    description: string;

    // 运行时资源统计
    cpuPercent: number;
    memUsage: number;          // int64 字节
    memLimit: number;
}

export interface PsProcessData {
    PID: number;
    name: string;
    PPID: number;
    username: string;
    status: string;
    startTime: string;
    numThreads: number;
    numConnections: number;
    cpuPercent: string;
    cpuValue: number;
    rss: string;
    rssValue: number;
    vms: string;
    diskRead: string;
    diskWrite: string;
    cmdLine: string;
}

// ---- Node API ----

export const listNodes = () => {
    return http.get<NodeInfo[]>('/core/nodes');
};

export const createNode = (req: NodeCreate) => {
    return http.post('/core/nodes', req);
};

export const updateNode = (id: number, req: NodeUpdate) => {
    return http.put(`/core/nodes/${id}`, req);
};

export const deleteNode = (id: number) => {
    return http.delete(`/core/nodes/${id}`);
};

export interface NodeTest {
    id?: number;
    name: string;
    host: string;
    port: number;
    apiKey?: string;
    tags?: string;
    serverName?: string;
    security?: string;
    allowInsecure?: boolean;
    pinnedPeerCertSha256?: string;
    certPem?: string;
}

export const testNodeConnection = (req: NodeTest) => {
    return http.post('/core/nodes/test', req);
};

export const getNodeDashboard = (id: number) => {
    return http.get<NodeDashboard>(`/core/nodes/${id}/dashboard`);
};

export const getNodeContainers = (id: number) => {
    return http.get<ContainerInfo[]>(`/core/nodes/${id}/containers`);
};

export const getNodeProcesses = (id: number) => {
    return http.get<PsProcessData[]>(`/core/nodes/${id}/processes`);
};

export const searchMonitorHistory = (req: MonitorSearchReq) => {
    return http.post<MonitorHistoryItem[]>('/core/nodes/monitor/search', req);
};

export interface NodeHeatmapData {
    times: string[];
    nodes: string[];
    values: number[][];
}

export interface NodeHeatmapParams {
    hours?: number;
    stepMinutes?: number;
}

export const getNodeHeatmap = (params?: NodeHeatmapParams) => {
    return http.get<NodeHeatmapData>('/core/nodes/heatmap', params);
};
