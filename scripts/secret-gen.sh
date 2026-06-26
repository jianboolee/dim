#!/bin/bash

# JWT Secret 生成器
# 支持 macOS 和 Linux

set -e

# 颜色定义（可选）
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 显示帮助信息
show_help() {
    cat << EOF
用法: $0 [选项]

选项:
  -l, --length LENGTH    生成密钥的长度（默认: 32）
  -b, --base64           输出 Base64 编码格式
  -h, --hex              输出十六进制格式（默认）
  -u, --uuid             生成 UUID 格式的密钥
  -s, --simple           生成简单的字母数字密钥
  -n, --no-newline       输出不包含换行符
  -q, --quiet            静默模式，只输出密钥
  --help                 显示此帮助信息

示例:
  $0                     # 生成 32 字节十六进制密钥
  $0 -l 64               # 生成 64 字节十六进制密钥
  $0 -b                  # 生成 Base64 编码密钥
  $0 -u                  # 生成 UUID 格式密钥
  $0 -s -l 48            # 生成 48 字符字母数字密钥

EOF
}

# 默认参数
LENGTH=32
FORMAT="hex"
NO_NEWLINE=false
QUIET=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -l|--length)
            LENGTH="$2"
            shift 2
            ;;
        -b|--base64)
            FORMAT="base64"
            shift
            ;;
        -h|--hex)
            FORMAT="hex"
            shift
            ;;
        -u|--uuid)
            FORMAT="uuid"
            shift
            ;;
        -s|--simple)
            FORMAT="simple"
            shift
            ;;
        -n|--no-newline)
            NO_NEWLINE=true
            shift
            ;;
        -q|--quiet)
            QUIET=true
            shift
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}错误: 未知选项 $1${NC}" >&2
            show_help
            exit 1
            ;;
    esac
done

# 检测操作系统并选择生成方式
generate_secret() {
    local os_type
    os_type=$(uname -s)
    
    case $FORMAT in
        uuid)
            # UUID 格式
            if command -v uuidgen &> /dev/null; then
                # macOS/Linux 都支持 uuidgen
                uuidgen
            elif [[ -f /proc/sys/kernel/random/uuid ]]; then
                # Linux 内核接口
                cat /proc/sys/kernel/random/uuid
            else
                # 使用 openssl 生成 UUID v4
                openssl rand -hex 16 | sed 's/\(..\)\(..\)\(..\)\(..\)\(..\)\(..\)\(..\)\(..\)/\1\2\3\4-\5\6-\7\8-9\1-\2\3\4\5\6\7\8/'
            fi
            ;;
            
        base64)
            # Base64 编码
            if command -v openssl &> /dev/null; then
                openssl rand -base64 "$LENGTH" | tr -d '\n'
            elif [[ "$os_type" == "Darwin" ]]; then
                # macOS 使用 dd + base64
                dd if=/dev/urandom bs=1 count="$LENGTH" 2>/dev/null | base64 | tr -d '\n'
            else
                # Linux 使用 head + base64
                head -c "$LENGTH" /dev/urandom | base64 | tr -d '\n'
            fi
            ;;
            
        simple)
            # 简单字母数字密钥
            if command -v openssl &> /dev/null; then
                openssl rand -base64 "$LENGTH" 2>/dev/null | tr -dc 'A-Za-z0-9' | cut -c1-"$LENGTH"
            else
                tr -dc 'A-Za-z0-9' < /dev/urandom 2>/dev/null | head -c "$LENGTH"
            fi
            ;;
            
        hex|*)
            # 十六进制格式（默认）
            if command -v openssl &> /dev/null; then
                openssl rand -hex "$LENGTH" 2>/dev/null
            elif [[ "$os_type" == "Darwin" ]]; then
                # macOS: 使用 dd + hexdump
                dd if=/dev/urandom bs=1 count="$LENGTH" 2>/dev/null | hexdump -v -e '1/1 "%02x"'
            else
                # Linux: 使用 head + hexdump
                head -c "$LENGTH" /dev/urandom | hexdump -v -e '1/1 "%02x"'
            fi
            ;;
    esac
}

# 生成并输出密钥
SECRET=$(generate_secret)

# 根据选项决定是否添加换行符
if [[ "$NO_NEWLINE" == true ]]; then
    OUTPUT="$SECRET"
else
    OUTPUT="$SECRET"$'\n'
fi

# 输出结果
if [[ "$QUIET" == true ]]; then
    echo -n "$OUTPUT"
else
    # 显示生成信息
    echo -e "${GREEN}✓ JWT Secret 生成成功${NC}" >&2
    echo -e "${YELLOW}格式: ${FORMAT}${NC}" >&2
    echo -e "${YELLOW}长度: ${LENGTH}${NC}" >&2
    echo "---" >&2
    echo -n "$OUTPUT"
    echo "---" >&2
    echo -e "${GREEN}提示: 请将以上密钥复制到 .env 文件中${NC}" >&2
fi