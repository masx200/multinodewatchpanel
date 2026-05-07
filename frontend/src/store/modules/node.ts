import { defineStore } from 'pinia';
import { listNodes } from '@/api/modules/node';
import type { NodeInfo } from '@/api/modules/node';

interface NodeState {
    nodes: NodeInfo[];
    currentNodeId: number | null;
    loading: boolean;
}

export const useNodeStore = defineStore('NodeStore', {
    state: (): NodeState => ({
        nodes: [],
        currentNodeId: null,
        loading: false,
    }),
    getters: {
        currentNode: (state): NodeInfo | null => {
            if (!state.currentNodeId) return null;
            return state.nodes.find((n) => n.id === state.currentNodeId) ?? null;
        },
        onlineNodes: (state): NodeInfo[] => {
            return state.nodes.filter((n) => n.status === 1);
        },
    },
    actions: {
        async fetchNodes() {
            this.loading = true;
            try {
                const res = await listNodes();
                this.nodes = res.data ?? [];
                // 如果当前没有选中节点，选第一个在线节点
                if (!this.currentNodeId && this.nodes.length > 0) {
                    const online = this.nodes.find((n) => n.status === 1);
                    this.currentNodeId = online?.id ?? this.nodes[0].id;
                }
            } finally {
                this.loading = false;
            }
        },
        setCurrentNode(id: number) {
            this.currentNodeId = id;
        },
    },
});
