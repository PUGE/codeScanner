import os
import json
import requests
import random
import string
import platform
import argparse
import webbrowser
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

class BrowserOpener:
    @staticmethod
    def open(url, browser=None, new_tab=True):
        """
        参数:
            url: 要打开的网址
            browser: None=系统默认 | 'chrome'|'firefox'|'edge'|'safari'
            new_tab: 是否在新标签页打开
        """
        try:
            if browser:
                # 处理Windows的Edge浏览器
                if browser.lower() == 'edge':
                    if platform.system() == 'Windows':
                        os.system(f'start microsoft-edge:{url}')
                        return True
                    browser = 'microsoft-edge'

                # 尝试获取指定浏览器
                try:
                    webbrowser.get(browser).open(url, new=new_tab)
                except webbrowser.Error:
                    return webbrowser.open(url, new=new_tab)
            else:
                return webbrowser.open(url, new=new_tab)
        except Exception as e:
            print(f"打开浏览器失败: {str(e)}")
            return False


def find_files(directory):
    """查找目录下所有PHP文件"""
    filesList = []
    ignored_dirs = {'vendor', 'node_modules', 'cache', 'temp'}
    for root, dirs, files in os.walk(directory):
        [dirs.remove(d) for d in list(dirs) if d in ignored_dirs]
        for file in files:
            file_path = os.path.join(root, file)
            # 检查文件大小是否超过1MB (1MB = 1024*1024 bytes)
            if os.path.getsize(file_path) <= 1024 * 1024:
                if file.endswith('.php'):
                    filesList.append([file_path, 'PHP'])
                if file.endswith('.py'):
                    filesList.append([file_path, 'Python'])
                if file.endswith('.bat'):
                    filesList.append([file_path, 'Bat'])
    return filesList

def safe_read_file(file_path, max_chars=60000):
    """安全读取文件，自动处理编码"""
    encodings_to_try = ['utf-8', 'gb18030', 'gbk', 'big5', 'utf-16']
    
    for encoding in encodings_to_try:
        try:
            with open(file_path, 'r', encoding=encoding) as f:
                content = f.read(max_chars + 1)  # 多读1个字符用于判断是否超限
                if len(content) > max_chars:
                    return None, f"跳过 {file_path} (文件大小超过{max_chars}字符限制)"
                return content, None
        except UnicodeDecodeError:
            continue
        except Exception as e:
            print(f"读取文件 {file_path} 时出错: {str(e)}")
            return None, str(e)
    
    # 如果所有编码都失败，尝试二进制读取
    try:
        with open(file_path, 'rb') as f:
            content = f.read().decode('utf-8', errors='replace')
            if len(content) > max_chars:
                return None, f"跳过 {file_path} (文件大小超过{max_chars}字符限制)"
            return content, None
    except Exception as e:
        print(f"最终尝试读取文件 {file_path} 失败: {str(e)}")
        return None, f"无法解码文件 {file_path}"

def post_file_content(url, file_path):
    """发送文件内容到指定URL"""
    try:
        # 读取文件内容
        content, error_msg = safe_read_file(file_path, 60000)
        if len(content) > 60000:  # 判断字符数
            return {"content": "跳过 " + file_path + " (文件大小超过限制)"}
        headers = {
            'Content-Type': 'application/json'
        }
        response = requests.post(url, data=json.dumps({'file_content': content, "file_name": file_path}), headers=headers)
        return response.json()
    except Exception as e:
        print(e)
        return {"content": str(e)}

def main():
    outPutInfo('查看代码检查报告访问: https://code.lamp.run/?id=' + taskID)
    parser = argparse.ArgumentParser(description='PHP文件扫描工具')
    parser.add_argument('--start-from', type=int, default=0,
                      help='从第几个文件开始扫描(0-based索引)')
    args = parser.parse_args()

    
    current_dir = os.getcwd()
    outPutInfo("正在搜索代码文件...")
    filesList = find_files(current_dir)
    total_files = len(filesList)
    outPutInfo(f"找到 {total_files} 个符合条件的代码文件")
    
    if total_files == 0:
        return
    
    
    # 使用tqdm创建进度条
    for i, fileInfo in enumerate(tqdm(filesList[args.start_from:], 
                                     desc="处理文件中",
                                     initial=args.start_from,
                                     total=len(filesList))):
        print(fileInfo)
        file_path = fileInfo[0]
        codeType = fileInfo[1]
        target_url = 'https://code.lamp.run/check' + codeType + '/' + taskID + '/' + str(total_files)
        post_file_content(target_url, file_path)

    BrowserOpener.open('https://code.lamp.run/?id=' + taskID, browser='chrome')


if __name__ == '__main__':
    main()