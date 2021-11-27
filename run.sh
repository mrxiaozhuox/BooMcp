# 使用本文件可直接编译并运行项目到 `env` 目录下（有利于配置生成）
# .env 目录下将自动加载相关数据信息

go build -o ./.env/boomcp

# UINT 会启用运行目录下的模板代码，而并非 Binary 包中的代码

./.env/boomcp UINT