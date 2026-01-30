# 说明

自动读取中兴 F50Pro 内 SIM 卡短信, 并通过 bark 通知到手机。

要求程序运行在能够访问中兴 F50Pro 的网络环境中。

# 短信 API 文档

当前[API文档](./api.md)为AI(kimi K2.5)在`F50ProV1.0.0B15`固件版本下分析并生成, 不保证正确性。

API文档仅作参考

# 使用方法

`./zte.exe -p [password] -b [barkKey] -s [sound] -url http://192.168.0.1`

- `-p` 中兴 F50Pro 登录密码
  - 如：`u7A1g`
- `-b` bark 通知 key
  - 如：`Rr2YbnXjGzqNB73ccTo2`（单设备端）
  - 如：`Rr2YbnXjGzqNB73ccTo2,Tr3Yb6XjUz9NB73ccFg1` (多设备端用英文逗号分隔)
- `-s` bark 通知铃声名称
  - 可选项，默认为：`healthnotification`
- `-url` 中兴 F50Pro 地址
  - 可选项，默认为：`http://192.168.0.1`

# 编译

```shell
go build ./...
```