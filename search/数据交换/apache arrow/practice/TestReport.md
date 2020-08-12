|操作|5K|467K|
|:---:|:---:|:---:|
|arrow 序列化|0.21|0.26|
|arrow 反序列化|0.10|0.11|
||||
|arrow ipc写|0.03|0.07|
|arrow ipc读|0.05|0.04|
||||
|pickle序列化|0.02|5.49|
|pickle反序列化|0.03|8.39|
||||
|arrow 序列化 dict|0.72|15.39|
|arrow 反序列化 dict|0.24|20.61|
|json.dumps dict|0.08|10.59|
|json.load to dict|0.06|10.49|
||||
|arrow 解析json文件|2.19|11.39|
|json.loads 解析json文件|0.16|12.25|
|arrow.Table to dict|0.45|50.29|
|dict to arrow.Table|0.18|9.94|
||||
|arrow 压缩|0.009|0.07|
|arrow 解压|0.008|0.04|
|arrow 压缩流|0.29|0.91|


|file|arrow Table|seialize buffer|seialize buffer compressed|ipc ouput|lz4 txt compressed|
|:---:|:-------:|:----:|:----:|:----:|:----|
|465K|262064|264128|66380|262904|2135|

|seialize dict|pickle dict|json.dumps dict|
|:----:|:----:|:----:|
|777664|298254|478199|