import os
import json
import requests
import random
import string
import argparse
from tqdm import tqdm

infoTemp = ''
def outPutInfo(info):
    global infoTemp
    print(info)
    infoTemp = infoTemp + info + '\n'



def generate_random_string(length=8):
    # 包含大小写字母和数字
    characters = string.ascii_letters + string.digits
    return ''.join(random.choice(characters) for _ in range(length))

taskID = generate_random_string(8) 


outPutInfo('查看详细报告访问: https://code.lamp.run/?id=' + taskID)

def find_php_files(directory):
    """查找目录下所有PHP文件"""
    php_files = []
    ignored_dirs = {'vendor', 'node_modules', 'cache', 'temp'}
    for root, dirs, files in os.walk(directory):
        [dirs.remove(d) for d in list(dirs) if d in ignored_dirs]
        for file in files:
            if file.endswith('.php'):
                file_path = os.path.join(root, file)
                # 检查文件大小是否超过1MB (1MB = 1024*1024 bytes)
                if os.path.getsize(file_path) <= 1024 * 1024:
                    php_files.append(file_path)
    return php_files

def post_file_content(url, file_path):
    """发送文件内容到指定URL"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        if len(content) > 60000:  # 判断字符数
            return {"content": "跳过 " + file_path + " (文件大小超过限制)"}
        headers = {
            'Content-Type': 'application/json'
        }
        response = requests.post(url, data=json.dumps({'file_content': content, "file_name": file_path}), headers=headers)
        return response.json()
    except Exception as e:
        return {"content": str(e)}

def main():
    parser = argparse.ArgumentParser(description='PHP文件扫描工具')
    parser.add_argument('--start-from', type=int, default=0,
                      help='从第几个文件开始扫描(0-based索引)')
    parser.add_argument('--output', default='检查结果.txt',
                      help='输出日志文件名')
    args = parser.parse_args()

    
    current_dir = os.getcwd()
    outPutInfo("正在搜索PHP文件...")
    php_files = find_php_files(current_dir)
    total_files = len(php_files)
    outPutInfo(f"找到 {total_files} 个符合条件的PHP文件")
    
    if total_files == 0:
        return
    
    target_url = 'https://code.lamp.run/checkPHP/' + taskID + '/' + str(total_files)
    # 使用tqdm创建进度条
    for i, file_path in enumerate(tqdm(php_files[args.start_from:], 
                                     desc="处理文件中",
                                     initial=args.start_from,
                                     total=len(php_files))):
        result = post_file_content(target_url, file_path)
        # outPutInfo(f"\n文件: {file_path}")
        # outPutInfo(f"检查结果: {result}")
        # outPutInfo("-" * 50)

    with open("./" + args.output, 'a', encoding='utf-8') as f:
        f.write(infoTemp)

if __name__ == '__main__':
    main()