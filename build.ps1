# =====================================================================
# build.ps1 - 星芒桌面版发布打包入口
#
# 打包逻辑在 Go 工具 tools/release（go run ./tools/release），
# 本脚本只负责默认参数，用法与 build.bat 一致。
#
# 用法：
#   .\build.ps1                                # 默认: windows/amd64 + 前端 + 出包 dist\windows-amd64.zip
#   .\build.ps1 -Version 0.736                 # 指定版本号，归档名 dist\V-0.736.zip
#   .\build.ps1 -SkipFrontend                  # 跳过前端构建；web\dist 过期会自动重建
#   .\build.ps1 -NoZip                         # 只出 dist\<os>-<arch>\ 目录，不打归档
#   .\build.ps1 -IncludeMigrate                # 额外编译 migrate.exe
#   .\build.ps1 -OS linux -Arch arm64          # 交叉编译 linux/arm64
#   .\build.ps1 -OS windows -Arch amd64,arm64  # 同时编译多架构
#
# 产物：
#   dist\<os>-<arch>\        xingmang.exe + README.txt
#   dist\<os>-<arch>.zip     未指定版本号时的覆盖式「最新包」
#   dist\V-<版本>.zip        指定 -Version 时的版本化归档（多架构时为 V-<版本>-<os>-<arch>.zip）
# =====================================================================

Set-Location $PSScriptRoot
& go run ./tools/release -Zip @args
exit $LASTEXITCODE
