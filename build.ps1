# =====================================================================
# build.ps1 - 星芒桌面版发布打包入口
#
# 打包逻辑与**全部前置检查**都在 Go 工具 tools/release（go run ./tools/release）：
#   - 目标产物 exe 正被占用（无法覆盖）
#   - 缺 web\node_modules（无法构建前端）
#   - 产物目录里有跑出来的作品/索引（目录会被整个删掉）
#   - 缺图标资源 rsrc_windows_<arch>.syso（产物没有程序图标）
# 这些检查都在前端构建之前跑完，且与 build.bat 共用同一份实现，不会两边漂移。
#
# 用法：
#   .\build.ps1                                # 默认: windows/amd64 + 前端 + 出包 dist\windows-amd64.zip
#   .\build.ps1 -Version 0.772                 # 指定版本号，归档名 dist\V-0.772.zip
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
#
# ⚠ 版本号：未传 -Version 时取自 git 短 commit（工作区有改动会带 -dirty），
#    形如 `d7896fe-dirty`。**要交付给用户请显式传 -Version**，例如 -Version 0.772，
#    否则用户在界面上看到的版本号是一串哈希，无法和发布说明对应。
#
# ⚠ 前端产物：web\dist 是 //go:embed 的直接来源。不带 -SkipFrontend 时会重新构建；
#    带了则按 mtime 判断是否过期，过期会自动重建（见 tools/release 的 buildFrontend）。
# =====================================================================

Set-Location $PSScriptRoot

# 交给 Go 打包工具。前置检查已在其内部完成，这里不再重复 ——
# 之前 build.ps1 自己查一遍、build.bat 一遍都不查，两边行为不一致。
& go run ./tools/release -Zip @args
exit $LASTEXITCODE
