import { Layout } from '@/routers/constant';

const monitorRouter = {
    sort: 2,
    path: '/monitor',
    name: 'Monitor-Menu',
    component: Layout,
    redirect: '/monitor/dashboard',
    meta: {
        icon: 'p-monitor-menu',
        title: 'menu.multiMonitor',
    },
    children: [
        {
            path: '/monitor/dashboard',
            name: 'MonitorDashboard',
            component: () => import('@/views/monitor/dashboard/index.vue'),
            meta: {
                icon: 'p-dashboard',
                title: 'menu.monitorDashboard',
                requiresAuth: false,
            },
        },
        {
            path: '/monitor/hosts',
            name: 'MonitorHosts',
            component: () => import('@/views/monitor/hosts/index.vue'),
            meta: {
                icon: 'p-host',
                title: 'menu.monitorHosts',
                requiresAuth: false,
            },
        },
        {
            path: '/monitor/containers',
            name: 'MonitorContainers',
            component: () => import('@/views/monitor/containers/index.vue'),
            meta: {
                icon: 'p-docker',
                title: 'menu.monitorContainers',
                requiresAuth: false,
            },
        },
        {
            path: '/monitor/history',
            name: 'MonitorHistory',
            component: () => import('@/views/monitor/history/index.vue'),
            meta: {
                icon: 'p-system-monitor-menu',
                title: 'menu.monitorHistory',
                requiresAuth: false,
            },
        },
    ],
};

export default monitorRouter;
