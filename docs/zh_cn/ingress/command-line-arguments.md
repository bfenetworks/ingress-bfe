# BEF Ingress 启动参数
bfe ingress controller支持的参数如下

| 选项 | 默认值 | 用途|
| --- | --- | --- |
| -n |  | namespace: 设置监听的 namespace。默认监听所有的 namespace ，多个 namespace 之间用`,`分割。与 `-f`选项互斥。 |


示例：
```shell script
./bfe_ingress_controller -n name1,name2
```