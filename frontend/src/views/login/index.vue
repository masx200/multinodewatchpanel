<template>
    <div class="login-page">
        <div class="absolute inset-0 bg-cover bg-center bg-no-repeat" :style="backgroundStyle"></div>
        <div class="login-container" id="login-container">
            <div v-if="showLogo" class="login-image-side">
                <img
                    v-show="imgLoaded"
                    :src="loadImage('loginImage')"
                    class="login-image"
                    alt="1panel"
                    @load="onImgLoad"
                    @error="onImgError"
                />
            </div>
            <div class="login-form-side">
                <LoginForm ref="loginRef"></LoginForm>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import LoginForm from './components/login-form.vue';
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { GlobalStore } from '@/store';
import { preloadImage } from '@/utils/browser';
defineOptions({ name: 'Login' });
const globalStore = GlobalStore();
const backgroundOpacity = ref(1);
const defaultLoginImage = new URL('@/assets/images/1panel-login.jpg', import.meta.url).href;
const defaultLoginBgImage = new URL('@/assets/images/1panel-login-bg.jpg', import.meta.url).href;
const loadedLoginImage = ref<string | null>(null);
const loadedBackgroundImage = ref<string | null>(null);
const backgroundStyle = ref<{ backgroundImage?: string; backgroundColor?: string }>({});
const imgLoaded = ref(false);

function onImgLoad() {
    imgLoaded.value = true;
}
const mySafetyCode = defineProps({
    code: {
        type: String,
        default: '',
    },
});

const getStatus = async () => {
    let code = mySafetyCode.code;
    if (code != '') {
        globalStore.entrance = code;
    }
};

const loadImage = (name: string) => {
    const { loginImage, loginBackground, loginBgType } = globalStore.themeConfig;
    if (name === 'loginImage') {
        return loginImage === 'loginImage' && loadedLoginImage.value ? loadedLoginImage.value : defaultLoginImage;
    }
    if (name === 'loginBackground') {
        if (loginBgType === 'image') {
            return loginBackground === 'loginBackground' && loadedBackgroundImage.value
                ? loadedBackgroundImage.value
                : defaultLoginBgImage;
        }
        if (loginBgType === 'color') {
            return loginBackground;
        }
        return defaultLoginBgImage;
    }
    return '';
};

const onImgError = (event: any) => {
    event.target.src = defaultLoginImage;
    imgLoaded.value = true;
};

onMounted(async () => {
    await getStatus();
    const loginImageUrl = `/api/v2/images/loginImage?t=${Date.now()}`;
    const backgroundImageUrl = `/api/v2/images/loginBackground?t=${Date.now()}`;
    loadedLoginImage.value = await preloadImage(loginImageUrl);
    loadedBackgroundImage.value = await preloadImage(backgroundImageUrl);
    if (globalStore.themeConfig.loginBgType === 'color') {
        backgroundStyle.value = {
            backgroundColor: globalStore.themeConfig.loginBackground,
        };
    } else {
        const img = new Image();
        const url = loadImage('loginBackground');
        img.onload = () => {
            backgroundStyle.value = {
                backgroundImage: `url(${url})`,
            };
        };
        img.onerror = () => {
            backgroundStyle.value = {
                backgroundImage: `url(${defaultLoginBgImage})`, // 你定义的默认图
            };
        };
        img.src = url;
    }
});

const FIXED_WIDTH = 1000;
const useWindowSize = () => {
    const width = ref(window.innerWidth);
    const height = ref(window.innerHeight);

    const updateSize = () => {
        width.value = window.innerWidth;
        height.value = window.innerHeight;
    };

    onMounted(() => window.addEventListener('resize', updateSize));
    onUnmounted(() => window.removeEventListener('resize', updateSize));

    return { width, height };
};
const { width } = useWindowSize();
const showLogo = computed(() => width.value >= FIXED_WIDTH);
</script>

<style scoped lang="scss">
.login-page {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    position: relative;
    background-color: #f3f4f6;
}

.login-container {
    width: 1000px;
    max-width: 100vw;
    display: flex;
    background: #fff;
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
    position: relative;
    z-index: 10;
    border: 1px solid #e5e7eb;
    overflow: hidden;
    opacity: v-bind(backgroundOpacity);

    @media (max-width: 999px) {
        width: 100%;
    }
}

.login-image-side {
    width: 50%;
    min-height: 420px;
    overflow: hidden;
    flex-shrink: 0;
    display: flex;
    align-items: stretch;

    @media (max-width: 999px) {
        display: none;
    }
}

.login-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.login-form-side {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1.5rem;
    min-width: 0;
}
</style>
