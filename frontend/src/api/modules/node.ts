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
    imageName: string;
    state: string;
    status: string;
    cpuPercent: number;
    memUsage: number;
    memLimit: number;
}

// ---- Node API ----

export const listNodes = () => {
    return http.get<NodeInfo[]>('/nodes');
};

export const createNode = (req: NodeCreate) => {
    return http.post('/nodes', req);
};

export const updateNode = (id: number, req: NodeUpdate) => {
    return http.put(`/nodes/${id}`, req);
};

export const deleteNode = (id: number) => {
    return http.delete(`/nodes/${id}`);
};

export const testNodeConnection = (req: NodeCreate) => {
    return http.post('/nodes/test', req);
};

export const getNodeDashboard = (id: number) => {
    return http.get<NodeDashboard>(`/nodes/${id}/dashboard`);
};

export const getNodeContainers = (id: number) => {
    return http.get<ContainerInfo[]>(`/nodes/${id}/containers`);
};

export const searchMonitorHistory = (req: MonitorSearchReq) => {
    return http.post<MonitorHistoryItem[]>('/nodes/monitor/search', req);
};
