基于Simhash算法的url降重工具 <br />
使用示例: <br />
方法查询
```
urlde_v1.0.exe -h

Usage of urlde_v1.0.exe:
  -f string
        指定待去重url文件路径 (default "./urls.txt")
  -o string
        去重结果输出至文件 (default "./output.txt")
  -p string
        自定义host:path:param:frag:scheme的比例,请参照默认值格式输入 (default "4:3:2:0.5:0.5")
  -s float
        指定相似度,去除比较结果中高于该相似度的url (default 0.95)
```
<img src="https://cdn.jsdelivr.net/gh/0Rec1us/picorep//img/20240430101625.png" width="75%" height="auto"><br />
<br />
结果保存示例:<br />
<img src="https://cdn.jsdelivr.net/gh/0Rec1us/picorep//img/20240430101749.png" width="75%" height="auto"><br />
<br />
编译使用: go build -o urlde.exe main.go<br />

参考xq17文章链接:<br />
[https://xz.aliyun.com/t/13121?time__1311=mqmxnDBDcD0GQ0KDsQoYK0%3DFwZxI2GhGDbD&alichlgref=https%3A%2F%2Fwww.google.com%2F](https://xz.aliyun.com/t/13121?time__1311=mqmxnDBDcD0GQ0KDsQoYK0%3DFwZxI2GhGDbD&alichlgref=https%3A%2F%2Fwww.google.com%2F)<br />
基于Simhash项目:<br />
[https://github.com/mfonda/simhash](https://github.com/mfonda/simhash)<br />
复现思路+工具开发文章:<br />
[https://xrect1fy.github.io/2024/04/27/URL%E7%9B%B8%E4%BC%BC%E5%BA%A6%E5%8E%BB%E9%87%8D%E4%BB%8E%E5%A4%8D%E7%8E%B0%E5%88%B0%E5%B7%A5%E5%85%B7%E5%BC%80%E5%8F%91/](https://xrect1fy.github.io/2024/04/27/URL%E7%9B%B8%E4%BC%BC%E5%BA%A6%E5%8E%BB%E9%87%8D%E4%BB%8E%E5%A4%8D%E7%8E%B0%E5%88%B0%E5%B7%A5%E5%85%B7%E5%BC%80%E5%8F%91/)
