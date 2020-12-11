#!/bin/python3

# Created by Rehtt on 2020/12/11
# 查看v2ray多用户流量统计
# 使用前需要开启v2ray自带的统计工具，配置方法：https://guide.v2fly.org/advanced/traffic.html#%E9%85%8D%E7%BD%AE%E7%BB%9F%E8%AE%A1%E5%8A%9F%E8%83%BD
# 本脚本使用email作为用户标识，使用tag作为组标识

import re
import os
import json
import getopt
import sys

v2ConfigFile = '/etc/v2ray/config.json'  # v2ray配置文件位置
v2ctlServer = '127.0.0.1:53844'  # v2ctl api配置信息


def dataSize(size, i=0):
    if size > 1e3:
        size /= 1e3
        s = dataSize(size, i+1)
    else:
        size = round(size, 2)
        if i == 1:
            s = str(size)+'KB'
        elif i == 2:
            s = str(size)+'MB'
        elif i >= 3:
            s = str(size)+'GB'
    return s


def red(f):
    return '\033[31m'+f+'\033[0m'


def help():
    print("-h,--help\t显示帮助")
    print("-g <tag>\t显示组的流量")
    print("\t-g A\t\t(显示tag:A组的流量)\n")
    # print("-u <email>\t显示单个用户流量")
    # print("\t-u rehtt@qq.com\t(显示email为rehtt@qq.com的流量)\n")
    print("-c <file>\t输入v2ray配置文件")
    print("\t-c /etc/v2ray/config.json\n")
    print("-s <server>\t输入v2ctl api server地址")
    print("\t-s 127.0.0.1:53844")


def getDownUp(name, res):
    i = 0
    z = 0
    down = 0
    up = 0
    try:
        for r in res:
            i += 1
            if name in r:
                z += 1
                if 'downlink' in r:
                    down = dataSize(float(re.sub('\D', '', res[i])))
                else:
                    up = dataSize(float(re.sub('\D', '', res[i])))
                if z == 2:
                    break
    except:
        pass
    return down, up


def main(argv):
    global v2ctlServer, v2ConfigFile
    c = '/usr/bin/v2ray/v2ctl api --server={server} StatsService.QueryStats \'pattern: "{q}" reset: false\''
    try:
        opts, args = getopt.getopt(argv, "hg:c:s:u:", ["help"])
    except getopt.GetoptError:
        help()
        sys.exit(2)
    q = 'user'
    G = 0
    for opt, arg in opts:
        if opt in ('-h', '--help'):
            help()
            sys.exit()
        elif opt == '-g':
            q = arg
            G = 1
        # elif opt == '-u':
        #     q = arg
        #     G = 2
        if opt == '-c':
            v2ConfigFile = arg
        if opt == '-s':
            v2ctlServer = arg
    tags = {}

    c = c.format(server=v2ctlServer, q=q)
    res = os.popen(c).readlines()
    if G == 0:
        with open(v2ConfigFile, 'r') as f:
            data = json.load(f)
            for i in data['inbounds']:
                if i['tag'] != 'api':
                    emails = []
                    for email in i['settings']['clients']:
                        emails.append(email['email'])
                    tags[i['tag']] = emails
        for tag in tags:
            for email in tags[tag]:
                down, up = getDownUp(email, res)
                print('Email:', red(email), '\tTag:'+red(tag))
                print('上传:', up, '\t下载:', down, '\n')
    else:
        if G==1:
            print('Tag:', red(q))
            down, up = getDownUp(q, res)
        # else:
        #     print('Email:', red(q))
        print('上传:', up, '\t下载:', down, '\n')



if __name__ == "__main__":
    main(sys.argv[1:])
