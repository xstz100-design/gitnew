import paramiko

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('43.134.85.80', username='ubuntu', password='qer1235A@', timeout=30)

def run(cmd):
    stdin, stdout, stderr = ssh.exec_command(cmd)
    out = stdout.read().decode()
    err = stderr.read().decode()
    if err.strip():
        print(f'[err] {err[:300]}')
    return out

# Test logo.png response locally
print('=== curl test ===')
print(run("curl -s -o /dev/null -w 'HTTP %{http_code} Content-Type: %{content_type}' https://tfyx.shop/logo.png"))
print()

# Check nginx config
print('=== nginx root and server_name ===')
print(run("nginx -T 2>/dev/null | grep -E 'server_name|root' | head -15"))
print()

# Check what sites are enabled
print('=== sites-enabled ===')
print(run("ls -la /etc/nginx/sites-enabled/ && echo --- && cat /etc/nginx/sites-enabled/* 2>/dev/null || echo no sites-enabled"))
print()

# Check main nginx conf include
print('=== main conf ===')
print(run("grep include /etc/nginx/nginx.conf 2>/dev/null | head -10"))

ssh.close()
