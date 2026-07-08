# EC2 agent

Synchronizes login credentials from the Instance Metadata Service (IMDS).

## Installation

### Debian

```bash
wget -P /tmp https://github.com/C2Devel/ec2-agent/releases/latest/download/ec2-agent_linux_amd64.deb
sudo apt install -y /tmp/ec2-agent_linux_amd64.deb
rm /tmp/ec2-agent_linux_amd64.deb
```

### RPM

```bash
wget -P /tmp https://github.com/C2Devel/ec2-agent/releases/latest/download/ec2-agent_linux_amd64.rpm
sudo yum install -y /tmp/ec2-agent_linux_amd64.rpm
rm /tmp/ec2-agent_linux_amd64.rpm
```
