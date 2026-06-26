# 生成JWT secret。

## 使用方法

保存为 `generate-jwt-secret`，然后：

```bash

cd scripts/generate-jwt-secret

# 添加执行权限
chmod +x generate-jwt-secret

# 基本使用（生成32字节十六进制密钥）
./generate-jwt-secret

# 生成长度64的十六进制密钥
./generate-jwt-secret -l 64

# 生成 Base64 格式密钥
./generate-jwt-secret -b

# 生成 UUID 格式
./generate-jwt-secret -u

# 生成简单字母数字密钥
./generate-jwt-secret -s -l 48

# 静默模式（只输出密钥，适合脚本调用）
./generate-jwt-secret -q

# 不换行输出
./generate-jwt-secret -n
```

## 输出示例

```bash
$ ./generate-jwt-secret
✓ JWT Secret 生成成功
格式: hex
长度: 32
---
a7f3b8c9d1e2f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8
---
提示: 请将以上密钥复制到 .env 文件中
```

这个脚本会自动检测系统环境，优先使用 `openssl`（如果可用），否则使用系统原生命令，确保在 macOS 和 Linux 上都能正常工作。