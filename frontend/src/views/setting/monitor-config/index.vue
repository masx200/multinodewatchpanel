<template>
    <LayoutContent :title="$t('monitor.monitorConfig')" :divider="true">
        <template #main>
            <el-form
                ref="formRef"
                :model="form"
                label-width="180px"
                style="max-width: 600px"
                v-loading="loading"
            >
                <el-form-item :label="$t('monitor.dataRetentionDays')" prop="dataRetentionDays">
                    <el-input-number
                        v-model="form.dataRetentionDays"
                        :min="1"
                        :max="365"
                        style="width: 180px"
                    />
                    <span class="ml-2 form-hint">{{ $t('monitor.dataRetentionDaysHint') }}</span>
                </el-form-item>

                <el-form-item :label="$t('monitor.collectInterval')" prop="collectInterval">
                    <el-input-number
                        v-model="form.collectInterval"
                        :min="10"
                        :max="43200"
                        :step="10"
                        style="width: 180px"
                    />
                    <span class="ml-2 form-hint">{{ $t('monitor.collectIntervalHint') }}</span>
                </el-form-item>

                <el-form-item :label="$t('monitor.dashboardRefresh')" prop="dashboardRefresh">
                    <el-input-number
                        v-model="form.dashboardRefresh"
                        :min="5"
                        :max="300"
                        :step="5"
                        style="width: 180px"
                    />
                    <span class="ml-2 form-hint">{{ $t('monitor.dashboardRefreshHint') }}</span>
                </el-form-item>

                <el-form-item>
                    <el-button type="primary" :loading="saving" @click="onSave">
                        {{ $t('commons.button.save') }}
                    </el-button>
                    <el-button @click="onReset">{{ $t('commons.button.reset') }}</el-button>
                </el-form-item>
            </el-form>
        </template>
    </LayoutContent>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import { getSettingBy, updateSetting } from '@/api/modules/setting';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const loading = ref(false);
const saving = ref(false);

const defaultForm = {
    dataRetentionDays: 30,
    collectInterval: 60,
    dashboardRefresh: 10,
};

const form = reactive({ ...defaultForm });

const fetchSettings = async () => {
    loading.value = true;
    try {
        const [r1, r2, r3] = await Promise.all([
            getSettingBy('DataRetentionDays'),
            getSettingBy('CollectInterval'),
            getSettingBy('DashboardRefresh'),
        ]);
        form.dataRetentionDays = parseInt(r1.data as unknown as string) || defaultForm.dataRetentionDays;
        form.collectInterval = parseInt(r2.data as unknown as string) || defaultForm.collectInterval;
        form.dashboardRefresh = parseInt(r3.data as unknown as string) || defaultForm.dashboardRefresh;
    } finally {
        loading.value = false;
    }
};

const onSave = async () => {
    saving.value = true;
    try {
        await Promise.all([
            updateSetting({ key: 'DataRetentionDays', value: String(form.dataRetentionDays) }),
            updateSetting({ key: 'CollectInterval', value: String(form.collectInterval) }),
            updateSetting({ key: 'DashboardRefresh', value: String(form.dashboardRefresh) }),
        ]);
        ElMessage.success(t('commons.msg.updateSuccess'));
    } finally {
        saving.value = false;
    }
};

const onReset = () => {
    Object.assign(form, defaultForm);
};

onMounted(fetchSettings);
</script>

<style scoped>
.form-hint {
    color: #909399;
    font-size: 12px;
}
</style>
