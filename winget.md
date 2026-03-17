> Winget 没有内置的全局安装位置设置，但你可以为支持自定义路径的包使用 --location 标志。

## 临时设置（每次安装时）：
```powershell
winget install --id=GNU.MinGW --location D:\SSoftwareFiles\winget
```

## 永久设置（创建 PowerShell 函数）：

### 编辑 PowerShell 配置文件：
```bash
nano $PROFILE
```
### 添加以下内容：
```powershell
function wingetd {
    param([string]$id, [string]$location = "D:\SSoftwareFiles\winget")
    winget install --id $id --location $location
}
```
### 保存并重启 PowerShell。

### 然后使用：
```bash
wingetd GNU.MinGW
```
> 如果 --location 不被 MinGW 包支持，请使用手动下载方式安装到 winget。

安装完成后，将 D:\SSoftwareFiles\winget\bin 添加到系统 PATH。