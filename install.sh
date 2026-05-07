#!/bin/bash

# 失败退出
set -e

lang_file=".selected_language" # 语言配置路径
edition_file=".selected_edition" # 发行版本配置路径

# 1Panel 安装主函数
main_install() {
    
    echo "开始安装 1Panel。"
    
    # 获取当前工作目录
    local current_dir=$(pwd)
    
    # 写入发行版本配置文件
    echo "$PANEL_EDITION" > "$edition_file"
    
    # 语言文件配置
    echo "$LANGUAGE" > "$current_dir/$lang_file" # 写入语言配置文件
    # 检查选择的语言是否存在
    local langfile="$current_dir/lang/$LANGUAGE.sh"
    if [ -f "$langfile" ]; then
        source "$langfile" # 导入选定语言的文本变量
    else
        echo "错误：未找到语言文件 $langfile"
        exit 1
    fi
    
    # 设置安装目录
    local use_existing=false # 标记是否使用现有数据
    # 验证安装目录是否为绝对路径
    if [[ "$PANEL_BASE_DIR" != /* ]]; then
        echo "错误：安装目录 $PANEL_BASE_DIR 无效，请提供目录的完整路径。"
        exit 1
    fi
    # 创建安装目录
    if [[ ! -d "$PANEL_BASE_DIR" ]]; then
        mkdir -p "$PANEL_BASE_DIR"
    fi
    # 检查安装目录下是否存在 1Panel 数据库文件
    if [[ -f "$PANEL_BASE_DIR/1panel/db/core.db" ]]; then
        use_existing=true
        echo "检测到现有 1Panel 数据，将使用现有配置。"
    fi
    
    # 设置面板端口
    # 验证端口号是否为数字且在合法范围内 (1-65535)
    if ! [[ "$PANEL_PORT" =~ ^[1-9][0-9]{0,4}$ && "$PANEL_PORT" -le 65535 ]]; then
        echo "错误：面板端口 $PANEL_PORT 无效，输入的端口号必须在 1 到 65535 之间。"
        exit 1
    fi
    
    # 检查宿主机的 docker.sock 是否已挂载到容器内
    if [ ! -S /var/run/docker.sock ]; then
        echo "警告：未检测到 /var/run/docker.sock，请确保已将宿主机的 docker.sock 挂载到容器内。"
    fi
    
    # 检测宿主机的 Docker Daemon 版本
    local docker_version=$(docker --version | grep -oE '[0-9]+\.[0-9]+' | head -n 1)
    local major_version=${docker_version%%.*}
    if [[ $major_version -lt 20 ]]; then
        echo "检测到宿主机 Docker Daemon 版本低于 20.x，建议手动升级以避免功能受限。"
    fi
    
    echo "正在配置 1Panel 服务。"
    
    local run_base_dir=$PANEL_BASE_DIR/1panel # 1Panel 实际运行目录
    # 创建实际运行目录
    mkdir -p "$run_base_dir"
    # 如果不使用现有配置就清理旧的运行目录
    if [[ "$use_existing" == false ]]; then
        rm -rf "$run_base_dir"/* 2>/dev/null
    fi
    
    cd "${current_dir}" || exit # 切换到当前脚本目录
    
    # 复制并设置 1panel-core 可执行文件
    cp ./1panel-core /usr/local/bin && chmod +x /usr/local/bin/1panel-core
    # 移除旧的软链接
    if [[ -e /usr/bin/1panel ]]; then
        rm -f /usr/bin/1panel
    fi
    # 兼容旧版，将 1panel-core 软链接到 1panel
    ln -s /usr/local/bin/1panel-core /usr/bin/1panel >/dev/null 2>&1
    # 确保 /usr/bin/1panel-core 存在软链接
    if [[ ! -f /usr/bin/1panel-core ]]; then
        ln -s /usr/local/bin/1panel-core /usr/bin/1panel-core >/dev/null 2>&1
    fi
    
    # 复制并设置 1panel-agent 可执行文件
    cp ./1panel-agent /usr/local/bin && chmod +x /usr/local/bin/1panel-agent
    # 确保 /usr/bin/1panel-agent 存在软链接
    if [[ ! -f /usr/bin/1panel-agent ]]; then
        ln -s /usr/local/bin/1panel-agent /usr/bin/1panel-agent >/dev/null 2>&1
    fi
    
    # 复制并配置 1pctl 控制工具
    cp ./1pctl /usr/local/bin && chmod +x /usr/local/bin/1pctl
    # 使用 sed 命令替换 1pctl 脚本中的配置变量
    sed -i -e "s#BASE_DIR=.*#BASE_DIR=${PANEL_BASE_DIR}#g" /usr/local/bin/1pctl
    sed -i -e "s#ORIGINAL_PORT=.*#ORIGINAL_PORT=${PANEL_PORT}#g" /usr/local/bin/1pctl
    sed -i -e "s#ORIGINAL_USERNAME=.*#ORIGINAL_USERNAME=${PANEL_USERNAME}#g" /usr/local/bin/1pctl
    # 密码中可能包含特殊字符，需要进行转义
    local escaped_panel_password=$(echo "$PANEL_PASSWORD" | sed 's/[!@#$%*_,.?]/\\&/g')
    sed -i -e "s#ORIGINAL_PASSWORD=.*#ORIGINAL_PASSWORD=${escaped_panel_password}#g" /usr/local/bin/1pctl
    sed -i -e "s#ORIGINAL_ENTRANCE=.*#ORIGINAL_ENTRANCE=${PANEL_ENTRANCE}#g" /usr/local/bin/1pctl
    sed -i -e "s#LANGUAGE=.*#LANGUAGE=${LANGUAGE}#g" /usr/local/bin/1pctl
    sed -i -e "s#^PANEL_EDITION=.*#PANEL_EDITION=${PANEL_EDITION}#g" /usr/local/bin/1pctl
    # 如果是使用现有数据，则在 1pctl 中添加或修改 CHANGE_USER_INFO 标记
    if [[ "$use_existing" == true ]]; then
        if grep -q "^CHANGE_USER_INFO=" "/usr/local/bin/1pctl"; then
            sed -i 's/^CHANGE_USER_INFO=.*/CHANGE_USER_INFO=use_existing/' "/usr/local/bin/1pctl"
        else
            sed -i '/^LANGUAGE=.*/a CHANGE_USER_INFO=use_existing' "/usr/local/bin/1pctl"
        fi
    fi
    # 确保 /usr/bin/1pctl 存在软链接
    if [[ ! -f /usr/bin/1pctl ]]; then
        ln -s /usr/local/bin/1pctl /usr/bin/1pctl >/dev/null 2>&1
    fi
    
    # 复制 GeoIP 数据
    if [ -d "$run_base_dir/geo" ]; then
        rm -rf "$run_base_dir/geo" # 清理旧的 GeoIP 目录
    fi
    mkdir -p "$run_base_dir/geo"
    cp -r ./GeoIP.mmdb "$run_base_dir/geo/"
    
    # 复制语言文件
    cp -r ./lang /usr/local/bin
    
    # 复制传统 sysvinit 的服务脚本到 init.d 满足 1Panel 后端和 1pctl 脚本的自检
    mkdir -p /etc/init.d # docker 容器中这个目录不存在，需要手动创建
    cp ./initscript/1panel-core.init /etc/init.d/1panel-core
    cp ./initscript/1panel-agent.init /etc/init.d/1panel-agent
    chmod +x /etc/init.d/1panel-core
    chmod +x /etc/init.d/1panel-agent
    
    # 复制资源文件
    if [[ ! -d "$run_base_dir/resource" ]]; then
        mkdir -p "$run_base_dir/resource"
    fi
    cp -r ./initscript "$run_base_dir/resource/"
    
    # 清理安装包和解压后的文件
    echo "安装已完成，清理安装文件。"
    cd ..
    rm -f "${package_file_name}"
    rm -rf "${extracted_dir}"
    
    echo
    echo "感谢您的耐心等待，安装已完成。"
    echo "面板端口：$PANEL_PORT"
    echo "安全入口：$PANEL_ENTRANCE"
    echo "面板用户：$PANEL_USERNAME"
    echo "面板密码：$PANEL_PASSWORD"
    echo "请确保 Docker 以 host 网络模式运行或已经配置端口映射 $PANEL_PORT。"
    echo "请确保安装目录 $PANEL_BASE_DIR 已映射到宿主机上的目录以免数据丢失。"
    echo
}

# 检查发行版本
if [[ "${PANEL_EDITION}" == "cn" ]]; then
    INSTALL_MODE="stable"
    elif [[ "${PANEL_EDITION}" == "intl" ]]; then
    INSTALL_MODE="dev"
else
    echo "PANEL_EDITION 参数无效：${PANEL_EDITION}。仅支持：cn、intl。"
    exit 1
fi

# 获取 GitHub 仓库信息（用于从 GitHub Release 下载安装包）
# GITHUB_REPO 默认为 masx200/multinodewatchpanel，可通过环境变量覆盖
GITHUB_REPO="${GITHUB_REPO:-masx200/multinodewatchpanel}"

# 获取当前系统架构
os_check=$(uname -a)
if [[ $os_check =~ 'x86_64' ]]; then
    architecture="amd64"
    elif [[ $os_check =~ 'arm64' ]] || [[ $os_check =~ 'aarch64' ]]; then
    architecture="arm64"
    elif [[ $os_check =~ 'armv7l' ]]; then
    architecture="armv7"
    elif [[ $os_check =~ 'ppc64le' ]]; then
    architecture="ppc64le"
    elif [[ $os_check =~ 's390x' ]]; then
    architecture="s390x"
    elif [[ $os_check =~ 'riscv64' ]]; then
    architecture="riscv64"
else
    echo "当前系统架构暂不支持，请参考官方文档选择受支持的系统与架构。"
    exit 1
fi

# 获取版本号
if [[ -n "${VERSION}" ]]; then
    echo "已手动指定版本号：${VERSION}。"
else
    echo "未手动指定版本号，尝试获取最新版本。"
    # 尝试从 GitHub 获取最新 release tag
    VERSION=$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [[ "x${VERSION}" == "x" ]]; then
        echo "获取最新版本失败，请手动指定 VERSION 环境变量。"
        exit 1
    else
        echo "最新版本：${VERSION}。"
    fi
fi

# 通过 GitHub Release API 查询匹配当前架构的实际 asset 文件名
# goreleaser snapshot 模式生成的文件名包含 SNAPSHOT 和 commit hash，
# 格式类似：1panel-vsha-98575f3-SNAPSHOT-c023714-linux-amd64.tar.gz
# 因此不能拼接文件名，必须从 API 获取实际的 asset name
echo "正在从 GitHub Release ${VERSION} 查询 ${architecture} 架构的安装包..."
ASSET_INFO=$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/tags/${VERSION}")

# 从 API 返回的 JSON 中提取匹配 linux-${architecture}.tar.gz 的 asset name
package_file_name=$(echo "${ASSET_INFO}" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for asset in data.get('assets', []):
    name = asset['name']
    if name.endswith('.tar.gz') and f'linux-{{sys.argv[1]}}.' in name:
        print(name)
        sys.exit(0)
print('', end='')
sys.exit(1)
" "${architecture}")

if [[ -z "${package_file_name}" ]]; then
    echo "错误：在 Release ${VERSION} 中未找到 linux-${architecture} 架构的安装包。"
    echo "可用的 assets："
    echo "${ASSET_INFO}" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for asset in data.get('assets', []):
    print(f'  - {asset[\"name\"]}')
" 2>/dev/null || echo "  (无法解析)"
    exit 1
fi

echo "找到安装包：${package_file_name}"

# 推断解压后的目录名：去掉 .tar.gz 后缀
extracted_dir="${package_file_name%.tar.gz}"

# 构建下载链接（使用 gh-proxy 加速）
package_download_url="https://gh-proxy.com/https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${package_file_name}"

# 检查本地是否已存在安装包，若有则直接解压安装
if [[ -f "${package_file_name}" ]]; then
    echo "检测到本地安装包，直接解压安装。"
    rm -rf "${extracted_dir}"
    tar zxf "${package_file_name}"
    cd "${extracted_dir}"
    main_install
    exit 0
fi

# 下载安装包
echo "准备下载 1Panel ${VERSION}（架构：${architecture}）。"
echo "下载地址：${package_download_url}"
curl -fsSL -o "${package_file_name}" "${package_download_url}"
if [[ ! -f "${package_file_name}" ]] || [[ ! -s "${package_file_name}" ]]; then
    echo "下载安装包失败，请检查网络连接或版本号是否正确。"
    echo "若使用 sha-xxxx 版本，请确认该版本已成功构建并发布到 GitHub Release。"
    exit 1
fi

# 解压安装包
tar zxf "${package_file_name}"
if [[ $? != 0 ]]; then
    echo "解压安装包失败，下载文件可能不完整或已损坏。"
    rm -f "${package_file_name}"
    exit 1
fi

# 进入解压后的安装目录并调用主安装函数
cd "${extracted_dir}"
main_install
