# ============================================================
# deploy_quick.py — 一键快捷部署脚本
# 用法：
#   python deploy_quick.py           部署前端 + 后端
#   python deploy_quick.py frontend  仅部署前端
#   python deploy_quick.py backend   仅部署后端
#
# 首次使用请复制 .deploy_config.example 为 .deploy_config
# 并填写目标服务器信息，该文件已被 .gitignore 保护
# ============================================================

import os
import sys
import subprocess
import json
import shutil
import tempfile
from pathlib import Path
from datetime import datetime

# ------------------------------------------------------------------
# 读取部署配置
# ------------------------------------------------------------------
CONFIG_FILE = Path(__file__).parent / ".deploy_config"
EXAMPLE_FILE = Path(__file__).parent / ".deploy_config.example"


def load_config():
    """从 .deploy_config 读取配置"""
    if not CONFIG_FILE.exists():
        print(f"❌ 未找到配置文件 {CONFIG_FILE}")
        print(f"   请复制 {EXAMPLE_FILE} 为 {CONFIG_FILE} 并填写服务器信息")
        sys.exit(1)

    config = {}
    with open(CONFIG_FILE, "r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            if "=" not in line:
                continue
            key, _, value = line.partition("=")
            key = key.strip()
            value = value.strip().strip('"').strip("'")
            config[key] = value
    return config


# ------------------------------------------------------------------
# 辅助工具
# ------------------------------------------------------------------
def log(msg):
    print(f"[{datetime.now().strftime('%H:%M:%S')}] {msg}")


def run_local(cmd, cwd=None, capture=False):
    """执行本地命令"""
    log(f"🏃 执行: {cmd}")
    result = subprocess.run(
        cmd if isinstance(cmd, list) else cmd,
        shell=isinstance(cmd, str),
        cwd=cwd or Path.cwd(),
        capture_output=capture,
        text=True,
    )
    if result.returncode != 0:
        print(f"   返回码: {result.returncode}")
        if capture:
            print(f"   stderr: {result.stderr}")
        raise RuntimeError(f"命令失败: {cmd}")
    return result


# ------------------------------------------------------------------
# 前端部署
# ------------------------------------------------------------------
def build_and_deploy_frontend(cfg, sftp, ssh):
    local_dir = Path(cfg.get("LOCAL_FRONTEND_DIR", "./frontend"))
    remote_dir = cfg["REMOTE_FRONTEND_DIR"]

    if not (local_dir / "package.json").exists():
        log(f"⚠️  未在 {local_dir} 找到 package.json，跳过前端部署")
        return

    # 1. 安装依赖（可选）
    if cfg.get("UPDATE_DEPS", "false").lower() == "true":
        run_local("npm install", cwd=str(local_dir))

    # 2. 构建
    log("🔨 构建前端...")
    run_local("npm run build", cwd=str(local_dir))
    log("✅ 前端构建完成")

    # 3. 上传
    dist_dir = local_dir / "dist"
    if not dist_dir.exists():
        log(f"❌ 构建产物目录 {dist_dir} 不存在")
        return

    log(f"📤 上传前端到 {remote_dir} ...")
    upload_dir_via_sftp(sftp, str(dist_dir), remote_dir)

    # 4. Nginx 重载
    log("🔄 重载 Nginx ...")
    try:
        ssh_exec(ssh, "sudo nginx -t && sudo nginx -s reload")
        log("✅ Nginx 重载成功")
    except RuntimeError as e:
        log(f"⚠️  Nginx 重载失败: {e}")

    # 5. 验证
    log("🔍 验证前端...")
    try:
        ssh_exec(ssh, "curl -s -o /dev/null -w '%{http_code}' https://khmerai.cn/ | grep -q 200")
        log("✅ 前端返回 HTTP 200")
    except RuntimeError:
        log("⚠️  HTTP 200 验证未通过（可能是正常域名解析问题）")


# ------------------------------------------------------------------
# 后端部署
# ------------------------------------------------------------------
def build_and_deploy_backend(cfg, sftp, ssh):
    local_dir = Path(cfg.get("LOCAL_BACKEND_DIR", "./backend-go"))
    remote_dir = cfg["REMOTE_BACKEND_DIR"]
    service = cfg.get("BACKEND_SERVICE", "wholesale")

    if not (local_dir / "go.mod").exists():
        log(f"⚠️  未在 {local_dir} 找到 go.mod，跳过后端部署")
        return

    # 1. 构建（交叉编译为 Linux amd64）
    log("🔨 构建后端 (Go → linux/amd64)...")
    binary_name = "wholesale_linux"
    env = os.environ.copy()
    env["GOOS"] = "linux"
    env["GOARCH"] = "amd64"
    result = subprocess.run(
        f"go build -o {binary_name}",
        shell=True,
        cwd=str(local_dir),
        env=env,
    )
    if result.returncode != 0:
        raise RuntimeError("后端构建失败")

    binary_path = local_dir / binary_name
    if not binary_path.exists():
        log(f"❌ 构建产物 {binary_path} 不存在")
        return
    log("✅ 后端构建完成")

    # 2. 上传并重启（先传到 /tmp，再 sudo 移动）
    tmp_binary = "/tmp/wholesale"
    remote_binary = f"{remote_dir}/wholesale"
    log(f"📤 上传后端到 {tmp_binary} ...")
    sftp.put(str(binary_path), tmp_binary)
    log("✅ 上传完成")

    log("🔄 重启后端服务...")
    try:
        ssh_exec(ssh, (
            f"sudo systemctl stop {service} && "
            f"sudo mkdir -p {remote_dir} && "
            f"sudo mv {tmp_binary} {remote_binary} && "
            f"sudo chmod +x {remote_binary} && "
            f"sudo systemctl start {service}"
        ))
        # 检查状态
        status = ssh_exec(ssh, f"sudo systemctl is-active {service}", capture=True)
        log(f"✅ 后端服务状态: {status.strip()}")
    except RuntimeError as e:
        log(f"❌ 后端重启失败: {e}")


# ------------------------------------------------------------------
# SFTP / SSH 辅助
# ------------------------------------------------------------------
def create_ssh_sftp(cfg):
    """创建 SSH 和 SFTP 连接"""
    import paramiko

    host = cfg["HOST"]
    port = int(cfg.get("PORT", 22))
    user = cfg["USER"]

    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())

    if "SSH_KEY_PATH" in cfg and cfg["SSH_KEY_PATH"]:
        key_path = os.path.expanduser(cfg["SSH_KEY_PATH"])
        log(f"🔑 使用 SSH 密钥连接 {user}@{host}:{port} ...")
        ssh.connect(host, port=port, username=user, key_filename=key_path)
    else:
        password = cfg.get("PASSWORD", "")
        log(f"🔑 使用密码连接 {user}@{host}:{port} ...")
        ssh.connect(host, port=port, username=user, password=password)

    sftp = ssh.open_sftp()
    log(f"✅ 已连接到 {host}")
    return ssh, sftp


def ssh_exec(ssh, command, capture=False):
    """在远程服务器上执行命令"""
    stdin, stdout, stderr = ssh.exec_command(command)
    exit_code = stdout.channel.recv_exit_status()
    output = stdout.read().decode("utf-8", errors="replace").strip()

    if exit_code != 0:
        err = stderr.read().decode("utf-8", errors="replace").strip()
        msg = f"远程命令失败 (exit={exit_code}): {command}"
        if err:
            msg += f"\n  错误: {err}"
        raise RuntimeError(msg)

    if capture:
        return output
    return output


def upload_dir_via_sftp(sftp, local_dir, remote_dir):
    """递归上传目录到远端"""
    import posixpath

    local_path = Path(local_dir)
    remote_dir = remote_dir.rstrip("/")

    # 确保远端目录存在
    _mkdir_sftp(sftp, remote_dir)

    for root, dirs, files in os.walk(local_path):
        rel_root = os.path.relpath(root, local_path)
        remote_path = posixpath.join(remote_dir, rel_root) if rel_root != "." else remote_dir

        for d in dirs:
            _mkdir_sftp(sftp, posixpath.join(remote_path, d))

        for f in files:
            local_file = os.path.join(root, f)
            remote_file = posixpath.join(remote_path, f)
            sftp.put(local_file, remote_file)

    log(f"  已上传 {len(list(local_path.rglob('*')))} 个文件")


def _mkdir_sftp(sftp, path):
    """递归创建远端目录"""
    import posixpath
    path = path.rstrip("/")
    parts = path.split("/")
    current = "/" if path.startswith("/") else ""
    for part in parts:
        if not part:
            continue
        current = posixpath.join(current, part)
        try:
            sftp.stat(current)
        except FileNotFoundError:
            sftp.mkdir(current)


# ------------------------------------------------------------------
# 入口
# ------------------------------------------------------------------
def show_config_summary(cfg):
    """显示配置摘要（隐藏密码）"""
    host = cfg.get("HOST", "?")
    user = cfg.get("USER", "?")
    pw = "***" if "PASSWORD" in cfg and cfg["PASSWORD"] else "🖴 SSH密钥"
    fe = cfg.get("REMOTE_FRONTEND_DIR", "?")
    be = cfg.get("REMOTE_BACKEND_DIR", "?")
    svc = cfg.get("BACKEND_SERVICE", "?")

    print("=" * 58)
    print("  部署配置摘要")
    print("=" * 58)
    print(f"  服务器:  {user}@{host}  [密码: {pw}]")
    print(f"  前端目录: {fe}")
    print(f"  后端目录: {be}")
    print(f"  服务名:   {svc}")
    print("=" * 58)


def main():
    # 解析参数
    target = "all"
    if len(sys.argv) > 1:
        arg = sys.argv[1].lower()
        if arg in ("frontend", "backend"):
            target = arg
        elif arg in ("-h", "--help"):
            print(__doc__)
            return
        else:
            print(f"❌ 未知参数: {arg}")
            print("   用法: python deploy_quick.py [frontend|backend]")
            sys.exit(1)

    # 加载配置
    cfg = load_config()
    show_config_summary(cfg)

    # 连接服务器
    ssh = sftp = None
    try:
        ssh, sftp = create_ssh_sftp(cfg)

        if target in ("all", "frontend"):
            print()
            log("📦===== 部署前端 =====")
            build_and_deploy_frontend(cfg, sftp, ssh)

        if target in ("all", "backend"):
            print()
            log("🖥️ ===== 部署后端 =====")
            build_and_deploy_backend(cfg, sftp, ssh)

        print()
        log("🎉 全部部署完成！")

    finally:
        if sftp:
            sftp.close()
        if ssh:
            ssh.close()


if __name__ == "__main__":
    main()
