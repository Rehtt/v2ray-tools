# v2ray-tools
v2ray工具
- [v2ray-flow.py](#v2ray-flow.py)

## [v2ray-flow.py](https://github.com/Rehtt/v2ray-tools/blob/main/v2ray-flow.py)
查看v2ray多用户流量统计

使用前需要开启v2ray自带的统计工具，配置方法：https://guide.v2fly.org/advanced/traffic.html#%E9%85%8D%E7%BD%AE%E7%BB%9F%E8%AE%A1%E5%8A%9F%E8%83%BD

本脚本使用email作为用户标识，使用tag作为组标识

### 使用
```
v2ray-flow.py [-h] [-g <tag>] [-c <file>] [-s <server>]
默认显示全部用户

-h,--help       显示帮助
-g <tag>        显示组的流量
        -g A            (显示tag:A组的流量)

-c <file>       输入v2ray配置文件
        -c /etc/v2ray/config.json

-s <server>     输入v2ctl api server地址
        -s 127.0.0.1:53844
```
