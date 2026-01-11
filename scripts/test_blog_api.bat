@echo off
REM 博客功能测试脚本 (Windows版本)
REM 使用方法: test_blog_api.bat

setlocal enabledelayedexpansion

set BASE_URL=http://localhost:8080

echo =========================================
echo 博客 API 测试脚本
echo =========================================
echo.

REM 检查服务器是否运行
echo 检查服务器状态...
curl -s "%BASE_URL%/health" >nul 2>&1
if errorlevel 1 (
    echo [错误] 服务器未运行！
    echo 请先启动服务器: go run main.go
    exit /b 1
)
echo [成功] 服务器正在运行
echo.

REM 1. 测试博客列表接口
echo =========================================
echo 1. 测试博客列表接口
echo =========================================
echo.

echo [测试] 获取默认列表（第1页，每页10条）
curl -s "%BASE_URL%/blog/list"
echo.
echo -----------------------------------------
echo.

echo [测试] 获取列表（第1页，每页5条）
curl -s "%BASE_URL%/blog/list?page=1&page_size=5"
echo.
echo -----------------------------------------
echo.

echo [测试] 按分类筛选（教程类）
curl -s "%BASE_URL%/blog/list?category=tutorials"
echo.
echo -----------------------------------------
echo.

echo [测试] 获取精选文章
curl -s "%BASE_URL%/blog/list?is_featured=true"
echo.
echo -----------------------------------------
echo.

echo [测试] 关键词搜索
curl -s "%BASE_URL%/blog/list?keyword=快速"
echo.
echo -----------------------------------------
echo.

echo [测试] 按热门排序
curl -s "%BASE_URL%/blog/list?sort=popular"
echo.
echo -----------------------------------------
echo.

REM 2. 测试文章详情接口
echo =========================================
echo 2. 测试文章详情接口
echo =========================================
echo.

echo [测试] 获取文章详情（存在）
curl -s "%BASE_URL%/blog/getting-started-with-01agent"
echo.
echo -----------------------------------------
echo.

echo [测试] 获取文章详情（不存在）
curl -s "%BASE_URL%/blog/non-existent-slug"
echo.
echo -----------------------------------------
echo.

REM 3. 测试 Sitemap 接口
echo =========================================
echo 3. 测试 Sitemap 接口
echo =========================================
echo.

echo [测试] 获取 sitemap 数据
curl -s "%BASE_URL%/blog/sitemap"
echo.
echo -----------------------------------------
echo.

REM 4. 测试相关文章接口
echo =========================================
echo 4. 测试相关文章推荐
echo =========================================
echo.

echo [测试] 获取相关文章（默认3条）
curl -s "%BASE_URL%/blog/getting-started-with-01agent/related"
echo.
echo -----------------------------------------
echo.

echo [测试] 获取相关文章（5条）
curl -s "%BASE_URL%/blog/getting-started-with-01agent/related?limit=5"
echo.
echo -----------------------------------------
echo.

REM 5. 测试浏览量统计接口
echo =========================================
echo 5. 测试浏览量统计
echo =========================================
echo.

echo [测试] 增加浏览量
curl -s -X POST "%BASE_URL%/blog/getting-started-with-01agent/view"
echo.
echo -----------------------------------------
echo.

echo.
echo =========================================
echo 测试完成！
echo =========================================
pause

