import paramiko
import socket

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('43.134.85.80', username='ubuntu', password='qer1235A@', timeout=30)

def run(cmd):
    stdin, stdout, stderr = ssh.exec_command(cmd)
    out = stdout.read().decode()
    err = stderr.read().decode()
    if err.strip():
        print(f'[err] {err[:200]}')
    return out

# Check DNS resolution
print('=== DNS resolution from server ===')
print(run("getent hosts tfyx.shop khmerai.cn 2>/dev/null || (dig +short tfyx.shop; dig +short khmerai.cn)"))

# Direct curl test with verbose
print('=== verbose curl for logo.png ===')
print(run("curl -v -o /dev/null https://tfyx.shop/logo.png 2>&1 | head -30"))

print('=== fetch logo.png directly ===')
print(run("curl -s -D - https://tfyx.shop/logo.png -o /dev/null 2>&1 | head -20"))

# Check nginx mime types
print('=== mime types check ===')
print(run("grep -i png /etc/nginx/mime.types | head -5"))

# Check the actual wholesale config file
print('=== full wholesale config ===')
print(run("cat /etc/nginx/sites-available/wholesale"))

# Check if there's another conf for tfyx.shop
print('=== search for tfyx in nginx ===')
print(run("grep -r 'tfyx' /etc/nginx/ 2>/dev/null || echo 'no tfyx references found'"))

# Test with localhost direct
print('=== test direct to nginx local ===')
print(run("curl -s -o /dev/null -w '%{http_code} %{content_type}' --resolve khmerai.cn:443:127.0.0.1 https://khmerai.cn/logo.png"))

ssh.close()
