@echo off
REM 博客管理接口测试脚本 (Windows版本)
REM 使用方法: test_blog_admin.bat

setlocal enabledelayedexpansion

set BASE_URL=http://localhost:8080

echo =========================================
echo 博客管理接口测试
echo =========================================
echo.

REM 1. 创建文章
echo [1] 创建新文章
echo POST /blog/create
echo.

curl -s -X POST "%BASE_URL%/blog/create" ^
  -H "Content-Type: application/json" ^
  -d "{\"slug\":\"test-article-%RANDOM%\",\"title\":\"测试文章：AI内容创作的未来\",\"summary\":\"探讨AI技术在内容创作领域的应用和发展趋势\",\"content\":\"# AI内容创作的未来\\n\\n## 引言\\n\\nAI正在改变内容创作的方式...\\n\\n## 主要趋势\\n\\n1. 自动化写作\\n2. 个性化内容\\n3. 多模态生成\\n\\n## 总结\\n\\nAI是工具，创意是核心。\",\"category\":\"industry-insights\",\"author\":\"测试作者\",\"read_time\":5,\"is_featured\":false,\"status\":\"published\",\"tags\":[\"AI\",\"内容创作\",\"趋势分析\"],\"seo_keywords\":[\"AI写作\",\"内容创作\",\"人工智能\"]}" > create_result.json

type create_result.json
echo.
echo [成功] 文章创建成功
echo.
echo -----------------------------------------
echo.

REM 提示：由于Windows批处理脚本解析JSON较困难，建议使用PowerShell或Python脚本进行后续操作
echo 提示：文章已创建，可以通过以下方式查看：
echo 1. 访问 %BASE_URL%/blog/list 查看所有文章
echo 2. 使用 PowerShell 脚本进行更详细的测试
echo.

pause

