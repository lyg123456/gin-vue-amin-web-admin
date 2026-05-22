#!/usr/bin/env bash
# Linux: 安装 LibreOffice + FFmpeg，并检测可执行路径
# 用法（在仓库根目录）:
#   chmod +x scripts/install-office-tools-linux.sh
#   sudo ./scripts/install-office-tools-linux.sh
#
# 安装后请在 server/config.yaml 的 office-tools 中填写（若未自动加入 PATH）:
#   libreoffice-path: "/usr/bin/soffice"
#   ffmpeg-path: "/usr/bin/ffmpeg"

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CONFIG="${REPO_ROOT}/server/config.yaml"

echo "==> 检测包管理器..."
if command -v apt-get >/dev/null 2>&1; then
  sudo apt-get update -qq
  sudo DEBIAN_FRONTEND=noninteractive apt-get install -y \
    libreoffice libreoffice-writer libreoffice-calc libreoffice-impress \
    ffmpeg
elif command -v dnf >/dev/null 2>&1; then
  sudo dnf install -y libreoffice ffmpeg
elif command -v yum >/dev/null 2>&1; then
  sudo yum install -y libreoffice ffmpeg
elif command -v apk >/dev/null 2>&1; then
  sudo apk add --no-cache libreoffice ffmpeg
elif command -v pacman >/dev/null 2>&1; then
  sudo pacman -Sy --noconfirm libreoffice-fresh ffmpeg
elif command -v zypper >/dev/null 2>&1; then
  sudo zypper install -y libreoffice ffmpeg
else
  echo "未识别包管理器，请手动安装 libreoffice 与 ffmpeg 后配置 config.yaml"
  exit 1
fi

SOFFICE=""
FFMPEG=""
for c in /usr/bin/soffice /usr/local/bin/soffice "$(command -v soffice 2>/dev/null || true)"; do
  if [[ -n "$c" && -x "$c" ]]; then SOFFICE="$c"; break; fi
done
for c in /usr/bin/ffmpeg /usr/local/bin/ffmpeg "$(command -v ffmpeg 2>/dev/null || true)"; do
  if [[ -n "$c" && -x "$c" ]]; then FFMPEG="$c"; break; fi
done

echo ""
echo "==> 检测结果"
if [[ -n "$SOFFICE" ]]; then
  echo "LibreOffice: $SOFFICE"
  "$SOFFICE" --version | head -1 || true
else
  echo "LibreOffice: 未找到 soffice，请检查安装"
fi
if [[ -n "$FFMPEG" ]]; then
  echo "FFmpeg: $FFMPEG"
  "$FFMPEG" -version | head -1 || true
else
  echo "FFmpeg: 未找到 ffmpeg"
fi

echo ""
echo "==> 请在 server/config.yaml 的 office-tools 中设置（示例）:"
echo "    libreoffice-path: \"${SOFFICE:-/usr/bin/soffice}\""
echo "    ffmpeg-path: \"${FFMPEG:-/usr/bin/ffmpeg}\""
echo ""
echo "Docker 部署时需在镜像 Dockerfile 中安装上述包，不能只装 Go 二进制。"
echo "无图形界面服务器一般无需额外配置，soffice 使用 --headless 模式。"

if [[ -f "$CONFIG" ]]; then
  echo "配置文件: $CONFIG"
fi
