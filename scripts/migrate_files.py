#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
文件迁移脚本
用于从其他项目读取文件并复制到当前项目的 temp/ 目录中
主要用于代码迁移和分析
"""

import os
import shutil
import sys
import argparse
from pathlib import Path

# 修复Windows下的编码问题
if sys.platform == 'win32':
    import io
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8', errors='replace')
    sys.stderr = io.TextIOWrapper(sys.stderr.buffer, encoding='utf-8', errors='replace')


# 源项目路径
SOURCE_PROJECT_PATH = r"D:\project\2025\01editor\01editor-server"
# 目标临时目录
TEMP_DIR = "temp"
# 目标项目名称（在 temp 目录下的子目录名）
TARGET_PROJECT_NAME = "01editor-server"


def print_file_content(file_path, max_lines=50):
    """
    打印文件内容用于校验
    
    Args:
        file_path: 文件路径
        max_lines: 最大打印行数，默认50行
    """
    try:
        with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
            lines = f.readlines()
            total_lines = len(lines)
            
            print(f"\n{'='*80}")
            print(f"文件: {file_path}")
            print(f"总行数: {total_lines}")
            print(f"{'='*80}")
            
            # 打印前 max_lines 行
            for i, line in enumerate(lines[:max_lines], 1):
                print(f"{i:4d} | {line.rstrip()}")
            
            if total_lines > max_lines:
                print(f"\n... (还有 {total_lines - max_lines} 行未显示)")
            
            print(f"{'='*80}\n")
            
    except Exception as e:
        print(f"❌ 读取文件失败: {file_path}")
        print(f"错误: {e}")


def copy_file_to_temp(source_file, source_base_path, temp_base_dir):
    """
    将文件复制到 temp 目录，保持目录结构
    
    Args:
        source_file: 源文件路径（绝对路径或相对于 source_base_path 的路径）
        source_base_path: 源项目基础路径
        temp_base_dir: 临时目录基础路径
    
    Returns:
        tuple: (success: bool, dest_path: str)
    """
    try:
        # 处理路径
        if os.path.isabs(source_file):
            # 绝对路径
            full_source_path = source_file
            # 检查是否在源项目路径内
            if not full_source_path.startswith(source_base_path):
                print(f"⚠️  警告: 文件不在源项目路径内: {source_file}")
                return False, None
        else:
            # 相对路径，相对于源项目路径
            full_source_path = os.path.join(source_base_path, source_file)
        
        # 检查文件是否存在
        if not os.path.exists(full_source_path):
            print(f"❌ 文件不存在: {full_source_path}")
            return False, None
        
        if not os.path.isfile(full_source_path):
            print(f"❌ 不是文件: {full_source_path}")
            return False, None
        
        # 计算相对路径（相对于源项目基础路径）
        rel_path = os.path.relpath(full_source_path, source_base_path)
        
        # 构建目标路径
        dest_path = os.path.join(temp_base_dir, TARGET_PROJECT_NAME, rel_path)
        
        # 创建目标目录
        dest_dir = os.path.dirname(dest_path)
        os.makedirs(dest_dir, exist_ok=True)
        
        # 复制文件
        shutil.copy2(full_source_path, dest_path)
        
        print(f"✅ 已复制: {rel_path}")
        print(f"   源: {full_source_path}")
        print(f"   目标: {dest_path}")
        
        return True, dest_path
        
    except Exception as e:
        print(f"❌ 复制文件失败: {source_file}")
        print(f"错误: {e}")
        return False, None


def migrate_file(file_path, source_base_path, temp_base_dir, show_content=True):
    """
    迁移单个文件：读取、打印、复制
    
    Args:
        file_path: 文件路径
        source_base_path: 源项目基础路径
        temp_base_dir: 临时目录基础路径
        show_content: 是否显示文件内容
    """
    # 处理路径
    if os.path.isabs(file_path):
        full_path = file_path
    else:
        full_path = os.path.join(source_base_path, file_path)
    
    # 检查文件是否存在
    if not os.path.exists(full_path):
        print(f"❌ 文件不存在: {full_path}")
        return False
    
    if not os.path.isfile(full_path):
        print(f"❌ 不是文件: {full_path}")
        return False
    
    # 打印文件内容用于校验
    if show_content:
        print_file_content(full_path)
    
    # 复制文件
    success, dest_path = copy_file_to_temp(full_path, source_base_path, temp_base_dir)
    
    if success:
        print(f"✅ 文件迁移成功: {file_path}")
        return True
    else:
        print(f"❌ 文件迁移失败: {file_path}")
        return False


def migrate_directory(dir_path, source_base_path, temp_base_dir, show_content=False, 
                     file_extensions=None, exclude_dirs=None):
    """
    迁移整个目录
    
    Args:
        dir_path: 目录路径（相对于 source_base_path）
        source_base_path: 源项目基础路径
        temp_base_dir: 临时目录基础路径
        show_content: 是否显示每个文件的内容
        file_extensions: 要包含的文件扩展名列表，None 表示所有文件
        exclude_dirs: 要排除的目录列表（如 .git, node_modules 等）
    """
    if exclude_dirs is None:
        exclude_dirs = ['.git', '.svn', 'node_modules', '__pycache__', '.idea', '.vscode']
    
    full_dir_path = os.path.join(source_base_path, dir_path)
    
    if not os.path.exists(full_dir_path):
        print(f"❌ 目录不存在: {full_dir_path}")
        return
    
    if not os.path.isdir(full_dir_path):
        print(f"❌ 不是目录: {full_dir_path}")
        return
    
    print(f"\n📁 开始迁移目录: {dir_path}")
    print(f"源路径: {full_dir_path}\n")
    
    copied_count = 0
    failed_count = 0
    
    # 遍历目录
    for root, dirs, files in os.walk(full_dir_path):
        # 排除指定目录
        dirs[:] = [d for d in dirs if d not in exclude_dirs]
        
        # 排除 .git 等隐藏目录
        dirs[:] = [d for d in dirs if not d.startswith('.')]
        
        for file in files:
            # 过滤文件扩展名
            if file_extensions:
                ext = os.path.splitext(file)[1]
                if ext not in file_extensions:
                    continue
            
            file_path = os.path.join(root, file)
            rel_path = os.path.relpath(file_path, source_base_path)
            
            # 复制文件
            success, _ = copy_file_to_temp(file_path, source_base_path, temp_base_dir)
            
            if success:
                copied_count += 1
                # 可选：显示文件内容（仅前几个文件）
                if show_content and copied_count <= 5:
                    print_file_content(file_path, max_lines=20)
            else:
                failed_count += 1
    
    print(f"\n📊 迁移完成:")
    print(f"   成功: {copied_count} 个文件")
    print(f"   失败: {failed_count} 个文件")


def main():
    parser = argparse.ArgumentParser(
        description='从其他项目迁移文件到当前项目的 temp/ 目录',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
示例:
  # 迁移单个文件
  python scripts/migrate_files.py src/main.py
  
  # 迁移单个文件（绝对路径）
  python scripts/migrate_files.py "D:\\project\\2025\\01editor\\01editor-server\\src\\main.py"
  
  # 迁移多个文件
  python scripts/migrate_files.py src/main.py src/config.py src/utils.py
  
  # 迁移整个目录
  python scripts/migrate_files.py --dir src
  
  # 迁移整个目录，只包含特定扩展名的文件
  python scripts/migrate_files.py --dir src --ext .py .js .ts
  
  # 迁移文件但不显示内容
  python scripts/migrate_files.py src/main.py --no-content
        """
    )
    
    parser.add_argument(
        'files',
        nargs='*',
        help='要迁移的文件路径（相对或绝对路径）'
    )
    
    parser.add_argument(
        '--dir', '-d',
        help='迁移整个目录'
    )
    
    parser.add_argument(
        '--source', '-s',
        default=SOURCE_PROJECT_PATH,
        help=f'源项目路径（默认: {SOURCE_PROJECT_PATH}）'
    )
    
    parser.add_argument(
        '--temp', '-t',
        default=TEMP_DIR,
        help=f'临时目录（默认: {TEMP_DIR}）'
    )
    
    parser.add_argument(
        '--no-content',
        action='store_true',
        help='不显示文件内容'
    )
    
    parser.add_argument(
        '--ext',
        nargs='+',
        help='只迁移指定扩展名的文件（用于目录迁移）'
    )
    
    args = parser.parse_args()
    
    # 检查源项目路径
    if not os.path.exists(args.source):
        print(f"❌ 源项目路径不存在: {args.source}")
        sys.exit(1)
    
    # 创建临时目录
    temp_base_dir = os.path.abspath(args.temp)
    os.makedirs(temp_base_dir, exist_ok=True)
    
    print(f"📦 文件迁移工具")
    print(f"源项目: {args.source}")
    print(f"临时目录: {temp_base_dir}")
    print()
    
    # 处理目录迁移
    if args.dir:
        file_extensions = None
        if args.ext:
            file_extensions = args.ext
        
        migrate_directory(
            args.dir,
            args.source,
            temp_base_dir,
            show_content=not args.no_content,
            file_extensions=file_extensions
        )
    
    # 处理文件迁移
    if args.files:
        success_count = 0
        failed_count = 0
        
        for file_path in args.files:
            if migrate_file(
                file_path,
                args.source,
                temp_base_dir,
                show_content=not args.no_content
            ):
                success_count += 1
            else:
                failed_count += 1
        
        print(f"\n📊 迁移完成:")
        print(f"   成功: {success_count} 个文件")
        print(f"   失败: {failed_count} 个文件")
    
    if not args.dir and not args.files:
        parser.print_help()


if __name__ == '__main__':
    main()

