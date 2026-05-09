import { GlobalStore } from '@/store';
import { storeToRefs } from 'pinia';
import type { ComputedRef, Ref } from 'vue';
import type { GlobalState } from '@/store/interface';

type GlobalStoreInstance = ReturnType<typeof GlobalStore>;

type GlobalStoreRefs = {
    [K in keyof GlobalState]: Ref<GlobalState[K]>;
} & {
    isDarkTheme: ComputedRef<boolean>;
    isDarkGoldTheme: ComputedRef<boolean>;
    docsUrl: ComputedRef<string>;
    isMaster: ComputedRef<boolean>;
};

export const useGlobalStore = () => {
    const globalStore = GlobalStore();
    return {
        globalStore,
        ...storeToRefs(globalStore),
    } as { globalStore: GlobalStoreInstance } & GlobalStoreRefs;
};
