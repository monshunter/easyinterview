#!/usr/bin/env bash
# easyInterview UI 原型一键运行脚本
# 用法:
#   ./run.sh                # 默认端口 5173, 入口 EasyInterview.html
#   ./run.sh -p 8080        # 指定端口
#   ./run.sh -f some.html   # 指定入口文件 (相对本目录)
#   ./run.sh --no-open      # 不自动打开浏览器
set -euo pipefail

cd "$(dirname "$0")"

PORT=5173
FILE="EasyInterview.html"
OPEN_BROWSER=1

while [[ $# -gt 0 ]]; do
  case "$1" in
    -p|--port) PORT="$2"; shift 2 ;;
    -f|--file) FILE="$2"; shift 2 ;;
    --no-open) OPEN_BROWSER=0; shift ;;
    -h|--help) sed -n '2,9p' "$0"; exit 0 ;;
    *) echo "未知参数: $1" >&2; exit 1 ;;
  esac
done

if [[ ! -f "$FILE" ]]; then
  echo "找不到入口文件: $FILE" >&2
  echo "本目录可用 .html:" >&2
  ls -1 *.html 2>/dev/null | sed 's/^/  /' >&2 || true
  exit 1
fi

# 端口被占用则自动 +1, 最多尝试 20 次
for _ in $(seq 1 20); do
  if ! lsof -nP -iTCP:"$PORT" -sTCP:LISTEN >/dev/null 2>&1; then break; fi
  PORT=$((PORT + 1))
done

# URL 编码入口文件名 (空格/中文/特殊字符)
url_encode() {
  python3 -c "import sys, urllib.parse as u; print(u.quote(sys.argv[1]))" "$1"
}
ENC_FILE="$(url_encode "$FILE")"
URL="http://localhost:${PORT}/${ENC_FILE}"

echo "[run.sh] 端口: $PORT  入口: $FILE"
echo "[run.sh] 访问: $URL"

if [[ "$OPEN_BROWSER" == "1" ]]; then
  ( sleep 1
    if command -v open >/dev/null 2>&1; then open "$URL"
    elif command -v xdg-open >/dev/null 2>&1; then xdg-open "$URL"
    fi
  ) &
fi

if command -v python3 >/dev/null 2>&1; then
  echo "[run.sh] 使用 python3 启动静态服务器"
  exec python3 -m http.server "$PORT" --bind 127.0.0.1
elif command -v python >/dev/null 2>&1; then
  echo "[run.sh] 使用 python 启动静态服务器"
  exec python -m SimpleHTTPServer "$PORT"
elif command -v npx >/dev/null 2>&1; then
  echo "[run.sh] 使用 npx serve 启动静态服务器"
  exec npx --yes serve -l "$PORT" .
else
  echo "未检测到 python3 / python / npx, 请先安装其一" >&2
  exit 1
fi
